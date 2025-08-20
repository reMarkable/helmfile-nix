FROM nixos/nix:2.29.0 AS nix

ADD nix-static.nix  nix-static.nix
RUN echo "filter-syscalls = false" >> /etc/nix/nix.conf
RUN nix-build  ./nix-static.nix
RUN chmod 755 result/bin/nix && nix-shell -p gcc --run 'strip result/bin/nix'


FROM alpine:3.19@sha256:3be987e6cde1d07e873c012bf6cfe941e6e85d16ca5fc5b8bedc675451d2de67

COPY --from=nix ./result/bin/nix /bin

