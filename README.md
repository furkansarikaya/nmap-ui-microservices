# Modern Nmap UI Microservices

Modern Nmap UI, ağ taramalarını yönetmek ve görselleştirmek için feature-based microservice mimarisi ile geliştirilmiş bir uygulamadır.

## 📋 İçindekiler

- [Mimari Genel Bakış](#-mimari-genel-bakış)
- [Microservices](#-microservices)
- [Başlangıç](#-başlangıç)
- [Geliştirme](#-geliştirme)
- [Feature-Based Organizasyon](#-feature-based-organizasyon)
- [API Referansları](#-api-referansları)
- [Deployment](#-deployment)
- [Monitoring](#-monitoring)
- [Katkıda Bulunma](#-katkıda-bulunma)
- [Lisans](#-lisans)

## 🏗 Mimari Genel Bakış

Bu proje, feature-based microservice mimarisini kullanmaktadır. Her bir servis, kendi içinde feature'lara göre organize edilmiştir. Servisler arasındaki iletişim gRPC ve REST API'ler üzerinden sağlanmaktadır.
### Mimari Prensipler

- **Domain-Driven Design (DDD)**: Her feature kendi domain modellerine sahiptir
- **Hexagonal Architecture**: Adapters, ports ve domain katmanları ayrılmıştır
- **Single Responsibility**: Her servis ve feature tek bir sorumluluğa sahiptir
- **Loose Coupling**: Servisler arasında gevşek bağlantı vardır
- **High Cohesion**: İlgili fonksiyonellikler bir arada tutulmuştur

## 🧩 Microservices

### Scanner Service
Nmap taramalarını yönetir ve tarama sonuçlarını analiz eder.

**Features:**
- **Scan**: Tarama işlemleri
- **History**: Tarama geçmişi yönetimi
- **Report**: Tarama sonuç raporları

### API Gateway
Tüm servisler için tek giriş noktası sağlar, yönlendirme ve kimlik doğrulama yapar.

**Features:**
- **Routing**: İstekleri ilgili servislere yönlendirme
- **Auth**: API endpointleri için kimlik doğrulama
- **Rate-limiting**: İstek sınırlama

### Web UI Service
Kullanıcı arayüzünü sunar, tarama sonuçlarını görselleştirir.

**Features:**
- **Dashboard**: Ana sayfa ve özet bilgiler
- **Scan-management**: Tarama başlatma ve yönetme
- **Visualizations**: Tarama sonuçlarını görselleştirme

### Storage Service
Tarama sonuçlarını ve kullanıcı verilerini depolar.

**Features:**
- **Scan-results**: Tarama sonuçları saklama
- **User-data**: Kullanıcı tercihlerini saklama
- **Export**: Verileri dışa aktarma

### Auth Service
Kullanıcı kimlik doğrulama ve yetkilendirme işlemlerini yönetir.

**Features:**
- **Authentication**: Kullanıcı kimlik doğrulama
- **Authorization**: Yetki kontrolleri
- **User-management**: Kullanıcı yönetimi

## 🚀 Başlangıç

### Ön Koşullar

- Go 1.21 veya üzeri
- Docker ve Docker Compose
- Nmap 7.x veya üzeri
- kubectl (opsiyonel, Kubernetes deployment için)

### Nmap Kurulumu

```bash
# MacOS
brew install nmap

# Ubuntu/Debian
sudo apt-get install nmap

# CentOS/RHEL
sudo yum install nmap
```

### Projeyi Klonlama ve Kurma

```bash
# Repo'yu klonlayın
git clone https://github.com/furkansarikaya/nmap-ui-microservices.git
cd nmap-ui-microservices

# Tüm servisleri derleyin
./tools/scripts/build.sh

# Geliştirme ortamında çalıştırın
docker compose -f deploy/docker-compose.yml up
```

### Ortam Değişkenleri

Her servis için gereken ortam değişkenleri ilgili servisin `configs` dizininde bulunabilir. Örnek:

```bash
# Scanner Service
SCANNER_PORT=8080
NMAP_PATH=/usr/bin/nmap
LOG_LEVEL=info

# API Gateway
API_GATEWAY_PORT=8000
JWT_SECRET=your-jwt-secret
```

## 💻 Geliştirme

### Dizin Yapısı

```
nmap-ui-microservices/
├── scanner-service/         # Nmap taramalarını yönetir
├── api-gateway/             # API gateway hizmeti
├── web-ui-service/          # Web UI hizmeti
├── storage-service/         # Veri depolama hizmeti
├── auth-service/            # Kimlik doğrulama hizmeti
├── shared-lib/              # Tüm servisler tarafından kullanılan ortak kod
├── deploy/                  # Deployment yapılandırmaları
├── tools/                   # Geliştirme ve operasyon araçları
└── docs/                    # Proje dokümantasyonu
```

### Tek Servis Geliştirme

```bash
# Scanner Service'i çalıştırma
cd scanner-service
go run cmd/main/main.go

# Testleri çalıştırma
go test ./...

# Yeni bir feature ekleme
mkdir -p internal/features/new-feature/{domain,handlers,repository}
```

### Tüm Servisleri Geliştirme

```bash
# Docker Compose ile tüm servisleri geliştirme modunda çalıştırma
docker compose -f deploy/docker-compose.dev.yml up
```

## 🧱 Feature-Based Organizasyon

Her servis içinde, ilgili özellikler şu şekilde düzenlenmiştir:

```
internal/features/feature-name/
├── domain/                # Domain modelleri ve servisleri
│   ├── models.go
│   └── service.go
├── handlers/              # HTTP ve gRPC endpoint'leri
│   ├── http.go
│   └── grpc.go
└── repository/            # Veri erişim katmanı
    └── repository.go
```

Bu yaklaşım, ilgili kodların bir arada tutulmasını ve her feature'ın bağımsız olarak geliştirilmesini sağlar.

## 📜 API Referansları

Her servis için API dokümantasyonu, ilgili servisin `api` dizininde OpenAPI veya Protocol Buffers formatında bulunabilir.

### REST API'ler

- Scanner Service: http://localhost:8081/swagger/index.html
- API Gateway: http://localhost:8000/swagger/index.html
- Storage Service: http://localhost:8083/swagger/index.html
- Auth Service: http://localhost:8084/swagger/index.html

### gRPC API'ler

Protocol Buffers tanımları `api/grpc/v1` dizinlerinde bulunabilir.

## 🚢 Deployment

### Docker Compose ile Deployment

```bash
# Üretim ortamı için
docker compose -f deploy/docker-compose.yml up -d
```

### Kubernetes ile Deployment

```bash
# Kubernetes namespace oluşturma
kubectl apply -f deploy/kubernetes/namespace.yaml

# Tüm servisleri deploy etme
kubectl apply -f deploy/kubernetes/
```

### Terraform ile Cloud Deployment

```bash
cd deploy/terraform
terraform init
terraform apply
```

## 📄 Lisans

Bu proje MIT Lisansı altında lisanslanmıştır - detaylar için [LICENSE](LICENSE.md) dosyasına bakın.

---

## 📞 İletişim

Furkan Sarıkaya - [@furkansarikaya](https://github.com/furkansarikaya)

Proje Linki: [https://github.com/furkansarikaya/nmap-ui-microservices](https://github.com/furkansarikaya/nmap-ui-microservices)