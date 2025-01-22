FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o backend ./cmd/backend.go

FROM alpine:3.20.3
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/backend .
RUN chmod +x /app/backend

CMD ["/app/backend"]
