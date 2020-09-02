# syntax = docker/dockerfile:1-experimental
FROM golang:1.15.1 AS base
WORKDIR /src
# ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

COPY go.* .
RUN go mod download
RUN go mod verify
# Get Dependencies
RUN go get github.com/go-delve/delve/cmd/dlv
COPY . .

FROM base as build
RUN --mount=type=cache,target=/root/.cache/go-build \
go build -gcflags="all=-N -l" -o /out/server

FROM debian:buster AS bin-unix
EXPOSE 8080 40000
COPY --from=build /go/bin/dlv /
COPY --from=build /out/server /
CMD ["/out/server"]
