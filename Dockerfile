FROM golang:1.26-alpine AS build
WORKDIR /app

RUN apk add --no-cache \
    build-base \
    pkgconfig \
    vips-dev \
    ca-certificates

COPY ./gateway/go.mod ./gateway/go.sum ./
RUN go mod download

COPY . .
WORKDIR /app/gateway

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

RUN go build -ldflags="-s -w" -o app ./src
RUN go build -ldflags="-s -w" -o worker-app ./worker

FROM alpine:latest

RUN apk add --no-cache \
    ffmpeg \
    vips \
    ca-certificates

RUN adduser -D appuser
USER appuser

WORKDIR /app
COPY --from=build /app/gateway/app ./app
COPY --from=build /app/gateway/worker-app ./worker

EXPOSE 5000

CMD ["./app"]
