{
  description = "helmfile-nix";

  inputs.devenv.url = "github:cachix/devenv";
  inputs.devenv.inputs.nixpkgs.follows = "nixpkgs";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.gomod2nix.url = "github:nix-community/gomod2nix";
  inputs.gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
  inputs.gomod2nix.inputs.flake-utils.follows = "flake-utils";

  outputs =
    {
      devenv,
      flake-utils,
      gomod2nix,
      nixpkgs,
      self,
    }@inputs:
    (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.callPackage ./. {
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
        };
        packages.devenv-up = self.devShells.${system}.default.config.procfileScript;
        packages.devenv-test = self.devShells.${system}.default.config.test;
        devShells.default = devenv.lib.mkShell {
          inherit inputs pkgs;
          modules = [
            (import ./devenv.nix {
              inherit inputs pkgs system;
            })
          ];
        };
      }
    ));
}
