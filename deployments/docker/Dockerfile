FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.sum go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/loadbalancer ./cmd/loadbalancer




FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/loadbalancer ./loadbalancer

COPY internal/config/config.yaml ./config.yaml

EXPOSE 3030

CMD ["./loadbalancer", "-config", "./config.yaml"]