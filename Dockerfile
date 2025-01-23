FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache build-base cmake libde265-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Enable CGO and set flags
ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-Wno-deprecated-declarations"

COPY . .
RUN go build -o backend ./cmd/backend.go

# Final stage
FROM alpine:3.20.3

# Add runtime dependencies and fix paths
RUN apk --no-cache add ca-certificates libstdc++ && \
    mkdir /lib64 && \
    ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

WORKDIR /app
COPY --from=builder /app/backend .
RUN chmod +x /app/backend

CMD ["/app/backend"]