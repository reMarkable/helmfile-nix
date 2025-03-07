{ var, lib, ... }:
[
  {
    bases = [
      "../bases/environments.yaml"
      "../bases/defaults.yaml"
    ];
  }
  {
    repositories = [
      {
        name = "external-secrets";
        url = "https://charts.external-secrets.io/";
      }
    ];
    releases = [
      {
        name = "external-secrets";
        chart = "external-secrets/external-secrets";
        namespace = "external-secrets";
        version = "0.14.3";
        values = [
          {
            project = var.values.project;
          }
        ];
      }
    ];
  }
]
