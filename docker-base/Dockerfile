FROM nixos/nix:2.31.0 AS nix

ADD nix-static.nix  nix-static.nix
RUN echo "filter-syscalls = false" >> /etc/nix/nix.conf
RUN nix-build  ./nix-static.nix
RUN chmod 755 result/bin/nix && nix-shell -p gcc --run 'strip result/bin/nix'


FROM alpine:3.22@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1

COPY --from=nix ./result/bin/nix /bin

