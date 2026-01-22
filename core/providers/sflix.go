package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/demonkingswarn/luffy/core"
)

const (
	SFLIX_BASE_URL   = "https://sflix.is"
	SFLIX_SEARCH_URL = SFLIX_BASE_URL + "/search"
	SFLIX_AJAX_URL   = SFLIX_BASE_URL + "/ajax"
)

type Sflix struct {
	Client *http.Client
}

func NewSflix(client *http.Client) *Sflix {
	return &Sflix{Client: client}
}

func (s *Sflix) newRequest(method, url string) (*http.Request, error) {
	req, err := core.NewRequest(method, url)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", SFLIX_BASE_URL+"/")
	return req, nil
}

func (s *Sflix) Search(query string) ([]core.SearchResult, error) {
	search := strings.ReplaceAll(query, " ", "-")
	req, _ := s.newRequest("GET", SFLIX_SEARCH_URL+"/"+search)
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []core.SearchResult

	doc.Find("div.flw-item").EachWithBreak(func(i int, sel *goquery.Selection) bool {
		if i >= 10 {
			return false
		}

		title := sel.Find("h2.film-name a").AttrOr("title", "Unknown")
		href := sel.Find("div.film-poster a").AttrOr("href", "")
		poster := sel.Find("img.film-poster-img").AttrOr("data-src", "")
		typeStr := strings.TrimSpace(sel.Find("span.fdi-type").Text())

		mediaType := core.Movie
		if strings.EqualFold(typeStr, "TV") || strings.EqualFold(typeStr, "Series") {
			mediaType = core.Series
		}

		if href != "" {
			results = append(results, core.SearchResult{
				Title:  title,
				URL:    SFLIX_BASE_URL + href,
				Type:   mediaType,
				Poster: poster,
			})
		}
		return true
	})

	if len(results) == 0 {
		return nil, errors.New("no results")
	}

	return results, nil
}

func (s *Sflix) GetMediaID(url string) (string, error) {
	req, _ := s.newRequest("GET", url)
	resp, err := s.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	id := doc.Find("#watch-block").AttrOr("data-id", "")
	if id == "" {
		id = doc.Find("div.detail_page-watch").AttrOr("data-id", "")
	}
	if id == "" {
		id = doc.Find("#movie_id").AttrOr("value", "")
	}

	if id == "" {
		return "", fmt.Errorf("could not find media ID")
	}
	return id, nil
}

func (s *Sflix) GetSeasons(mediaID string) ([]core.Season, error) {
	url := fmt.Sprintf("%s/season/list/%s", SFLIX_AJAX_URL, mediaID)
	req, _ := s.newRequest("GET", url)
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var seasons []core.Season
	doc.Find(".dropdown-item").Each(func(i int, sel *goquery.Selection) {
		id := sel.AttrOr("data-id", "")
		name := strings.TrimSpace(sel.Text())
		if id != "" {
			seasons = append(seasons, core.Season{ID: id, Name: name})
		}
	})
	return seasons, nil
}

func (s *Sflix) GetEpisodes(id string, isSeason bool) ([]core.Episode, error) {
	var url string
	if isSeason {
		url = fmt.Sprintf("%s/season/episodes/%s", SFLIX_AJAX_URL, id)
	} else {
		url = fmt.Sprintf("%s/episode/list/%s", SFLIX_AJAX_URL, id)
	}

	req, _ := s.newRequest("GET", url)
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var episodes []core.Episode
	
	if isSeason {
		doc.Find("a.eps-item").Each(func(i int, sel *goquery.Selection) {
			id := sel.AttrOr("data-id", "")
			name := strings.TrimSpace(sel.AttrOr("title", sel.Text()))
			if id != "" {
				episodes = append(episodes, core.Episode{ID: id, Name: name})
			}
		})
	} else {
		// Movies: List of servers (treated as episodes/servers)
		doc.Find(".link-item").Each(func(i int, sel *goquery.Selection) {
			id := sel.AttrOr("data-id", "")
			name := strings.TrimSpace(sel.Find("span").Text())
			if id != "" {
				episodes = append(episodes, core.Episode{ID: id, Name: name})
			}
		})
	}

	return episodes, nil
}

func (s *Sflix) GetServers(episodeID string) ([]core.Server, error) {
	url := fmt.Sprintf("%s/episode/servers/%s", SFLIX_AJAX_URL, episodeID)
	req, _ := s.newRequest("GET", url)
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var servers []core.Server
	doc.Find(".link-item").Each(func(i int, sel *goquery.Selection) {
		id := sel.AttrOr("data-id", "")
		name := strings.TrimSpace(sel.Find("span").Text())
		if id != "" {
			servers = append(servers, core.Server{ID: id, Name: name})
		}
	})
	return servers, nil
}

func (s *Sflix) GetLink(serverID string) (string, error) {
	url := fmt.Sprintf("%s/episode/sources/%s", SFLIX_AJAX_URL, serverID)
	req, _ := s.newRequest("GET", url)
	resp, err := s.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res struct {
		Link string `json:"link"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	//fmt.Printf("DEBUG: Embedded URL: %s\n", res.Link)
	return res.Link, nil
}
