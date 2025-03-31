FROM golang:1.21-alpine AS builder

# Gerekli paketleri kur
RUN apk add --no-cache git gcc musl-dev

# Çalışma dizinini ayarla
WORKDIR /app

# Go modüllerini kopyala ve indir
COPY go.mod go.sum ./
RUN go mod download

# Kaynak kodu kopyala
COPY . .

# Uygulamayı derle
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scanner-service ./cmd/main

# Runtime image
FROM alpine:3.18

# Nmap ve gerekli paketleri kur
RUN apk add --no-cache nmap nmap-scripts ca-certificates tzdata

# Çalışma dizinini ayarla
WORKDIR /app

# Konfigürasyon dizinini oluştur
RUN mkdir -p /app/configs

# Derlenmiş uygulamayı kopyala
COPY --from=builder /app/scanner-service .

# Konfigürasyon dosyasını kopyala
COPY --from=builder /app/configs/config.yaml ./configs/

# Uygulamayı çalıştır
ENTRYPOINT ["/app/scanner-service"]