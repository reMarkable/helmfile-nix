FROM nixos/nix:2.35.4 AS nix

ADD nix-static.nix  nix-static.nix
RUN echo "filter-syscalls = false" >> /etc/nix/nix.conf
RUN nix-build  ./nix-static.nix
RUN chmod 755 result/bin/nix && nix-shell -p gcc --run 'strip result/bin/nix'


FROM alpine:3.24@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a

COPY --from=nix ./result/bin/nix /bin

