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
