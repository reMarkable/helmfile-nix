with import <nixpkgs> { };
with builtins;

rec {
  yaml2nix =
    path:
    let
      jsonOutputDrv = runCommand "from-yaml" {
        nativeBuildInputs = [ yq-go ];
      } ''yq -M -o json . "${path}" > "$out"'';
    in
    fromJSON (readFile jsonOutputDrv);

  mlVals = val: ''
    {{ "${val}"|fetchSecretValue }}
  '';
  vals = val: ''{{"${val}"|fetchSecretValue}}'';

  # render helmfile to object
  render =
    state: env:
    let
      hf = import /${state}/helmfile.nix;
      var = {
        values = lib.mergeAttrs (yaml2nix /${state}/env/defaults.yaml) (yaml2nix /${state}/env/${env}.yaml);
        environment.name = env;
      };
    in
    hf {
      inherit
        lib
        var
        vals
        mlVals
        ;
    };
}
