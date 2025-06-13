# stage builder
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY ./ ./

RUN go build -ldflags="-extldflags '-static'" -o build/wb-service main.go

# stage production
FROM alpine:latest

ENV GIN_MODE=release

WORKDIR /app

COPY --from=builder /app/build/wb-service ./

EXPOSE 8003

CMD ["/app/wb-service"]
