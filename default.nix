{
  pkgs ? import <nixpkgs> {},
}: let
  inherit (pkgs) lib git buildGoModule runCommand;
  version =
    builtins.readFile (runCommand "get-git-tag" {nativeBuildInputs = [git];}
      "git --git-dir=${./.git} describe --always --tags | tr -d '\n' > $out");
  revision =
    builtins.readFile (runCommand "get-git-revision" {nativeBuildInputs = [git];}
      "git --git-dir=${./.git} rev-parse HEAD | tr -d '\n' > $out");
in
  buildGoModule (finalAttrs: {
    inherit version;
    pname = "clamav-rest";

    src = lib.fileset.toSource {
      root = ./.;
      fileset = lib.fileset.gitTracked ./.;
    };

    vendorHash = "sha256-akM/oHaSsuqucPQEY2434NDdPLcNVdcWZL4Zs6H0Ky8=";

    proxyVendor = true;
    ldflags = [
      "-s"
      "-w"
      "-X main.Version=${finalAttrs.version}"
      "-X main.Commit=${revision}"
    ];

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
