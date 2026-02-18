FROM golang:1.26-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o worker

FROM alpine:latest
RUN apt-get update && \
    apt-get install -y libvips libvips-dev && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /app/worker .
CMD ["./worker"]
