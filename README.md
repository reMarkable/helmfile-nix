# helmfile-nix

A small wrapper around helmfile to allow writing helmfile.yaml in Nix.

## Basic usage

```sh
helmfile-nix render
```

Looks for helmfile.nix in the current directory and renders it to helmfile.yaml.

```sh
helmfile-nix -f foo/helmfile.nix -e stage diff
```

Renders the helmfile in stage and passes it on to helmfile diff.

For convenience we default to 'dev' if env is not set.

## Structure of your helmfile.nix

```nix
{ ... }: [
  { environments = { dev = { values = [ ]; }; }; }
  {
    repositories = [{
      name = "grafana";
      url = "https://grafana.github.io/helm-charts";
    }];
    releases = [{
      name = "grafana";
      chart = "grafana/grafana";
    }];
  }
]
```

Note that we expect an array of yaml documents, typically the first document is the environment
configuration and any defaults. The follow optional attributes are supported:

- lib: nixpkgs stdlib
- vals: A function to render secrets in your helmfile. See fetchSecretValue for more info.
- var: This will contain your environment variables, as well as the environment name. Follows the same structure as
  helmfile (var.environment.name / var.values.foo).

## Caveats

- We expect a env structure with a env/ directory in the same directory as the helmfile.nix
  file containing default.yaml and a $env.yaml file for each environment.
