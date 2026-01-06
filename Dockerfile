# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS build-stage

WORKDIR /app

RUN apk add --no-cache build-base sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . . 

RUN CGO_ENABLED=1 GOOS=linux go build -o /docker-na-raslabot



FROM build-stage AS run-test-stage
RUN apk add --no-cache build-base sqlite-dev
RUN go test -v ./...



FROM alpine:latest AS build-release-stage

WORKDIR /app

RUN apk add --no-cache ca-certificates sqlite-libs
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
RUN mkdir -p /data \
    && chown -R nonroot:nonroot /data \
    && touch /data/storage.db \
    && chown nonroot:nonroot /data/storage.db

COPY --from=build-stage --chown=nonroot:nonroot /docker-na-raslabot ./docker-na-raslabot

USER nonroot

ENTRYPOINT ["./docker-na-raslabot"]