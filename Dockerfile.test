FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY ./ ./

RUN go build -ldflags="-extldflags '-static'" -o build/ca-service main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/build/ca-service ./

ENV GIN_MODE=release

EXPOSE 8000

CMD ["/app/ca-service"]
