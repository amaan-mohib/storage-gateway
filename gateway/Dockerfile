FROM golang:1.26-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o worker

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/worker .
CMD ["./worker"]
