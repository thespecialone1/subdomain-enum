# Subdomain Enumerator

A fast and efficient subdomain enumeration tool that discovers subdomains using multiple sources including Wayback Machine and Certificate Transparency logs (crt.sh).

## Features

- 🔍 **Multiple Data Sources**: Queries both Wayback Machine and Certificate Transparency logs
- 🚀 **Real-time Streaming**: Results appear as they're discovered
- 🌐 **Web Interface**: Clean, modern web UI for easy usage
- ⚡ **Fast & Concurrent**: Efficient parallel processing
- 🔒 **SSL Probe**: Built-in HTTP/HTTPS probing with SSL certificate validation bypass
- 🛑 **Cancellable Jobs**: Stop running enumeration jobs at any time

## Quick Start (No Go Installation Required)

### Windows Users
1. Download `subdomain-enum-windows.exe` from the releases
2. Open Command Prompt or PowerShell
3. Navigate to the download folder
4. Run: `subdomain-enum-windows.exe`
5. Open your browser and go to `http://localhost:8080`

### macOS Users
1. Download `subdomain-enum-macos` from the releases
2. Open Terminal
3. Navigate to the download folder
4. Make it executable: `chmod +x subdomain-enum-macos`
5. Run: `./subdomain-enum-macos`
6. Open your browser and go to `http://localhost:8080`

### Linux Users
1. Download `subdomain-enum-linux` from the releases
2. Open Terminal
3. Navigate to the download folder
4. Make it executable: `chmod +x subdomain-enum-linux`
5. Run: `./subdomain-enum-linux`
6. Open your browser and go to `http://localhost:8080`

## How to Use

1. **Start the Server**: Run the appropriate executable for your operating system
2. **Open Web Interface**: Navigate to `http://localhost:8080` in your web browser
3. **Enter Target Domain**: Type the domain you want to enumerate (e.g., `example.com`)
4. **Choose Sources**: Select Wayback Machine, crt.sh, or both
5. **Start Enumeration**: Click the start button and watch results stream in real-time
6. **Probe Subdomains**: Click on any discovered subdomain to probe its HTTP status and title

## Building from Source (For Developers)

If you have Go installed and want to build from source:

```bash
# Clone the repository
git clone https://github.com/thespecialone1/subdomain-enum.git
cd subdomain-enum

# Run directly
go run cmd/server/main.go

# Or build an executable
go build -o subdomain-enum cmd/server/main.go
```

### Cross-compilation for Different Platforms

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o subdomain-enum-windows.exe cmd/server/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o subdomain-enum-macos cmd/server/main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o subdomain-enum-linux cmd/server/main.go
```

## API Endpoints

- `GET /` - Web interface
- `GET /api/wayback/stream?target=domain.com` - Stream subdomains from Wayback Machine
- `GET /api/crtsh/stream?target=domain.com` - Stream subdomains from Certificate Transparency logs
- `GET /api/probe?url=https://subdomain.domain.com` - Probe a URL for HTTP status and title
- `POST /api/abort?target=domain.com` - Cancel running enumeration jobs

## Configuration

The server runs on port 8080 by default. The web interface files should be in a `public/` directory relative to the executable.

## Troubleshooting

**Port already in use**: If port 8080 is busy, stop other services using that port or modify the source code to use a different port.

**Permission denied (macOS/Linux)**: Make sure to run `chmod +x` on the executable file.

**Antivirus warnings**: Some antivirus software may flag the executable. This is a false positive common with Go binaries.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is open source and available under the MIT License.

## Disclaimer

This tool is for educational and authorized security testing purposes only. Always ensure you have permission before testing domains you don't own.
