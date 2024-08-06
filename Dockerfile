# syntax=docker/dockerfile:1
FROM golang:1.22.5

WORKDIR /app

RUN go install github.com/air-verse/air@latest

RUN --mount=type=cache,target=/go/pkg/mod/,sharing=locked \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    go build -o /bin/api

CMD air