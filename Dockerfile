# syntax=docker/dockerfile:1
FROM golang:1.19 AS build
WORKDIR /src
ARG TARGETOS
ARG TARGETARCH
RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -o /go/bin/app ./cmd/server
RUN mkdir -p /tmp/sds-asm/public

FROM gcr.io/distroless/base-debian11@sha256:4b22ca3c68018333c56f8dddcf1f8b55f32889f2dd12d28ab60856eba1130d04
WORKDIR /
COPY --from=build /go/bin/app /
USER 101
COPY --from=build --chown=101:101 /tmp/sds-asm/public /tmp/sds-asm/public
VOLUME /tmp/sds-asm/public
ENTRYPOINT ["/app"]
