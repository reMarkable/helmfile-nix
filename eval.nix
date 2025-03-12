with builtins;

rec {
  # pin nixpkgs.lib
  lib =
    (builtins.getFlake "github:nix-community/nixpkgs.lib/4b620020fd73bdd5104e32c702e65b60b6869426").lib;

  # Multiline secret values
  mlVals = val: ''
    {{ "${val}"|fetchSecretValue }}
  '';
  # expand secret values
  vals = val: ''{{"${val}"|fetchSecretValue}}'';

  # Escape go template variables
  escape_var = var: ''{{"${var}"}}'';

  # render helmfile to object
  render =
    state: env: val:
    let
      hf = import /${state}/helmfile.nix;
      var = {
        values = fromJSON (readFile val);
        environment.name = env;
      };
    in
    hf {
      inherit
        escape_var
        lib
        var
        vals
        mlVals
        ;
      val = var.values;
    };
}
