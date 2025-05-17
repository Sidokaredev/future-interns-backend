# stage builder
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY ./ ./

RUN go mod download

RUN go build -ldflags="-extldflags '-static'" -o build/ca-service main.go

# stage production
FROM scratch

ENV GIN_MODE=release

WORKDIR /

COPY --from=builder /app/build/ca-service ./

EXPOSE 8000

CMD ["./ca-service"]