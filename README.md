# Advanced Subdomain Enumeration Tool v2.0

A comprehensive, high-performance subdomain enumeration platform that discovers subdomains using multiple sources and techniques. Features real-time streaming results, intelligent timeout handling, and an intuitive web interface.

## 🚀 New Features v2.0

### Multi-Source Discovery
- **🔍 Wayback Machine**: Historical subdomain discovery from archived web pages
- **🔒 Certificate Transparency**: Subdomain discovery from SSL/TLS certificate logs (crt.sh)
- **🌐 DNS Brute Force**: Dictionary-based DNS resolution with 500+ common subdomains
- **🔎 Search Engine Scraping**: Automated Google search result parsing
- **🔄 Permutation Generation**: Intelligent subdomain permutation with common patterns
- **📡 Zone Transfer Attempts**: DNS zone transfer testing (for misconfigured servers)

### Enhanced Interface
- **⚡ Real-time Streaming**: Results appear instantly as they're discovered
- **📊 Live Statistics**: Real-time counters and active source monitoring
- **🎛️ Source Selection**: Enable/disable individual discovery methods
- **🔍 Advanced Probing**: HTTP/HTTPS status checking with title extraction
- **📋 Export Options**: Copy results by source or combined
- **⏱️ Smart Timeouts**: Automatic termination prevents infinite scanning

## Quick Start

### Using Pre-built Binaries (Recommended)

#### Windows
```cmd
# Download the latest release
curl -L -o subdomain-enum-windows.exe https://github.com/thespecialone1/subdomain-enum/releases/latest/download/subdomain-enum-windows.exe

# Run the tool
subdomain-enum-windows.exe

# Open browser and navigate to http://localhost:8080
```

#### macOS
```bash
# Download and make executable
curl -L -o subdomain-enum-macos https://github.com/thespecialone1/subdomain-enum/releases/latest/download/subdomain-enum-macos
chmod +x subdomain-enum-macos

# Run the tool
./subdomain-enum-macos

# Open browser and navigate to http://localhost:8080
```

#### Linux
```bash
# Download and make executable
curl -L -o subdomain-enum-linux https://github.com/thespecialone1/subdomain-enum/releases/latest/download/subdomain-enum-linux
chmod +x subdomain-enum-linux

# Run the tool
./subdomain-enum-linux

# Open browser and navigate to http://localhost:8080
```

### Using Docker

```bash
# Build the image
docker build -t subdomain-enum:v2 .

# Run the container
docker run -p 8080:8080 subdomain-enum:v2

# Access via http://localhost:8080
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/thespecialone1/subdomain-enum.git
cd subdomain-enum

# Install dependencies
go mod download

# Run directly
go run cmd/server/main.go

# Or build an executable
go build -o subdomain-enum cmd/server/main.go
```

## 📚 Usage Guide

### Web Interface
1. **Start the Server**: Run the executable for your platform
2. **Open Browser**: Navigate to `http://localhost:8080`
3. **Configure Sources**: Select which discovery methods to use
4. **Enter Target**: Input the domain to enumerate (e.g., `example.com`)
5. **Start Scan**: Click "Start Scan" and watch results stream in real-time
6. **Review Results**: Click any subdomain for detailed HTTP probe information
7. **Export Data**: Use "Copy All" buttons to export results

### Discovery Methods Explained

#### 1. Wayback Machine
- Searches Internet Archive's historical web crawl data
- Discovers subdomains from archived URLs
- Excellent for finding old/deprecated subdomains
- **Timeout**: 5 minutes

#### 2. Certificate Transparency (crt.sh)
- Queries public SSL/TLS certificate logs
- Finds subdomains from certificate Subject Alternative Names
- Great for discovering active HTTPS subdomains
- **Timeout**: 5 minutes

#### 3. DNS Brute Force
- Tests 500+ common subdomain patterns
- Uses multiple DNS servers (8.8.8.8, 1.1.1.1)
- Concurrent resolution with rate limiting
- **Timeout**: 10 minutes

#### 4. Search Engine Scraping
- Automated Google search with "site:" operator
- Extracts subdomains from search results
- Respects search engine rate limits
- **Timeout**: 5 minutes

#### 5. Permutation Generation
- Creates intelligent subdomain variations
- Combines prefixes, suffixes, and patterns
- Tests common development/staging patterns
- **Timeout**: 10 minutes

#### 6. Zone Transfer Attempts
- Tests for DNS zone transfer misconfigurations
- Attempts AXFR requests against nameservers
- Rarely successful but worth checking
- **Timeout**: 2 minutes

## 🌐 Cloud Deployment

