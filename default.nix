{
  stdenv,
  testers,
  fzf,
  mpv,
  yt-dlp,
  chafa,
  lib,
  makeWrapper,
  fetchurl,
}:
stdenv.mkDerivation (finalAttrs: {
  pname = "luffy";
  version = "1.0.6";

  src = fetchurl {
    url = "https://github.com/DemonKingSwarn/luffy/releases/download/v${finalAttrs.version}/luffy";
    hash = "sha256-pa6mlh3oigp5oehlyjuytgs2go462rjk74h4obioel6inpwde5jq";
  };

  nativeBuildInputs = [makeWrapper];

  dontBuild = true;

  installPhase = ''
    runHook preInstall;
    mkdir -p $out/bin
    cp $src $out/bin/luffy
    chmod +x $out/bin/luffy
    runHook postInstall
  '';

  postInstall = ''
    wrapProgram $out/bin/luffy \
      --prefix PATH : ${lib.makeBinPath [
        fzf
        mpv
        yt-dlp
        chafa
      ]}
  '';

  passthru.tests.version = testers.testVersion {
    package = finalAttrs.finalPackage;
  };

  meta = {
    description = "CLI to watch Movies/TV Shows from the terminal";
    homepage = "https://github.com/demonkingswarn/luffy";
    license = lib.licenses.gpl3Only;
    maintainers = with lib.maintainers; [];
    mainProgram = "luffy";
    platforms = lib.platforms.unix;
  };
})
