package core

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var mpv_executable string = "mpv"

func checkAndroid() bool {
	cmd := exec.Command("uname", "-o")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "Android"
}

func Play(url, title, referer, userAgent string, subtitles []string) error {
	
	if runtime.GOOS == "windows" {
		mpv_executable = "mpv.exe"
	} else {
		mpv_executable = "mpv"
	}

	var cmd *exec.Cmd

	if checkAndroid() {
		fmt.Println("~ Android Detected ~")
		args := []string{
			"start",
			"--user", "0",
			"-a", "android.intent.action.VIEW",
			"-d", url,
			"-n", "org.videolan.vlc/org.videolan.vlc.gui.video.VideoPlayerActivity",
			"-e", "title", fmt.Sprintf("Playing %s", title),
		}

		if len(subtitles) > 0 {
			args = append(args, "--es", "subtitles_location", subtitles[0])
		}

		cmd = exec.Command("am", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		fmt.Printf("Starting VLC on Android for %s...\n", title)
		return cmd.Run()
	}

	switch runtime.GOOS {
	case "darwin":
		args := []string{
			"--no-stdin",
			"--keep-running",
			fmt.Sprintf("--mpv-referrer=%s", referer),
			fmt.Sprintf("--mpv-user-agent=%s", userAgent),
			url,
			fmt.Sprintf("--mpv-force-media-title=Playing %s", title),
		}
		for _, sub := range subtitles {
			args = append(args, fmt.Sprintf("--mpv-sub-files=%s", sub))
		}
		cmd = exec.Command("iina", args...)

	default:
		cfg := LoadConfig()
		if cfg.Player == "mpv" {
			args := []string{
				url,
				fmt.Sprintf("--referrer=%s", referer),
				fmt.Sprintf("--user-agent=%s", userAgent),
				fmt.Sprintf("--force-media-title=Playing %s", title),
			}
			for _, sub := range subtitles {
				if sub != "" {
					args = append(args, fmt.Sprintf("--sub-file=%s", sub))
				}
			}

			cmd = exec.Command(mpv_executable, args...)
		}
	}

	if cmd != nil {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if len(subtitles) > 0 {
		fmt.Printf("Subtitles found: %d\n", len(subtitles))
	}

	fmt.Printf("Starting player for %s...\n", title)
	return cmd.Run()
}