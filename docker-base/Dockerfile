FROM nixos/nix:2.32.4 AS nix

ADD nix-static.nix  nix-static.nix
RUN echo "filter-syscalls = false" >> /etc/nix/nix.conf
RUN nix-build  ./nix-static.nix
RUN chmod 755 result/bin/nix && nix-shell -p gcc --run 'strip result/bin/nix'


FROM alpine:3.23@sha256:51183f2cfa6320055da30872f211093f9ff1d3cf06f39a0bdb212314c5dc7375

COPY --from=nix ./result/bin/nix /bin

