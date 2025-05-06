{ var, ... }:
[
  {
    environments = {
      dev = {
        values = [
          {
            bar = "baz";
          }
        ];
      };
    };
  }
  {
    releases = [
      {
        name = "test-nix";
        nixChart = "../nixChart/";
        values = {
          namespace = "test";
          cluster_name = "test-cluster";
          karpenter_instance_profile_role = var.values.bar;
        };
      }
    ];
  }
]
