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

## Cloud Deployment

Deploy this application to various cloud platforms:

### Render.com

1. **Fork/Clone** this repository to your GitHub account
2. **Connect to Render**:
   - Go to [render.com](https://render.com) and sign up
   - Click "New" → "Web Service"
   - Connect your GitHub repository
3. **Configuration**:
   - **Name**: `subdomain-enum` (or your preferred name)
   - **Environment**: `Go`
   - **Region**: Choose closest to your users
   - **Branch**: `master` (or `main`)
   - **Root Directory**: `.` (leave empty or use dot for repository root)
   - **Build Command**: `go build -o main cmd/server/main.go`
   - **Start Command**: `./main`
   - **Instance Type**: `Free` (or upgrade as needed)
4. **Advanced Settings** (Optional):
   - **Auto-Deploy**: `Yes` (deploys on every push)
   - **Environment Variables**: None required (PORT is auto-set)
5. **Deploy**: Click "Create Web Service"

**Important**: Since your main.go is in `cmd/server/`, make sure the build command points to the correct path. The root directory should be `.` (repository root) so Render can access all files.

### Railway.app

1. **Connect Repository**:
   - Go to [railway.app](https://railway.app) and sign up
   - Click "New Project" → "Deploy from GitHub repo"
   - Select your forked repository
2. **Configuration** (Railway auto-detects Go):
   - **Build Command**: `go build -o main cmd/server/main.go`
   - **Start Command**: `./main`
3. **Environment Variables** (if needed):
   - `PORT`: Railway provides this automatically
4. **Deploy**: Railway will automatically deploy

### Heroku

1. **Install Heroku CLI** and login:
   ```bash
   heroku login
   ```

2. **Create Heroku app**:
   ```bash
   heroku create your-app-name
   ```

3. **Add Go buildpack**:
   ```bash
   heroku buildpacks:set heroku/go
   ```

4. **Create Procfile** in your project root:
   ```
   web: ./bin/subdomain-enum
   ```

5. **Update go.mod for Heroku** (add this to ensure proper module path):
   ```bash
   go mod tidy
   ```

6. **Deploy**:
   ```bash
   git add .
   git commit -m "Add Heroku deployment config"
   git push heroku master
   ```

### Google Cloud Run

1. **Create Dockerfile** in project root:
   ```dockerfile
   # Build stage
   FROM golang:1.21-alpine AS builder
   WORKDIR /app
   COPY go.mod go.sum ./
   RUN go mod download
   COPY . .
   RUN go build -o main cmd/server/main.go
   
   # Runtime stage
   FROM alpine:latest
   RUN apk --no-cache add ca-certificates
   WORKDIR /root/
   COPY --from=builder /app/main .
   COPY --from=builder /app/public ./public
   EXPOSE 8080
   CMD ["./main"]
   ```

2. **Deploy to Cloud Run**:
   ```bash
   gcloud run deploy subdomain-enum \
     --source . \
     --platform managed \
     --region us-central1 \
     --allow-unauthenticated
   ```

### DigitalOcean App Platform

1. **Connect Repository**:
   - Go to DigitalOcean → Apps → Create App
   - Connect your GitHub repository

2. **App Spec Configuration**:
   ```yaml
   name: subdomain-enum
   services:
   - name: web
     source_dir: /
     github:
       repo: your-username/subdomain-enum
       branch: master
     run_command: ./main
     build_command: go build -o main cmd/server/main.go
     environment_slug: go
     instance_count: 1
     instance_size_slug: basic-xxs
     http_port: 8080
   ```

### Fly.io

1. **Install Fly CLI** and login:
   ```bash
   flyctl auth login
   ```

2. **Initialize Fly app**:
   ```bash
   flyctl launch
   ```

3. **Update fly.toml** if needed:
   ```toml
   [build]
     builder = "paketobuildpacks/builder:base"
   
   [[services]]
     http_checks = []
     internal_port = 8080
     processes = ["app"]
     protocol = "tcp"
     script_checks = []
   
     [[services.ports]]
       force_https = true
       handlers = ["http"]
       port = 80
   
     [[services.ports]]
       handlers = ["tls", "http"]
       port = 443
   ```

4. **Deploy**:
   ```bash
   flyctl deploy
   ```

### Environment Variables for Cloud Deployment

Most cloud platforms will automatically set `PORT`, but you can customize:

- `PORT`: The port your app listens on (default: 8080)
- `GO_ENV`: Set to `production` for production builds

### Important Notes for Cloud Deployment

1. **Port Configuration**: Most cloud platforms expect your app to listen on the port specified by the `PORT` environment variable. You may need to modify the main.go to use `os.Getenv("PORT")` instead of hardcoded `:8080`.

2. **Static Files**: Ensure the `public/` directory is included in your deployment.

3. **Build Optimization**: For production, consider using build flags:
   ```bash
   go build -ldflags="-s -w" -o main cmd/server/main.go
   ```

4. **Health Checks**: Some platforms require health check endpoints. Consider adding a `/health` endpoint.

### Render-Specific Troubleshooting

**Root Directory Issues:**
- **Problem**: "No such file or directory" errors during build
- **Solution**: Set Root Directory to `.` (dot) or leave empty
- **Explanation**: Render needs access to the entire repository structure

**Build Command Issues:**
- **Problem**: Build fails with "package not found"
- **Solution**: Use full path: `go build -o main cmd/server/main.go`
- **Alternative**: If issues persist, try: `go build -o main ./cmd/server`

**Static Files Not Loading:**
- **Problem**: Web interface shows "file not found"
- **Solution**: Ensure `public/` directory is in repository root
- **Check**: Verify in Render logs that files are copied during build

**Common Render Settings:**
```
Name: subdomain-enum
Environment: Go
Build Command: go build -o main cmd/server/main.go
Start Command: ./main
Root Directory: . (or leave empty)
```

### Quick Port Fix for Cloud Deployment

To make your app work with cloud platforms, update the main.go server startup:

```go
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}
log.Printf("Listening on port %s...", port)
log.Fatal(http.ListenAndServe(":"+port, mux))
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
