# syntax=docker/dockerfile:1

## Build
FROM golang:1.20.3-bullseye AS build
ARG VERSION=dev
WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=$VERSION" -o /go/bin/inpxer

## Deploy
FROM gcr.io/distroless/base-debian11

COPY --from=build /go/bin/inpxer /bin/inpxer

EXPOSE 8080/tcp
CMD ["inpxer", "serve"]
