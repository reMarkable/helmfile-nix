{
  pkgs,
  inputs,
  system,
  ...
}:
let
  goEnv = inputs.gomod2nix.legacyPackages.${system}.mkGoEnv { pwd = ./.; };
in
{
  packages = with pkgs; [
    goEnv
    inputs.gomod2nix.legacyPackages.${system}.gomod2nix
    docker
    helmfile
    kubernetes-helm
  ];
  pre-commit.hooks = {
    gofmt.enable = true;
    govet.enable = true;
    golangci-lint.enable = true;
    gotest.enable = true;
    commitizen.enable = true;
  };
  enterTest = ''
    nix build .
  '';
}
