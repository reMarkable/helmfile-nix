let
  nixpkgsCommit = "96ec055edbe5ee227f28cdbc3f1ddf1df5965102";

  nixpkgsTarball = builtins.fetchTarball {
    url = "https://github.com/NixOS/nixpkgs/archive/${nixpkgsCommit}.tar.gz";
    sha256 = "7doLyJBzCllvqX4gszYtmZUToxKvMUrg45EUWaUYmBg=";

  };
  pkgs = import nixpkgsTarball { };
in
pkgs.pkgsStatic.nixVersions.nix_2_29
