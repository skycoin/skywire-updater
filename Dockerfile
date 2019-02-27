FROM golang:1.11.5-alpine as build

ARG GOOS=linux

COPY . /go/src/github.com/watercompany/skywire-updater
WORKDIR /go/src/github.com/watercompany/skywire-updater/cmd/skywire-updater

# Compile to system architecture. Snippet for retrieving architecture taked from
# https://github.com/containous/traefik-library-image/blob/7dec7b825ca16d0524626fbbca35284adfe3ef58/alpine/Dockerfile#L3
RUN set -ex; \
        apkArch="$(apk --print-arch)"; \
        case "$apkArch" in \
            armhf) arch='arm' ;; \
            aarch64) arch='arm64' ;; \
            x86_64) arch='amd64' ;; \
            *) echo >&2 "error: unsupported architecture: $apkArch"; exit 1 ;; \
    esac; \
    GOARCH=$arch GOOS=$GOOS CGO=ENABLED=0 go build -o skywire-updater

FROM alpine:3.8

# needed to perform https requests
RUN apk --no-cache add ca-certificates

RUN mkdir -p skywire-updater/scripts

COPY --from=build /go/src/github.com/watercompany/skywire-updater/cmd/skywire-updater/skywire-updater /usr/bin/skywire-updater
COPY --from=build /go/src/github.com/watercompany/skywire-updater/scripts /skywire-updater/scripts

WORKDIR skywire-updater

ENTRYPOINT ["skywire-updater"]