# helmfile-nix

![tests](https://github.com/reMarkable/helmfile-nix/actions/workflows/test.yml/badge.svg)
![last-commit](https://img.shields.io/github/last-commit/reMarkable/helmfile-nix)

A small wrapper around [helmfile](https://github.com/helmfile/helmfile/) to
allow writing your declarations in the [nix language](https://nix.dev/tutorials/nix-language).
This avoids any YAML or go templating, while still taking advantage of
[helmfile's features](https://helmfile.readthedocs.io/en/stable/).

## Basic usage

```sh
helmfile-nix render
```

Looks for `helmfile.nix` in the current directory and renders it to a
`helmfile.yaml` to stdout.

```sh
helmfile-nix -f foo/helmfile.nix -e stage diff
```

Renders the helmfile in stage and passes it on to helmfile diff.

For convenience we default to 'dev' if env is not set.

- You can also check out this [presentation](./docs/presentation.html) given to the [Oslo NixOS User Group](https://www.meetup.com/oslo-nixos-user-group/) for a quick overview.

## Structure of your `helmfile.nix`

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

Note that we expect an array of YAML documents, typically the first document is the
environment configuration and any defaults. The follow optional attributes can
be imported in your helmfile.nix:

| Attribute  | Description                                                                        |
| ---------- | ---------------------------------------------------------------------------------- |
| lib        | nixpkgs stdlib                                                                     |
| vals       | A function to render secrets in your helmfile. See fetchSecretValue for more info. |
| var        | This will contain your environment variables, as well as the environment name.     |
|            | Follows the same structure as helmfile (var.environment.name / var.values.foo).    |
| escape_var | A function to escape a string for use in a helmfile template.                      |

## Options

helmfile-nix support [all the helmfile options](https://helmfile.readthedocs.io/en/stable/#cli-reference), in addition to:

| Option            | Description                                                                      |
| ----------------- | -------------------------------------------------------------------------------- |
| --show-trace      | Show a stack trace on error. This is passed to nix for the rendering and is      |
|                   | meant to be used when you see an error during render. In most cases the          |
|                   | error will point you to the right place though.                                  |
| --state-value-set | helmfile-nix will use this to override values, but it is also                    |
|                   | passed on to helmfile. This is useful if you want to override a state value      |
|                   | at runtime. For example, if you want to override the image of a pod temporarily. |
| -e env            | The environment to use. Defaults to 'dev'.                                       |
| -f file           | The helmfile.nix to use. Defaults to looking in the current directory.           |

## Useful links

- [helmfile](https://github.com/helmfile/helmfile/) - A declarative helm wrapper.
- [noogle](https://noogle.dev) - A nix function search engine.
- [nix.dev](https://nix.dev) - The official nix documentation.

## Caveats

- We expect a env structure with a `env/`` directory in the same directory as the helmfile.nix
file containing a `default.yaml` and a $env.yaml file for each environment.
- Even if your helmfile gets further values, they can not be processed by nix.
