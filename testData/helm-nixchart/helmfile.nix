{ var, ... }:
[
  {
    environments = {
      dev = {
        values = [
          { }
        ];
      };
    };
  }
  {
    helmDefaults.kubeContext = "kind-chart-testing";
    releases = [
      {
        name = "test-nix";
        nixChart = "../nixChart/";
        createNamespace = true;
        namespace = "test";
        values = {
          replicas = 2;
          version = var.values.nginx_version;
        };
      }
    ];
  }
]
