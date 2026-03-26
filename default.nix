{
  pkgs ? import <nixpkgs> {},
  version ? "0.0.0+unknown",
}: let
  inherit (pkgs) lib buildGoModule;
  goMod = builtins.readFile ./go.mod;
  goVersionLine =
    let
      matches = lib.filter (line: lib.hasPrefix "go " line) (lib.splitString "\n" goMod);
    in
      if matches == []
      then throw "Could not find Go version in go.mod"
      else builtins.head matches;
  goVersion = lib.removePrefix "go " goVersionLine;
  goVersionParts = lib.splitString "." goVersion;
  doCheck = false;
  goAttr = "go_${builtins.elemAt goVersionParts 0}_${builtins.elemAt goVersionParts 1}";
  go =
    lib.attrByPath [goAttr] (throw "Go package attribute ${goAttr} is not available in nixpkgs") pkgs;
in
  (buildGoModule.override {inherit go;}) (finalAttrs: {
    inherit version;
    pname = "clamav-rest";

    src = lib.fileset.toSource {
      root = ./.;
      fileset = lib.fileset.gitTracked ./.;
    };

    proxyVendor = true;
    vendorHash = "sha256-BFmkBqzxxbkRTYUot8Hf2tmCQztejSL0DK1I16Dpgh4=";

    meta = with lib; {
      description = "ClamAV virus/malware scanner with REST API. ";
      longDescription = ''
        This is a two in one docker image which runs the open source virus scanner
        ClamAV (https://www.clamav.net/), performs automatic virus definition updates
        as a background process and provides a REST API interface to interact
        with the ClamAV process.
      '';
      homepage = "https://github.com/ajilach/clamav-rest";
      license = licenses.mit;
    };
  })
