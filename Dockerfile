# syntax=docker/dockerfile:1

## Build
FROM golang:1.19.0-bullseye AS build
WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /go/bin/inpxer

## Deploy
FROM gcr.io/distroless/base-debian11

COPY --from=build /go/bin/inpxer /bin/inpxer

EXPOSE 8080/tcp
CMD ["inpxer", "serve"]
