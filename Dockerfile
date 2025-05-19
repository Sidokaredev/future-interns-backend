# stage builder
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY ./ ./

RUN go mod download

RUN go build -ldflags="-extldflags '-static'" -o build/wb-service main.go

# stage production
FROM scratch

ENV GIN_MODE=release

WORKDIR /

COPY --from=builder /app/build/wb-service ./

EXPOSE 8003

CMD ["./wb-service"]