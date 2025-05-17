# stage builder
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY ./ ./

RUN go mod download

RUN go build -ldflags="-extldflags '-static'" -o build/migration main.go

# stage production
FROM scratch

ENV GIN_MODE=release

WORKDIR /

COPY --from=builder /app/build/migration ./

CMD ["./migration", "migrate"]
