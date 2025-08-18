# Etapa de build
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/src/cmd/
RUN go mod download
RUN go build -o /app/main .

# Etapa final
FROM alpine:latest
ENV SUPABASE_DSN="postgresql://postgres.cezpglskssirudmkqqze:Rc9652!$@aws-0-sa-east-1.pooler.supabase.com:6543/postgres"
ENV SUPABASE_API="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImNlenBnbHNrc3NpcnVkbWtxcXplIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc0NTI4MjE4NCwiZXhwIjoyMDYwODU4MTg0fQ.QEPZ9yLs0FuorBM6e2tsvcL6S4HqY25b_LN6lx8Am60"
ENV SUPABASE_ID="cezpglskssirudmkqqze"
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
EXPOSE 8000
ENTRYPOINT ["./main"]