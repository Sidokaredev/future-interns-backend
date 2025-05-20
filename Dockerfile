# stage builder
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY ./ ./

RUN go build -ldflags="-extldflags '-static'" -o build/noc-service main.go

# stage production
FROM alpine:latest

ENV GIN_MODE=release

WORKDIR /

COPY --from=builder /app/build/noc-service ./

EXPOSE 8004

CMD ["./noc-service"]
