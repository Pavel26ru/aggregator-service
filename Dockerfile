FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ../../GoTour/aggregator-service .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o aggregator ./cmd/aggregator

FROM alpine:3.18

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/aggregator /app/aggregator

EXPOSE 8080
EXPOSE 9090

ENTRYPOINT ["/app/aggregator"]