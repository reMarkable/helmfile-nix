with builtins;

rec {
  # pin nixpkgs.lib
  lib =
    (builtins.getFlake "github:nix-community/nixpkgs.lib/4b620020fd73bdd5104e32c702e65b60b6869426").lib;

  # render chart to object
  render =
    file: state: val:
    let
      chart = import "/${state}/${file}";
      var = {
        values = fromJSON (readFile val);
      };
    in
    chart {
      inherit lib var;
      val = var.values;
    };
}
