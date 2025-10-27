FROM nixos/nix:2.32.2 AS nix

ADD nix-static.nix  nix-static.nix
RUN echo "filter-syscalls = false" >> /etc/nix/nix.conf
RUN nix-build  ./nix-static.nix
RUN chmod 755 result/bin/nix && nix-shell -p gcc --run 'strip result/bin/nix'


FROM alpine:3.22@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412

COPY --from=nix ./result/bin/nix /bin

