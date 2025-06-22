FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY ./ ./

RUN go build -ldflags="-extldflags '-static'" -o build/wt-service main.go

# stage production
FROM alpine:latest

ENV GIN_MODE=release

WORKDIR /app

COPY --from=builder /app/build/wt-service ./

EXPOSE 8002

CMD ["/app/wt-service"]
