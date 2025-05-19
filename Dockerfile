# stage builder
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY ./ ./

RUN go mod download

RUN go build -ldflags="-extldflags '-static'" -o build/wt-service main.go

# stage production
FROM scratch

ENV GIN_MODE=release

WORKDIR /

COPY --from=builder /app/build/wt-service ./

EXPOSE 8002

CMD ["./wt-service"]