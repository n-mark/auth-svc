FROM golang:1.25 AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY go.mod go.sum ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o user-service ./cmd/server
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/user-service .
EXPOSE 8000
CMD ["./user-service"]

