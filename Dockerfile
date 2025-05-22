# stage builder
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY ./ ./

RUN go build -tags timetzdata -ldflags="-extldflags '-static'" -o build/main-service main.go

# stage production
FROM scratch

ENV GIN_MODE=release

WORKDIR /

COPY --from=builder /app/build/main-service ./

EXPOSE 3000

CMD ["./main-service", "serve"]
