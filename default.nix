{
  pkgs ? import <nixpkgs> {},
  version ? "0.0.0+unknown",
}: let
  inherit (pkgs) lib buildGoModule;
in
  buildGoModule (finalAttrs: {
    inherit version;
    pname = "clamav-rest";

    src = lib.fileset.toSource {
      root = ./.;
      fileset = lib.fileset.gitTracked ./.;
    };

    proxyVendor = true;
    vendorHash = "sha256-wwPGs9/oI+8DopN+MIWVNwX05D4qQ2pMdWfewil4H8M=";

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
