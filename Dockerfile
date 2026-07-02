# Etapa de build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/src/cmd/
RUN go build -mod=vendor -o /app/main .

# Etapa final
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/main .
ENTRYPOINT ["./main"]