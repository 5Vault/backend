# Etapa de build
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/src/cmd/
RUN go mod download
RUN go build -o /app/main .

# Etapa final
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
ENTRYPOINT ["./main"]