### Render.com (Recommended)
1. Fork this repository to your GitHub account
2. Connect to [Render.com](https://render.com) and create a new Web Service
3. Configure settings:
   - **Build Command**: `go build -o main cmd/server/main.go`
   - **Start Command**: `./main`
   - **Environment**: Go
   - **Instance Type**: Free tier supported

### Railway.app
1. Connect repository at [Railway.app](https://railway.app)
2. Railway auto-detects Go and configures build
3. Automatic HTTPS and custom domains available

### Heroku
```bash
# Create app and set buildpack
heroku create your-app-name
heroku buildpacks:set heroku/go

# Create Procfile
echo "web: ./bin/subdomain-enum" > Procfile

# Deploy
git push heroku main
```

### Google Cloud Run
```bash
# Deploy with Cloud Build
gcloud run deploy subdomain-enum \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

### DigitalOcean App Platform
- Connect GitHub repository
- Select Go environment
- Deploy with automatic scaling

## 🔧 API Endpoints

### Streaming Endpoints
- `GET /api/wayback/stream?target=domain.com` - Wayback Machine results
- `GET /api/crtsh/stream?target=domain.com` - Certificate Transparency results  
- `GET /api/dns/stream?target=domain.com` - DNS brute force results
- `GET /api/search/stream?target=domain.com` - Search engine results
- `GET /api/permute/stream?target=domain.com` - Permutation results
- `GET /api/zone/stream?target=domain.com` - Zone transfer results

### Control Endpoints
- `GET /api/probe?url=https://subdomain.domain.com` - Probe URL for HTTP status/title
- `POST /api/abort?target=domain.com` - Cancel all running scans for target
- `GET /api/status?target=domain.com` - Get scan status and statistics

### Response Formats

#### Stream Response (SSE)
```
data: subdomain.example.com

data: api.example.com

data: www.example.com
```

#### Probe Response (JSON)
```json
{
  "status": "200",
  "title": "Example Website",
  "error": ""
}
```

## ⚙️ Configuration

### Environment Variables
- `PORT`: Server port (default: 8080)
- `TIMEOUT_WAYBACK`: Wayback timeout in minutes (default: 5)
- `TIMEOUT_CRTSH`: Certificate transparency timeout (default: 5)
- `TIMEOUT_DNS`: DNS brute force timeout (default: 10)
- `TIMEOUT_SEARCH`: Search engine timeout (default: 5)
- `TIMEOUT_PERMUTE`: Permutation timeout (default: 10)
- `TIMEOUT_ZONE`: Zone transfer timeout (default: 2)

### Performance Tuning
- DNS queries are limited to 50 concurrent requests
- HTTP probes have 10-second timeouts
- Search engine queries respect rate limits
- Memory usage optimized with result streaming

## 🔍 Advanced Usage

### Batch Processing
```bash
# Use curl to automate scans
curl -N "http://localhost:8080/api/wayback/stream?target=example.com" | \
  while read line; do
    echo "Found: ${line#data: }"
  done
```

### Custom Wordlists
The DNS brute force uses a built-in wordlist of 500+ common subdomains. To use custom wordlists, modify the `commonSubdomains` array in `main.go`.

### Rate Limiting
- Built-in rate limiting prevents API abuse
- Concurrent DNS queries are limited to prevent flooding
- HTTP probes include delays between requests

## 🚨 Security Considerations

### Ethical Usage
- Only scan domains you own or have explicit permission to test
- Respect robots.txt and rate limits
- Some techniques may trigger security monitoring

### Firewall Considerations
- Tool makes outbound HTTPS requests to various APIs
- DNS queries to 8.8.8.8 and other public resolvers
- HTTP/HTTPS probes to discovered subdomains

### Privacy
- No scan data is stored permanently
- Results are only kept in memory during scan session
- No tracking or analytics implemented

## 📊 Performance Benchmarks

### Typical Performance
- **Small Domain** (< 100 subdomains): 2-5 minutes
- **Medium Domain** (100-1000 subdomains): 5-15 minutes  
- **Large Domain** (1000+ subdomains): 15-30 minutes

### Resource Usage
- **Memory**: 50-200MB during active scans
- **CPU**: Low to moderate during DNS brute force
- **Network**: Moderate outbound traffic for API calls

## 🛠️ Development

### Project Structure
```
subdomain-enum/
├── cmd/server/main.go          # Main application
├── public/index.html           # Web interface
├── go.mod                      # Go dependencies
├── Dockerfile                  # Container configuration
└── README.md                   # Documentation
```

### Adding New Sources
1. Create new stream handler function
2. Add endpoint registration in main()
3. Update HTML interface with new source panel
4. Add JavaScript event handling

### Contributing
1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 🐛 Troubleshooting

### Common Issues

#### "Port already in use"
```bash
# Find process using port 8080
lsof -i :8080

# Kill process or use different port
PORT=8081 ./subdomain-enum
```

#### "Permission denied" (macOS/Linux)
```bash
# Make executable
chmod +x subdomain-enum-linux
```

#### "No results found"
- Verify domain name is correct
- Check internet connectivity
- Some APIs may be temporarily unavailable
- Try different source combinations

#### High CPU usage
- Normal during DNS brute force phase
- Reduce concurrent DNS queries in source code if needed
- Consider running on more powerful hardware

### Debug Mode
```bash
# Enable verbose logging
DEBUG=1 ./subdomain-enum
```

## 📈 Changelog

### v2.0.0 (Current)
- ✅ Improved error handling and status reporting
- ✅ Added Docker support with distroless base image
- ✅ Enhanced security with non-root container execution

### v1.0.0
- ✅ Basic Wayback Machine integration
- ✅ Certificate Transparency (crt.sh) support
- ✅ Real-time streaming interface
- ✅ HTTP/HTTPS probing
- ✅ Basic web interface

## 📋 Roadmap

### v2.1.0 (Planned)
- 🔄 Custom wordlist upload support
- 🔄 API rate limiting configuration
- 🔄 Result export to JSON/CSV
- 🔄 Subdomain takeover detection
- 🔄 Integration with external DNS APIs

### v2.2.0 (Future)
- 🔄 Machine learning for subdomain prediction
- 🔄 Integration with threat intelligence feeds
- 🔄 Advanced filtering and sorting options
- 🔄 Automated report generation
- 🔄 Multi-domain batch processing

## 🤝 Community

### Getting Help
- 📖 Check this README for common solutions
- 🐛 Report bugs via GitHub Issues
- 💡 Request features via GitHub Discussions
- 📧 Contact: [your-email@domain.com]

### Contributing
We welcome contributions! Areas where help is needed:
- 🌐 Additional discovery sources
- 🎨 UI/UX improvements
- 📚 Documentation enhancements
- 🧪 Testing and quality assurance
- 🔒 Security auditing

### Recognition
Special thanks to:
- Internet Archive for Wayback Machine API
- Certificate Transparency community
- Open source DNS resolver providers
- Security research community

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

```
MIT License

Copyright (c) 2024 Subdomain Enumeration Tool Contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## ⚠️ Disclaimer

This tool is designed for legitimate security testing and educational purposes only. Users are solely responsible for complying with applicable laws and regulations. The developers assume no liability for misuse of this software.

### Legal Considerations
- ✅ Only test domains you own or have explicit written permission to test
- ✅ Respect rate limits and terms of service of external APIs
- ✅ Be aware that some enumeration techniques may be detected by security monitoring
- ✅ Consider informing domain owners of discovered vulnerabilities responsibly
- ✅ Ensure compliance with local laws and regulations regarding security testing

### Ethical Guidelines
1. **Permission First**: Always obtain proper authorization before testing
2. **Responsible Disclosure**: Report findings through appropriate channels
3. **Minimize Impact**: Use techniques that don't disrupt normal operations
4. **Stay Legal**: Understand and comply with applicable laws in your jurisdiction
5. **Professional Use**: Use this tool to improve security, not to cause harm

---

## 🔧 Technical Specifications

### System Requirements
- **Minimum RAM**: 512MB
- **Recommended RAM**: 2GB or higher
- **CPU**: Any modern processor (ARM64 and AMD64 supported)
- **Storage**: 50MB for application, additional space for logs
- **Network**: Outbound internet access required

### Supported Platforms
- **Windows**: 10, 11, Server 2016+
- **macOS**: 10.15+ (Catalina and newer)
- **Linux**: Most distributions with kernel 3.10+
- **Docker**: Any platform supporting Docker 20.10+

### Browser Compatibility
- **Chrome/Chromium**: 80+
- **Firefox**: 75+
- **Safari**: 13+
- **Edge**: 80+

### Network Requirements
The tool requires outbound access to:
- `web.archive.org` (port 443) - Wayback Machine API
- `crt.sh` (port 443) - Certificate Transparency logs
- `8.8.8.8` (port 53) - Google DNS
- `1.1.1.1` (port 53) - Cloudflare DNS
- `www.google.com` (port 443) - Search engine queries
- Target domains (ports 80, 443) - HTTP/HTTPS probing

---

**Made with ❤️ by the Security Research Community**

*Star ⭐ this repository if you find it useful!* Added DNS brute force enumeration
- ✅ Added search engine scraping
- ✅ Added permutation generation  
- ✅ Added zone transfer attempts
- ✅ Implemented smart timeouts
- ✅ Enhanced web interface with source selection
- ✅ Added real-time statistics
- ✅

## License

This project is open source and available under the MIT License.

## Disclaimer

This tool is for educational and authorized security testing purposes only. Always ensure you have permission before testing domains you don't own.
