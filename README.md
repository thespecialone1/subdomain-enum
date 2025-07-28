# 🚀 Advanced Subdomain Enumeration Tool v2.2

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)](LICENSE)
[![Release](https://img.shields.io/github/v/release/thespecialone1/subdomain-enum?style=for-the-badge)](https://github.com/thespecialone1/subdomain-enum/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/thespecialone1/subdomain-enum?style=for-the-badge)](https://ghcr.io/thespecialone1/subdomain-enum)

A **professional-grade** subdomain enumeration platform that discovers subdomains using multiple sources and techniques. Features real-time streaming results, comprehensive monitoring, and an intuitive web interface.

![Subdomain Enum Demo](https://via.placeholder.com/800x400/0a0a0a/4ade80?text=Advanced+Subdomain+Enumeration+Tool+v2.2)

## ✨ What's New in v2.2

### 🐳 **Production-Ready Docker Support**
- Multi-stage Docker builds for optimized images
- Multi-architecture support (AMD64, ARM64)
- Development and production configurations
- Docker Compose with monitoring stack

### 📊 **Enhanced Monitoring & Metrics**
- Prometheus metrics integration
- Grafana dashboards for visualization
- Health checks and readiness probes
- Real-time performance monitoring

### 🖥️ **Improved Web Interface**
- Professional SVG icons throughout
- Enhanced metrics dashboard
- Real-time statistics and system health
- Better mobile responsiveness

### 🔧 **Advanced Configuration**
- Comprehensive environment variable support
- Command-line interface with flags
- Configurable timeouts and concurrency
- Enhanced security settings

### 🛡️ **Security & Reliability**
- Rate limiting and request validation
- Enhanced error handling
- Graceful stream completion
- Security scanning with Trivy

## 🎯 Key Features

### **Multi-Source Discovery**
- 🏛️ **Wayback Machine**: Historical subdomain discovery from web archives
- 🔒 **Certificate Transparency**: SSL/TLS certificate logs analysis (crt.sh)
- 🌐 **DNS Brute Force**: Dictionary-based resolution with 500+ patterns
- 🔍 **Search Engine Scraping**: Automated Google search result parsing
- 🔄 **Permutation Generation**: Intelligent subdomain variations
- 📡 **Zone Transfer**: DNS misconfiguration testing

### **Professional Interface**
- ⚡ **Real-time Streaming**: Results appear instantly as discovered
- 📊 **Live Statistics**: Comprehensive metrics and monitoring
- 🎛️ **Source Selection**: Enable/disable individual discovery methods
- 🔍 **Advanced Probing**: HTTP/HTTPS status checking with titles
- 📋 **Export Options**: Multiple formats for integration
- ⏱️ **Smart Timeouts**: Automatic completion detection

### **Enterprise Features**
- 🏥 **Health Monitoring**: System health and readiness endpoints
- 📈 **Prometheus Metrics**: Production-ready monitoring
- 🐳 **Container Ready**: Optimized Docker images and Kubernetes support
- 🔐 **Security First**: Rate limiting, input validation, and secure defaults
- 📊 **Comprehensive Logging**: Structured logging with multiple levels

## 🚀 Quick Start

### Option 1: Docker (Recommended)

```bash
# Run with Docker
docker run -p 8080:8080 -p 9090:9090 ghcr.io/thespecialone1/subdomain-enum:latest

# Or with Docker Compose (includes monitoring)
curl -O https://raw.githubusercontent.com/thespecialone1/subdomain-enum/main/docker-compose.yml
docker-compose up -d
```

### Option 2: Pre-built Binaries

```bash
# Linux
wget https://github.com/thespecialone1/subdomain-enum/releases/latest/download/subdomain-enum-linux-amd64.tar.gz
tar -xzf subdomain-enum-linux-amd64.tar.gz
./subdomain-enum-linux-amd64

# macOS
wget https://github.com/thespecialone1/subdomain-enum/releases/latest/download/subdomain-enum-darwin-amd64.tar.gz
tar -xzf subdomain-enum-darwin-amd64.tar.gz
./subdomain-enum-darwin-amd64

# Windows
Invoke-WebRequest -Uri "https://github.com/thespecialone1/subdomain-enum/releases/latest/download/subdomain-enum-windows-amd64.exe.zip" -OutFile "subdomain-enum.zip"
Expand-Archive subdomain-enum.zip
.\subdomain-enum\subdomain-enum-windows-amd64.exe
```

### Option 3: Build from Source

```bash
# Prerequisites: Go 1.24+, Git
git clone https://github.com/thespecialone1/subdomain-enum.git
cd subdomain-enum

# Using Make (recommended)
make build
./dist/subdomain-enum

# Or manually
go mod tidy
go build -o subdomain-enum cmd/server/main.go
./subdomain-enum
```

## 🌐 Usage

### Web Interface
1. **Start the application** using any method above
2. **Open your browser** to `http://localhost:8080`
3. **Enter target domain** (e.g., `example.com`)
4. **Select discovery sources** you want to use
5. **Click "Start Scan"** and watch results stream in real-time
6. **Review results** and export in your preferred format

### API Access

```bash
# Start a scan via API
curl -N "http://localhost:8080/api/wayback/stream?target=example.com"

# Get system statistics
curl "http://localhost:8080/api/stats" | jq .

# Health check
curl "http://localhost:8080/health"

# Get Prometheus metrics
curl "http://localhost:8080/metrics"
```

### Command Line Options

```bash
# Show version information
./subdomain-enum --version

# Show help
./subdomain-enum --help

# Use custom port
./subdomain-enum --port 9080

# Health check (for containers)
./subdomain-enum --health-check
```

## 📊 Discovery Methods Explained

| Method | Description | Timeout | Best For |
|--------|-------------|---------|----------|
| **Wayback Machine** | Historical web crawl data | 5 min | Finding old/deprecated subdomains |
| **Certificate Transparency** | SSL/TLS certificate logs | 5 min | Active HTTPS subdomains |
| **DNS Brute Force** | Dictionary-based resolution | 10 min | Comprehensive discovery |
| **Search Engine** | Google search scraping | 5 min | Publicly indexed subdomains |
| **Permutation** | Intelligent pattern generation | 10 min | Development/staging patterns |
| **Zone Transfer** | DNS misconfiguration testing | 2 min | Misconfigured nameservers |

## ⚙️ Configuration

### Environment Variables

```bash
# Core Settings
export PORT=8080                    # Main server port
export METRICS_PORT=9090            # Metrics server port
export LOG_LEVEL=INFO               # Logging level

# DNS Configuration
export DNS_SERVERS=8.8.8.8:53,1.1.1.1:53
export DNS_CONCURRENCY=50           # Concurrent DNS queries
export DNS_TIMEOUT=3s               # DNS query timeout

# Rate Limiting
export RATE_LIMIT_RPS=10            # Requests per second
export RATE_LIMIT_BURST=20          # Burst capacity

# Timeouts (in minutes)
export TIMEOUT_WAYBACK=5m
export TIMEOUT_CRTSH=5m
export TIMEOUT_DNS=10m
export TIMEOUT_SEARCH=5m
export TIMEOUT_PERMUTE=10m
export TIMEOUT_ZONE=2m

# Security
export HTTP_SKIP_TLS_VERIFY=true    # Skip TLS verification
export MAX_CONCURRENT_JOBS=10       # Maximum simultaneous scans
```

### Docker Configuration

```yaml
# docker-compose.yml
version: '3.8'
services:
  subdomain-enum:
    image: ghcr.io/thespecialone1/subdomain-enum:latest
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - DNS_CONCURRENCY=100
      - RATE_LIMIT_RPS=20
      - LOG_LEVEL=DEBUG
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 1G
```

## 📈 Monitoring & Observability

### Prometheus Metrics

The tool exposes comprehensive metrics at `/metrics`:

```
# System metrics
subdomain_scanner_requests_total
subdomain_scanner_active_jobs
subdomain_scanner_subdomains_total
subdomain_scanner_dns_queries_total
subdomain_scanner_uptime_seconds

# Performance metrics
http_request_duration_seconds
http_requests_total
dns_query_duration_seconds
```

### Health Checks

```bash
# Basic health check
curl http://localhost:8080/health

# Kubernetes readiness probe
curl http://localhost:8080/ready

# Container health check
docker run --health-cmd="./subdomain-enum --health-check" \
  ghcr.io/thespecialone1/subdomain-enum:latest
```

### Grafana Dashboards

Pre-configured dashboards are available in the `monitoring/` directory:

- **System Overview**: Resource usage, request rates, response times
- **DNS Performance**: Query rates, resolution times, error rates
- **Discovery Analytics**: Sources performance, success rates

## 🐳 Docker Deployment

### Single Container
```bash
# Production deployment
docker run -d \
  --name subdomain-enum \
  -p 8080:8080 \
  -p 9090:9090 \
  --restart unless-stopped \
  ghcr.io/thespecialone1/subdomain-enum:latest
```

### Docker Compose (Full Stack)
```bash
# Download compose file
curl -O https://raw.githubusercontent.com/thespecialone1/subdomain-enum/main/docker-compose.yml

# Start all services (app + monitoring)
docker-compose --profile monitoring up -d

# Access services
# - App: http://localhost:8080
# - Prometheus: http://localhost:9092
# - Grafana: http://localhost:3000 (admin/admin123)
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: subdomain-enum
spec:
  replicas: 2
  selector:
    matchLabels:
      app: subdomain-enum
  template:
    metadata:
      labels:
        app: subdomain-enum
    spec:
      containers:
      - name: subdomain-enum
        image: ghcr.io/thespecialone1/subdomain-enum:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        env:
        - name: DNS_CONCURRENCY
          value: "100"
        - name: RATE_LIMIT_RPS
          value: "20"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
```

## 📊 Export & Integration

### Export Formats

| Format | Use Case | Command |
|--------|----------|---------|
| **Plain Text** | Simple lists | Copy all hosts |
| **CSV** | Spreadsheet analysis | Include status & titles |
| **JSON** | API integration | Structured data with metadata |
| **XML** | Legacy systems | Formatted for older tools |

### Security Tool Integration

| Tool | Format | Export Option |
|------|--------|---------------|
| **Nmap** | Host list | Plain text targets |
| **Masscan** | IP ranges | Formatted for high-speed scanning |
| **Burp Suite** | Scope definition | URL patterns for web testing |
| **Amass** | Configuration | INI format for OWASP Amass |

### API Integration Examples

```bash
# Python integration
import requests
response = requests.get('http://localhost:8080/api/stats')
stats = response.json()

# PowerShell integration
$stats = Invoke-RestMethod -Uri 'http://localhost:8080/api/stats'

# curl + jq processing
curl -s http://localhost:8080/api/stats | jq '.total_subdomains'
```

## 🔧 Development

### Development Setup
```bash
# Clone and setup
git clone https://github.com/thespecialone1/subdomain-enum.git
cd subdomain-enum
make dev-setup

# Run with hot reloading
make run-dev

# Run tests
make test

# Build for all platforms
make build-all
```

### Project Structure
```
subdomain-enum/
├── cmd/server/main.go          # Main application
├── public/index.html           # Web interface
├── monitoring/                 # Grafana dashboards & Prometheus config
├── .github/workflows/          # CI/CD pipelines
├── Dockerfile                  # Multi-stage Docker build
├── docker-compose.yml          # Full deployment stack
├── Makefile                    # Build automation
└── README.md                   # This documentation
```

### Contributing
1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Run quality checks: `make ci`
5. Commit changes: `git commit -m 'Add amazing feature'`
6. Push to branch: `git push origin feature/amazing-feature`  
7. Open a Pull Request

## 🚨 Security Considerations

### Ethical Usage
- ⚠️ **Only scan domains you own** or have explicit permission to test
- ⚠️ **Respect rate limits** and robots.txt files
- ⚠️ **Some techniques may trigger** security monitoring systems
- ⚠️ **Consider legal implications** in your jurisdiction

### Security Features
- 🔒 **Rate limiting** prevents abuse and reduces detection
- 🔒 **Input validation** sanitizes all user inputs
- 🔒 **TLS verification** can be enabled for production
- 🔒 **User agent rotation** reduces fingerprinting
- 🔒 **Request timeouts** prevent resource exhaustion

### Network Considerations
- **Outbound HTTPS** to various APIs (Wayback, crt.sh)
- **DNS queries** to configured resolvers
- **HTTP/HTTPS probes** to discovered subdomains
- **No data storage** - results kept in memory only

## 📊 Performance & Benchmarks

### Typical Performance
| Domain Size | Discovery Time | Memory Usage | CPU Usage |
|-------------|----------------|--------------|-----------|
| Small (<100) | 2-5 minutes | 50-100 MB | Low |
| Medium (100-1K) | 5-15 minutes | 100-200 MB | Moderate |
| Large (1K+) | 15-30+ minutes | 200-500 MB | High |

### Optimization Tips
- **Adjust DNS concurrency** based on network capacity
- **Use shorter timeouts** for faster scanning
- **Enable specific sources** only when needed
- **Monitor resource usage** during large scans
- **Scale horizontally** with multiple instances

## 🐛 Troubleshooting

### Common Issues

**Port Already in Use**
```bash
# Find process using port
lsof -i :8080
# Use different port
./subdomain-enum --port 9080
```

**DNS Resolution Failures**
```bash
# Test DNS connectivity
nslookup google.com 8.8.8.8
# Try different DNS servers
export DNS_SERVERS=1.1.1.1:53,208.67.222.222:53
```

**Memory Usage Issues**
```bash
# Monitor memory
docker stats subdomain-enum
# Reduce concurrency
export DNS_CONCURRENCY=25
```

**API Timeouts**
```bash
# Increase timeouts
export TIMEOUT_WAYBACK=10m
export TIMEOUT_CRTSH=10m
```

### Debug Mode
```bash
# Enable debug logging
export LOG_LEVEL=DEBUG
./subdomain-enum

# Or with Docker
docker run -e LOG_LEVEL=DEBUG ghcr.io/thespecialone1/subdomain-enum:latest
```

## 🆚 Comparison with Other Tools

| Feature | This Tool | Subfinder | Amass | Sublist3r |
|---------|-----------|-----------|-------|-----------|
| **Web Interface** | ✅ Modern UI | ❌ CLI only | ❌ CLI only | ❌ CLI only |
| **Real-time Results** | ✅ Streaming | ❌ Batch | ❌ Batch | ❌ Batch |
| **Docker Support** | ✅ Production-ready | ⚠️ Basic | ⚠️ Basic | ❌ None |
| **Monitoring** | ✅ Prometheus/Grafana | ❌ None | ❌ None | ❌ None |
| **Multiple Sources** | ✅ 6+ sources | ✅ Many | ✅ Many | ⚠️ Limited |
| **HTTP Probing** | ✅ Built-in | ❌ External | ✅ Built-in | ❌ External |
| **Export Formats** | ✅ 8+ formats | ⚠️ Limited | ⚠️ Limited | ⚠️ Limited |
| **API Access** | ✅ REST + SSE | ❌ None | ❌ None | ❌ None |

## 📋 Changelog

### v2.2.0 (Latest)
- 🐳 Production-ready Docker support with multi-stage builds
- 📊 Enhanced metrics and monitoring with Prometheus integration
- 🖥️ Improved web interface with professional SVG icons
- 🏥 Advanced health checks for container deployments
- ⚙️ Comprehensive configuration via environment variables
- 🔄 Auto-completion detection prevents infinite refresh loops
- 🛡️ Enhanced security with rate limiting and input validation
- 🔧 Command-line interface with version and help flags
- 📈 Multi-architecture Docker images (AMD64, ARM64)
- 🔍 Automated security scanning with Trivy

### v2.1.0
- Fixed infinite refresh loops in web interface
- Enhanced stream completion handling
- Improved error messages and logging
- Better mobile responsiveness

### v2.0.0
- Complete rewrite with Go backend
- Real-time streaming results
- Multiple discovery sources
- Professional web interface
- HTTP probing with title extraction

## 📞 Support & Community

### Getting Help
- 📖 **Documentation**: This README and inline help
- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/thespecialone1/subdomain-enum/issues)
- 💡 **Feature Requests**: [GitHub Discussions](https://github.com/thespecialone1/subdomain-enum/discussions)
- 💬 **Community**: [Discord Server](#) (coming soon)

### Commercial Support
For enterprise deployments, custom features, or professional support:
- 📧 **Email**: support@subdomain-enum.com
- 📅 **Consulting**: Available for custom implementations
- 🏢 **Enterprise**: Volume licensing and dedicated support

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **OWASP** for security best practices
- **Wayback Machine** (Internet Archive) for historical data
- **Certificate Transparency** projects for SSL data
- **DNS community** for public resolvers
- **Go community** for excellent libraries
- **Contributors** who make this project better

## 🎯 Roadmap

### v2.3.0 (Next Release)
- 🔄 **Redis caching** for improved performance
- 🔍 **Advanced filtering** and search capabilities
- 📊 **Custom wordlists** upload functionality
- 🔗 **Webhook notifications** for completed scans
- 🌍 **Multi-language** interface support

### v3.0.0 (Future)
- 🤖 **Machine learning** for pattern detection
- 📱 **Mobile app** for iOS and Android
- 🏢 **Multi-tenant** architecture
- 🔐 **Advanced authentication** and authorization
- ☁️ **Cloud integrations** (AWS, GCP, Azure)

---

<div align="center">

**Made with ❤️ by Security Researchers**

[⭐ Star us on GitHub](https://github.com/thespecialone1/subdomain-enum) | [🐳 Pull from Docker Hub](https://ghcr.io/thespecialone1/subdomain-enum) | [📋 Report Issues](https://github.com/thespecialone1/subdomain-enum/issues)

</div>