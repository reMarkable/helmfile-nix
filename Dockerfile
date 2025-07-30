FROM  golang:1.24.3-bullseye AS builder

ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app/
ADD . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o helmfile-nix .


# Build the binary.
RUN go build -mod=readonly -v -o helmfile-nix .

FROM nixos/nix:2.29.0 AS nix

ADD nix-static.nix  nix-static.nix
RUN echo "filter-syscalls = false" >> /etc/nix/nix.conf
RUN nix-build  ./nix-static.nix
RUN chmod 755 result/bin/nix && nix-shell -p gcc --run 'strip result/bin/nix'


FROM alpine:3.19@sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b
ARG TARGETOS
ARG TARGETARCH


# renovate: datasource=github-releases depName=helmfile/helmfile
ARG HELMFILE_VERSION=v1.1.0
# renovate: datasource=github-releases depName=helm/helm
ARG HELM_VERSION=v3.18.3
# renovate: datasource=github-releases depName=databus23/helm-diff
ARG HELM_DIFF_VERSION=v3.12.0
# renovate: datasource=github-releases depName=kubernetes-sigs/kustomize
ARG KUSTOMIZE_VERSION=5.6.0

COPY --from=nix ./result/bin/nix /bin

COPY --from=builder /app/helmfile-nix /usr/local/bin/helmfile-nix

ENV INSTALL_PATH=/usr/local/bin
ENV HELM_PLUGINS=/usr/local/lib/helm-plugins

RUN echo "Building for ${TARGETOS}-${TARGETARCH}"

RUN apk add --update --no-cache bash curl git yq && \
  chmod +x ${INSTALL_PATH}/helmfile-nix && \
  # helmfile
  HELMFILE_STRIPPED=$(echo "$HELMFILE_VERSION" | cut -c2-) && \
  export HELMFILE_STRIPPED && \
  curl -sSL "https://github.com/helmfile/helmfile/releases/download/${HELMFILE_VERSION}/helmfile_${HELMFILE_STRIPPED}_linux_${TARGETARCH}.tar.gz" | tar -zx -C ${INSTALL_PATH} -f - helmfile && \
  chmod +x ${INSTALL_PATH}/helmfile && \
  # Install kustomize
  curl -sSL "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_${TARGETARCH}.tar.gz" | tar -zx -C ${INSTALL_PATH} -f - && \
  # Install helm and friends
  curl -sSL "https://get.helm.sh/helm-${HELM_VERSION}-linux-${TARGETARCH}.tar.gz" | tar -zx --strip-components=1 -C ${INSTALL_PATH} -f - linux-${TARGETARCH}/helm && \
  helm plugin install https://github.com/databus23/helm-diff --version ${HELM_DIFF_VERSION}

VOLUME /nix
RUN mkdir /etc/nix && \
  addgroup -g 1000 nixbld && \
  adduser -u 1000 -G nixbld -D nixbld && \
  echo experimental-features = nix-command flakes > /etc/nix/nix.conf && \
  echo substituters = https://cache.nixos.org >> /etc/nix/nix.conf

ENTRYPOINT ["/usr/local/bin/helmfile-nix" ]
