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
}
