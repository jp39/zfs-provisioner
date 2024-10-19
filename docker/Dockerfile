ARG GO_VERSION
ARG TARGETOS
ARG TARGETARCH
FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /go/src/zfs-provisioner

RUN --mount=readonly,target=. --mount=type=cache,target=/go/pkg/mod \
    go mod download && GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o /go/bin/zfs-provisioner .

FROM docker.io/library/alpine:3.20 AS image

RUN apk add zfs

COPY --from=build /go/bin/zfs-provisioner /usr/bin/

ENTRYPOINT ["/usr/bin/zfs-provisioner"]
