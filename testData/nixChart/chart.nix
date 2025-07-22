{ val, ... }:
let
  containerPort = val.containerPort or 80;
  replicas = val.replicas or 3;
  version = val.version or "1.14.2";
in
[
  {
    apiVersion = "apps/v1";
    kind = "Deployment";
    metadata = {
      name = "nginx-deployment";
      namespace = val.namespace;
      labels = {
        app = "nginx";
      };
    };
    spec = {
      inherit replicas;
      selector = {
        matchLabels = {
          app = "nginx";
        };
      };
      template = {
        metadata = {
          labels = {
            app = "nginx";
          };
        };
        spec = {
          containers = [
            {
              name = "nginx";
              image = "nginx:${version}";
              ports = [
                {
                  inherit containerPort;
                }
              ];
            }
          ];
        };
      };
    };
  }
]
