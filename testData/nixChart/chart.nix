{ val, ... }:
[
  {
    name = "test-chart";
    templates = [
      {
        apiVersion = "karpenter.k8s.aws/v1";
        kind = "EC2NodeClass";
        metadata = {
          name = "default";
          namespace = val.namespace;
        };
        spec = {
          amiSelectorTerms = [ { alias = "al2023@latest"; } ];
          userData = ''
                MIME-Version: 1.0
                Content-Type: multipart/mixed; boundary="//"

                --//
                Content-Type: application/node.eks.aws

            # Example custom nodeconfig which mounts individual drives on an instance
                apiVersion: node.eks.aws/v1alpha1
                kind: NodeConfig
                spec:
                  containerd:
                    config: |
                      [plugins."io.containerd.grpc.v1.cri".containerd]
                      discard_unpacked_layers = false
                  instance:
                    localStorage:
                      strategy: Raid0
                --//
          '';
          kubelet.maxPods = 110;
          metadataOptions.httpPutResponseHopLimit = 2;
          role = val.karpenter_instance_profile_role;
          securityGroupSelectorTerms = [ { "tags.karpenter.sh/discovery" = val.cluster_name; } ];
          subnetSelectorTerms = [ { tags."karpenter.sh/discovery" = val.cluster_name; } ];
          tags."karpenter.sh/discovery" = val.cluster_name;
        };
      }
    ];
  }
]
