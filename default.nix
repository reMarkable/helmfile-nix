{
  version,
  pkgs ? (
    let
      inherit (builtins) fetchTree fromJSON readFile;
      inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
    in
    import (fetchTree nixpkgs.locked) {
      overlays = [ (import "${fetchTree gomod2nix.locked}/overlay.nix") ];
    }
  ),
  buildGoApplication ? pkgs.buildGoApplication,
}:
let
  helmWrap =
    with pkgs;
    wrapHelm kubernetes-helm {
      plugins = with kubernetes-helmPlugins; [
        helm-diff
        helm-git
      ];
    };
  helmfileWrap = pkgs.helmfile-wrapped.override { inherit (helmWrap) pluginsDir; };
in
buildGoApplication {
  pname = "helmfile-nix";
  inherit version;
  ldflags = "-X main.version=${version} -w -s";
  pwd = ./.;
  src = ./.;
  nativeBuildInputs = with pkgs; [
    nix
    helmWrap
    helmfileWrap
  ];
  modules = ./gomod2nix.toml;
  preCheck = ''
    export HOME=$TMPDIR
    go test -race -v ./...
  '';
  postCheck = '''';
}
