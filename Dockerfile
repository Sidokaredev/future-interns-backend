# stage builder
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY ./ ./

RUN go mod download

RUN go build -ldflags="-extldflags '-static'" -o build/rt-service main.go

# stage production
FROM scratch

ENV GIN_MODE=release

WORKDIR /

COPY --from=builder /app/build/rt-service ./

EXPOSE 8001

CMD ["./rt-service"]