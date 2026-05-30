FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o cluster-service ./cmd/cluster-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/cluster-service .
COPY --from=builder /app/config ./config

EXPOSE 8088

CMD ["./cluster-service"]
