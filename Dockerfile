FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/inventory-api ./cmd/api

FROM alpine:3.20

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/inventory-api /app/inventory-api

USER appuser:appgroup

EXPOSE 8080

ENTRYPOINT ["/app/inventory-api"]

