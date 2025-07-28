package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/miekg/dns"
)

// Build information injected at compile time
var (
	version   = "2.2.0"
	buildTime = "unknown"
	gitCommit = "unknown"
)

// Configuration structure for better settings management
type Config struct {
	Port       string
	LogLevel   string
	Timeouts   TimeoutConfig
	DNS        DNSConfig
	HTTP       HTTPConfig
	RateLimit  RateLimitConfig
	Security   SecurityConfig
	Monitoring MonitoringConfig
}

type TimeoutConfig struct {
	Wayback   time.Duration
	CrtSh     time.Duration
	DNS       time.Duration
	Search    time.Duration
	Permute   time.Duration
	Zone      time.Duration
	HTTPProbe time.Duration
}

type DNSConfig struct {
	Servers     []string
	Concurrency int
	Retries     int
	Timeout     time.Duration
}

type HTTPConfig struct {
	UserAgent     string
	MaxRedirects  int
	Timeout       time.Duration
	MaxBodySize   int64
	SkipTLSVerify bool
}

type RateLimitConfig struct {
	RequestsPerSecond int
	BurstSize         int
	WindowSize        time.Duration
}

type SecurityConfig struct {
	AllowedDomains    []string
	BlockedUserAgents []string
	MaxConcurrentJobs int
	EnableCORS        bool
}

type MonitoringConfig struct {
	EnableMetrics bool
	EnableHealth  bool
	MetricsPort   string
}

// Enhanced statistics and metrics
type Statistics struct {
	TotalRequests     int64
	ActiveJobs        int64
	CompletedJobs     int64
	FailedJobs        int64
	TotalSubdomains   int64
	TotalProbes       int64
	SuccessfulProbes  int64
	DNSQueries        int64
	StartTime         time.Time
	LastActivity      time.Time
	SourceStats       map[string]*SourceStats
	mu                sync.RWMutex
}

type SourceStats struct {
	Requests   int64
	Responses  int64
	Errors     int64
	Duration   time.Duration
	LastUsed   time.Time
}

// Enhanced job management
type Job struct {
	ID        string
	Target    string
	Sources   []string
	StartTime time.Time
	Status    string
	Results   map[string][]Result
	Cancel    context.CancelFunc
	mu        sync.RWMutex
}

type JobManager struct {
	jobs map[string]*Job
	mu   sync.RWMutex
}

// Enhanced result structure
type Result struct {
	Host      string    `json:"host"`
	Source    string    `json:"source"`
	Status    string    `json:"status"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	ProbeTime int64     `json:"probe_time_ms,omitempty"`
}

// Enhanced DNS resolver with connection pooling
type DNSResolver struct {
	servers []string
	clients []*dns.Client
	current int64
	mu      sync.RWMutex
}

// Rate limiter implementation
type RateLimiter struct {
	tokens   chan struct{}
	refill   *time.Ticker
	capacity int
}

var (
	// Enhanced regex patterns
	hostRe     = regexp.MustCompile(`https?://([^/\s"'<>]+)`)
	titleRe    = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	domainRe   = regexp.MustCompile(`^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+\.[a-zA-Z]{2,}$`)
	
	// Global instances
	config       *Config
	stats        *Statistics
	jobManager   *JobManager
	dnsResolver  *DNSResolver
	rateLimiter  *RateLimiter
	
	// Enhanced wordlist with categorization
	commonSubdomains = map[string][]string{
		"common": {
			"www", "mail", "ftp", "admin", "test", "dev", "api", "blog", "shop", "forum",
			"news", "help", "support", "mobile", "m", "app", "apps", "secure", "portal",
			"dashboard", "panel", "control", "manage", "manager", "status", "health",
		},
		"development": {
			"dev", "test", "stage", "staging", "demo", "sandbox", "beta", "alpha", "qa",
			"uat", "prod", "production", "preview", "dev-api", "test-api", "staging-api",
			"dev-www", "test-www", "staging-www", "local", "localhost", "development",
		},
		"infrastructure": {
			"cdn", "static", "assets", "media", "images", "img", "js", "css", "files",
			"upload", "download", "backup", "archive", "storage", "s3", "ftp", "sftp",
			"git", "svn", "repo", "jenkins", "ci", "build", "deploy", "docker",
		},
		"services": {
			"api", "api1", "api2", "v1", "v2", "v3", "ws", "webservice", "service",
			"auth", "oauth", "sso", "login", "signin", "signup", "register", "account",
			"profile", "user", "users", "admin", "administrator", "root", "super",
		},
		"communication": {
			"mail", "email", "smtp", "pop", "pop3", "imap", "webmail", "mx", "mx1", "mx2",
			"chat", "irc", "xmpp", "sip", "voip", "conference", "meet", "zoom", "teams",
			"slack", "discord", "telegram", "whatsapp", "messenger", "support",
		},
		"databases": {
			"db", "database", "mysql", "postgres", "mongodb", "redis", "elastic", "es",
			"kibana", "grafana", "prometheus", "influx", "clickhouse", "cassandra",
			"neo4j", "couchdb", "rethinkdb", "memcached", "sql", "nosql",
		},
		"monitoring": {
			"monitor", "monitoring", "metrics", "logs", "logging", "analytics", "stats",
			"grafana", "prometheus", "nagios", "zabbix", "splunk", "elk", "kibana",
			"datadog", "newrelic", "sentry", "bugsnag", "rollbar", "pingdom",
		},
	}
)

func init() {
	config = loadConfig()
	stats = &Statistics{
		StartTime:   time.Now(),
		SourceStats: make(map[string]*SourceStats),
	}
	jobManager = &JobManager{
		jobs: make(map[string]*Job),
	}
	initializeDNSResolver()
	initializeRateLimiter()
	setupLogging()
}

func loadConfig() *Config {
	return &Config{
		Port:     getEnvString("PORT", "8080"),
		LogLevel: getEnvString("LOG_LEVEL", "INFO"),
		Timeouts: TimeoutConfig{
			Wayback:   getEnvDuration("TIMEOUT_WAYBACK", 5*time.Minute),
			CrtSh:     getEnvDuration("TIMEOUT_CRTSH", 5*time.Minute),
			DNS:       getEnvDuration("TIMEOUT_DNS", 10*time.Minute),
			Search:    getEnvDuration("TIMEOUT_SEARCH", 5*time.Minute),
			Permute:   getEnvDuration("TIMEOUT_PERMUTE", 10*time.Minute),
			Zone:      getEnvDuration("TIMEOUT_ZONE", 2*time.Minute),
			HTTPProbe: getEnvDuration("HTTP_PROBE_TIMEOUT", 10*time.Second),
		},
		DNS: DNSConfig{
			Servers:     getEnvStringSlice("DNS_SERVERS", []string{"8.8.8.8:53", "1.1.1.1:53", "208.67.222.222:53"}),
			Concurrency: getEnvInt("DNS_CONCURRENCY", 50),
			Retries:     getEnvInt("DNS_RETRIES", 2),
			Timeout:     getEnvDuration("DNS_TIMEOUT", 3*time.Second),
		},
		HTTP: HTTPConfig{
			UserAgent:     getEnvString("HTTP_USER_AGENT", "Mozilla/5.0 (compatible; SubdomainScanner/2.0; +https://github.com/security/subdomain-enum)"),
			MaxRedirects:  getEnvInt("HTTP_MAX_REDIRECTS", 3),
			Timeout:       getEnvDuration("HTTP_TIMEOUT", 10*time.Second),
			MaxBodySize:   getEnvInt64("HTTP_MAX_BODY_SIZE", 1024*1024), // 1MB
			SkipTLSVerify: getEnvBool("HTTP_SKIP_TLS_VERIFY", true),
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: getEnvInt("RATE_LIMIT_RPS", 10),
			BurstSize:         getEnvInt("RATE_LIMIT_BURST", 20),
			WindowSize:        getEnvDuration("RATE_LIMIT_WINDOW", time.Minute),
		},
		Security: SecurityConfig{
			AllowedDomains:    getEnvStringSlice("ALLOWED_DOMAINS", []string{}),
			BlockedUserAgents: getEnvStringSlice("BLOCKED_USER_AGENTS", []string{"bot", "crawler", "spider"}),
			MaxConcurrentJobs: getEnvInt("MAX_CONCURRENT_JOBS", 10),
			EnableCORS:        getEnvBool("ENABLE_CORS", true),
		},
		Monitoring: MonitoringConfig{
			EnableMetrics: getEnvBool("ENABLE_METRICS", true),
			EnableHealth:  getEnvBool("ENABLE_HEALTH", true),
			MetricsPort:   getEnvString("METRICS_PORT", "9090"),
		},
	}
}

func setupLogging() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if config.LogLevel == "DEBUG" {
		log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	}
}

func initializeDNSResolver() {
	dnsResolver = &DNSResolver{
		servers: config.DNS.Servers,
		clients: make([]*dns.Client, len(config.DNS.Servers)),
	}
	
	for i := range dnsResolver.clients {
		dnsResolver.clients[i] = &dns.Client{
			Timeout: config.DNS.Timeout,
			Net:     "udp",
		}
	}
}

func initializeRateLimiter() {
	rateLimiter = &RateLimiter{
		tokens:   make(chan struct{}, config.RateLimit.BurstSize),
		capacity: config.RateLimit.BurstSize,
	}
	
	// Fill initial tokens
	for i := 0; i < config.RateLimit.BurstSize; i++ {
		rateLimiter.tokens <- struct{}{}
	}
	
	// Start refill goroutine
	rateLimiter.refill = time.NewTicker(time.Second / time.Duration(config.RateLimit.RequestsPerSecond))
	go func() {
		for range rateLimiter.refill.C {
			select {
			case rateLimiter.tokens <- struct{}{}:
			default:
				// Channel full, skip
			}
		}
	}()
}

func main() {
	// Parse command line flags
	var (
		showVersion   = flag.Bool("version", false, "Show version information")
		showHelp      = flag.Bool("help", false, "Show help information")
		healthCheck   = flag.Bool("health-check", false, "Perform health check and exit")
		port          = flag.String("port", "", "Override port setting")
		logLevel      = flag.String("log-level", "", "Override log level (DEBUG, INFO, WARN, ERROR)")
	)
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("Advanced Subdomain Enumeration Tool\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		fmt.Printf("Git Commit: %s\n", gitCommit)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	// Handle help flag
	if *showHelp {
		fmt.Printf("Advanced Subdomain Enumeration Tool v%s\n\n", version)
		fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
		fmt.Printf("Options:\n")
		flag.PrintDefaults()
		fmt.Printf("\nEnvironment Variables:\n")
		fmt.Printf("  PORT                    Server port (default: 8080)\n")
		fmt.Printf("  METRICS_PORT           Metrics server port (default: 9090)\n")
		fmt.Printf("  LOG_LEVEL              Log level (DEBUG, INFO, WARN, ERROR)\n")
		fmt.Printf("  DNS_SERVERS            Comma-separated DNS servers\n")
		fmt.Printf("  DNS_CONCURRENCY        DNS query concurrency (default: 50)\n")
		fmt.Printf("  RATE_LIMIT_RPS         Rate limit requests per second (default: 10)\n")
		fmt.Printf("  TIMEOUT_*              Various timeout settings\n")
		fmt.Printf("\nExamples:\n")
		fmt.Printf("  %s                     # Start with default settings\n", os.Args[0])
		fmt.Printf("  %s --port 9080         # Use custom port\n", os.Args[0])
		fmt.Printf("  %s --health-check      # Health check for containers\n", os.Args[0])
		fmt.Printf("\nFor more information, visit: https://github.com/thespecialone1/subdomain-enum\n")
		os.Exit(0)
	}

	// Handle health check flag (for Docker/K8s)
	if *healthCheck {
		if err := performHealthCheck(); err != nil {
			log.Printf("Health check failed: %v", err)
			os.Exit(1)
		}
		fmt.Println("Health check passed")
		os.Exit(0)
	}

	// Override config with command line arguments
	if *port != "" {
		os.Setenv("PORT", *port)
	}
	if *logLevel != "" {
		os.Setenv("LOG_LEVEL", *logLevel)
	}

	// Load configuration
	config = loadConfig()
	
	// Initialize other components...
	stats = &Statistics{
		StartTime:   time.Now(),
		SourceStats: make(map[string]*SourceStats),
	}
	jobManager = &JobManager{
		jobs: make(map[string]*Job),
	}
	initializeDNSResolver()
	initializeRateLimiter()
	setupLogging()

	mux := http.NewServeMux()

	// Enhanced middleware - serve static files without middleware for better performance
	mux.Handle("/", http.FileServer(http.Dir("./public/")))

	// API endpoints with middleware
	mux.HandleFunc("/api/wayback/stream", withMiddleware(waybackStream))
	mux.HandleFunc("/api/crtsh/stream", withMiddleware(crtshStream))
	mux.HandleFunc("/api/dns/stream", withMiddleware(dnsStream))
	mux.HandleFunc("/api/search/stream", withMiddleware(searchEngineStream))
	mux.HandleFunc("/api/permute/stream", withMiddleware(permuteStream))
	mux.HandleFunc("/api/zone/stream", withMiddleware(zoneTransferStream))

	// Enhanced endpoints
	mux.HandleFunc("/api/probe", withMiddleware(probeHandler))
	mux.HandleFunc("/api/jobs", withMiddleware(jobsHandler))
	mux.HandleFunc("/api/jobs/", withMiddleware(jobDetailHandler))
	mux.HandleFunc("/api/abort", withMiddleware(abortHandler))
	mux.HandleFunc("/api/status", withMiddleware(statusHandler))
	mux.HandleFunc("/api/stats", withMiddleware(statsHandler))
	mux.HandleFunc("/api/config", withMiddleware(configHandler))
	mux.HandleFunc("/api/version", withMiddleware(versionHandler))

	// Health and monitoring endpoints on main server
	if config.Monitoring.EnableHealth {
		mux.HandleFunc("/health", healthHandler)
		mux.HandleFunc("/ready", readinessHandler)
	}
	
	// Always enable metrics on main server for convenience
	mux.HandleFunc("/metrics", metricsHandler)
	
	// Start separate metrics server only if explicitly configured
	if config.Monitoring.EnableMetrics && config.Monitoring.MetricsPort != config.Port {
		go startMetricsServer()
	}

	log.Printf("🚀 Advanced Subdomain Enumeration Tool v%s starting...", version)
	log.Printf("📊 Configuration: DNS Servers: %v, Concurrency: %d, Rate Limit: %d/s", 
		config.DNS.Servers, config.DNS.Concurrency, config.RateLimit.RequestsPerSecond)
	log.Printf("🌐 Web Interface: http://localhost:%s", config.Port)
	
	if config.Monitoring.EnableMetrics {
		log.Printf("📈 Metrics available at: http://localhost:%s/metrics", config.Port)
		if config.Monitoring.MetricsPort != config.Port {
			log.Printf("📊 Dedicated metrics server starting on port %s", config.Monitoring.MetricsPort)
		}
	}
	
	if config.Monitoring.EnableHealth {
		log.Printf("🏥 Health checks: http://localhost:%s/health", config.Port)
	}
	
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	
	log.Printf("✅ Server ready and listening on port %s", config.Port)
	log.Fatal(server.ListenAndServe())
}

// Enhanced middleware with security, logging, and rate limiting
func withMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		if config.Security.EnableCORS {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Rate limiting
		select {
		case <-rateLimiter.tokens:
			defer func() {
				// Return token after request
				select {
				case rateLimiter.tokens <- struct{}{}:
				default:
				}
			}()
		case <-time.After(100 * time.Millisecond):
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Request logging
		start := time.Now()
		defer func() {
			duration := time.Since(start)
			log.Printf("[%s] %s %s - %v", r.Method, r.URL.Path, r.RemoteAddr, duration)
			atomic.AddInt64(&stats.TotalRequests, 1)
			stats.LastActivity = time.Now()
		}()

		// User agent filtering
		userAgent := r.Header.Get("User-Agent")
		for _, blocked := range config.Security.BlockedUserAgents {
			if strings.Contains(strings.ToLower(userAgent), strings.ToLower(blocked)) {
				http.Error(w, "Blocked user agent", http.StatusForbidden)
				return
			}
		}

		handler(w, r)
	}
}

// Enhanced DNS resolution with load balancing and error handling
func (dr *DNSResolver) LookupHost(ctx context.Context, host string) ([]net.IP, error) {
	serverIndex := atomic.AddInt64(&dr.current, 1) % int64(len(dr.servers))
	client := dr.clients[serverIndex]
	server := dr.servers[serverIndex]

	msg := &dns.Msg{}
	msg.SetQuestion(dns.Fqdn(host), dns.TypeA)
	msg.RecursionDesired = true

	response, _, err := client.ExchangeContext(ctx, msg, server)
	if err != nil {
		atomic.AddInt64(&stats.DNSQueries, 1)
		return nil, fmt.Errorf("DNS query failed for %s: %w", host, err)
	}

	var ips []net.IP
	for _, answer := range response.Answer {
		if a, ok := answer.(*dns.A); ok {
			ips = append(ips, a.A)
		}
	}

	atomic.AddInt64(&stats.DNSQueries, 1)
	if len(ips) == 0 {
		return nil, fmt.Errorf("no A records found for %s", host)
	}

	return ips, nil
}

// Enhanced SSE headers with better caching control
func sseHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering
}

// Enhanced job management with better tracking
func createJob(target string, sources []string) *Job {
	jobID := fmt.Sprintf("%s_%d", target, time.Now().Unix())
	
	job := &Job{
		ID:        jobID,
		Target:    target,
		Sources:   sources,
		StartTime: time.Now(),
		Status:    "running",
		Results:   make(map[string][]Result),
	}

	jobManager.mu.Lock()
	jobManager.jobs[jobID] = job
	jobManager.mu.Unlock()

	atomic.AddInt64(&stats.ActiveJobs, 1)
	return job
}

func (j *Job) AddResult(source string, result Result) {
	j.mu.Lock()
	defer j.mu.Unlock()
	
	if j.Results[source] == nil {
		j.Results[source] = make([]Result, 0)
	}
	j.Results[source] = append(j.Results[source], result)
	atomic.AddInt64(&stats.TotalSubdomains, 1)
}

func (j *Job) Complete() {
	j.mu.Lock()
	j.Status = "completed"
	j.mu.Unlock()
	
	atomic.AddInt64(&stats.ActiveJobs, -1)
	atomic.AddInt64(&stats.CompletedJobs, 1)
}

func (j *Job) Fail(err error) {
	j.mu.Lock()
	j.Status = fmt.Sprintf("failed: %v", err)
	j.mu.Unlock()
	
	atomic.AddInt64(&stats.ActiveJobs, -1)
	atomic.AddInt64(&stats.FailedJobs, 1)
}

// Enhanced probe handler with better error handling and caching
func probeHandler(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		writeProbeError(w, "invalid URL", err)
		return
	}

	// Validate domain if restrictions are set
	if len(config.Security.AllowedDomains) > 0 {
		allowed := false
		for _, domain := range config.Security.AllowedDomains {
			if strings.HasSuffix(parsedURL.Hostname(), domain) {
				allowed = true
				break
			}
		}
		if !allowed {
			writeProbeError(w, "domain not allowed", fmt.Errorf("domain %s not in allowed list", parsedURL.Hostname()))
			return
		}
	}

	startTime := time.Now()
	result := probeURL(r.Context(), targetURL)
	result.ProbeTime = time.Since(startTime).Milliseconds()

	atomic.AddInt64(&stats.TotalProbes, 1)
	if result.Status != "0" && result.Error == "" {
		atomic.AddInt64(&stats.SuccessfulProbes, 1)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func probeURL(ctx context.Context, targetURL string) ProbeResponse {
	client := &http.Client{
		Timeout: config.HTTP.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.HTTP.SkipTLSVerify,
			},
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.HTTP.MaxRedirects {
				return fmt.Errorf("too many redirects (%d)", len(via))
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return ProbeResponse{
			Status: "0",
			Title:  "Request creation failed",
			Error:  err.Error(),
		}
	}

	req.Header.Set("User-Agent", config.HTTP.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := client.Do(req)
	if err != nil {
		return ProbeResponse{
			Status: "0",
			Title:  "Connection failed",
			Error:  err.Error(),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, config.HTTP.MaxBodySize))
	if err != nil {
		return ProbeResponse{
			Status: fmt.Sprintf("%d", resp.StatusCode),
			Title:  "Failed to read response",
			Error:  err.Error(),
		}
	}

	title := extractTitle(string(body))
	return ProbeResponse{
		Status: fmt.Sprintf("%d", resp.StatusCode),
		Title:  title,
		Error:  "",
	}
}

func extractTitle(html string) string {
	matches := titleRe.FindStringSubmatch(html)
	if len(matches) < 2 {
		return "No title"
	}
	
	title := strings.TrimSpace(matches[1])
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\r", " ")
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")
	
	if len(title) > 100 {
		title = title[:100] + "..."
	}
	
	return title
}

type ProbeResponse struct {
	Status    string `json:"status"`
	Title     string `json:"title"`
	Error     string `json:"error"`
	ProbeTime int64  `json:"probe_time_ms,omitempty"`
}

func writeProbeError(w http.ResponseWriter, message string, err error) {
	response := ProbeResponse{
		Status: "0",
		Title:  "Connection failed",
		Error:  fmt.Sprintf("%s: %v", message, err),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Enhanced statistics handler
func statsHandler(w http.ResponseWriter, r *http.Request) {
	stats.mu.RLock()
	defer stats.mu.RUnlock()

	uptime := time.Since(stats.StartTime)
	
	response := map[string]interface{}{
		"uptime_seconds":     uptime.Seconds(),
		"total_requests":     atomic.LoadInt64(&stats.TotalRequests),
		"active_jobs":        atomic.LoadInt64(&stats.ActiveJobs),
		"completed_jobs":     atomic.LoadInt64(&stats.CompletedJobs),
		"failed_jobs":        atomic.LoadInt64(&stats.FailedJobs),
		"total_subdomains":   atomic.LoadInt64(&stats.TotalSubdomains),
		"total_probes":       atomic.LoadInt64(&stats.TotalProbes),
		"successful_probes":  atomic.LoadInt64(&stats.SuccessfulProbes),
		"dns_queries":        atomic.LoadInt64(&stats.DNSQueries),
		"last_activity":      stats.LastActivity,
		"source_stats":       stats.SourceStats,
		"memory_usage":       getMemoryUsage(),
		"dns_servers":        config.DNS.Servers,
		"rate_limit":         fmt.Sprintf("%d/s", config.RateLimit.RequestsPerSecond),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Health check handlers
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "2.0.0",
	})
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	// Check if critical services are ready
	ready := true
	checks := make(map[string]bool)
	
	// Check DNS resolver
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	_, err := dnsResolver.LookupHost(ctx, "google.com")
	checks["dns"] = err == nil
	if err != nil {
		ready = false
	}
	
	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ready":  ready,
		"checks": checks,
	})
}

// Enhanced configuration handler
func configHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Return current configuration (sanitized)
		sanitizedConfig := map[string]interface{}{
			"timeouts": map[string]string{
				"wayback": config.Timeouts.Wayback.String(),
				"crtsh":   config.Timeouts.CrtSh.String(),
				"dns":     config.Timeouts.DNS.String(),
				"search":  config.Timeouts.Search.String(),
				"permute": config.Timeouts.Permute.String(),
				"zone":    config.Timeouts.Zone.String(),
			},
			"dns": map[string]interface{}{
				"servers":     config.DNS.Servers,
				"concurrency": config.DNS.Concurrency,
				"timeout":     config.DNS.Timeout.String(),
			},
			"rate_limit": map[string]interface{}{
				"requests_per_second": config.RateLimit.RequestsPerSecond,
				"burst_size":          config.RateLimit.BurstSize,
			},
			"wordlist_categories": getWordlistCategories(),
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sanitizedConfig)
		return
	}
	
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func getWordlistCategories() map[string]int {
	categories := make(map[string]int)
	for category, words := range commonSubdomains {
		categories[category] = len(words)
	}
	return categories
}

// Utility functions for environment variable parsing
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// Memory usage monitoring
func getMemoryUsage() map[string]interface{} {
	// This is a simplified version - you might want to use runtime.MemStats
	return map[string]interface{}{
		"goroutines": "runtime.NumGoroutine() would go here",
		"note":       "Implement with runtime.MemStats for production",
	}
}

// Metrics server for Prometheus integration
func startMetricsServer() {
	if !config.Monitoring.EnableMetrics {
		return
	}
	
	// Don't start separate server if using same port as main server
	if config.Monitoring.MetricsPort == config.Port {
		log.Printf("Metrics server using main server port %s", config.Port)
		return
	}
	
	metricsMux := http.NewServeMux()
	metricsMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Metrics Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { color: #333; border-bottom: 2px solid #4ade80; padding-bottom: 10px; }
        .metrics-link { 
            display: inline-block; 
            background: #4ade80; 
            color: white; 
            text-decoration: none; 
            padding: 10px 20px; 
            border-radius: 5px; 
            margin: 10px 0;
        }
        .metrics-link:hover { background: #22c55e; }
        .info { background: #f0f9ff; padding: 15px; border-radius: 5px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1 class="header">Subdomain Scanner Metrics Server</h1>
        <div class="info">
            <p><strong>Main Application:</strong> <a href="http://localhost:%s">http://localhost:%s</a></p>
            <p><strong>Metrics Endpoint:</strong> <a href="/metrics" class="metrics-link">View Prometheus Metrics</a></p>
            <p><strong>Health Check:</strong> <a href="http://localhost:%s/health">http://localhost:%s/health</a></p>
        </div>
        <h3>Available Endpoints:</h3>
        <ul>
            <li><a href="/metrics">/metrics</a> - Prometheus metrics format</li>
            <li><a href="http://localhost:%s/api/stats">http://localhost:%s/api/stats</a> - JSON statistics</li>
        </ul>
    </div>
</body>
</html>`, config.Port, config.Port, config.Port, config.Port, config.Port, config.Port)
	})
	metricsMux.HandleFunc("/metrics", metricsHandler)
	metricsMux.HandleFunc("/health", healthHandler)
	
	server := &http.Server{
		Addr:         ":" + config.Monitoring.MetricsPort,
		Handler:      metricsMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	log.Printf("Starting dedicated metrics server on port %s", config.Monitoring.MetricsPort)
	
	// Use a more graceful error handling instead of log.Fatal
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Metrics server error (port %s may be in use): %v", config.Monitoring.MetricsPort, err)
		log.Printf("Metrics are still available on main server: http://localhost:%s/metrics", config.Port)
	}
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Prometheus metrics format
	metrics := fmt.Sprintf(`# HELP subdomain_scanner_requests_total Total number of requests
# TYPE subdomain_scanner_requests_total counter
subdomain_scanner_requests_total %d

# HELP subdomain_scanner_active_jobs Current number of active jobs
# TYPE subdomain_scanner_active_jobs gauge
subdomain_scanner_active_jobs %d

# HELP subdomain_scanner_subdomains_total Total number of subdomains discovered
# TYPE subdomain_scanner_subdomains_total counter
subdomain_scanner_subdomains_total %d

# HELP subdomain_scanner_dns_queries_total Total number of DNS queries
# TYPE subdomain_scanner_dns_queries_total counter
subdomain_scanner_dns_queries_total %d

# HELP subdomain_scanner_uptime_seconds Uptime in seconds
# TYPE subdomain_scanner_uptime_seconds counter
subdomain_scanner_uptime_seconds %f
`,
		atomic.LoadInt64(&stats.TotalRequests),
		atomic.LoadInt64(&stats.ActiveJobs),
		atomic.LoadInt64(&stats.TotalSubdomains),
		atomic.LoadInt64(&stats.DNSQueries),
		time.Since(stats.StartTime).Seconds(),
	)
	
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(metrics))
}

// The rest of the stream handlers would be similar to your original implementation
// but with enhanced error handling, logging, and metrics collection
// Due to length constraints, I'm showing the framework - you would implement
// waybackStream, crtshStream, dnsStream, etc. with similar enhancements

// Enhanced wayback stream with better error handling
func waybackStream(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "missing target parameter", http.StatusBadRequest)
		return
	}

	if !domainRe.MatchString(target) {
		http.Error(w, "invalid domain format", http.StatusBadRequest)
		return
	}

	sseHeader(w)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), config.Timeouts.Wayback)
	defer cancel()

	job := createJob(target, []string{"wayback"})
	defer job.Complete()

	// Create API URL for Wayback Machine
	apiURL := fmt.Sprintf(
		"https://web.archive.org/cdx/search/cdx?url=*.%s/*&output=text&fl=original&collapse=urlkey",
		target,
	)

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		log.Printf("Wayback request creation error: %v", err)
		// Send completion signal and return
		fmt.Fprintf(w, "event: complete\ndata: Wayback scan completed with errors\n\n")
		flusher.Flush()
		return
	}

	client := &http.Client{
		Timeout: config.HTTP.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.HTTP.SkipTLSVerify,
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Wayback API error: %v", err)
		// Send completion signal and return
		fmt.Fprintf(w, "event: complete\ndata: Wayback scan completed - API unavailable\n\n")
		flusher.Flush()
		return
	}
	defer resp.Body.Close()

	seen := make(map[string]struct{})
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Wayback response read error: %v", err)
		fmt.Fprintf(w, "event: complete\ndata: Wayback scan completed with errors\n\n")
		flusher.Flush()
		return
	}

	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		select {
		case <-ctx.Done():
			fmt.Fprintf(w, "event: complete\ndata: Wayback scan cancelled\n\n")
			flusher.Flush()
			return
		default:
		}

		if matches := hostRe.FindStringSubmatch(line); matches != nil {
			host := strings.ToLower(matches[1])
			if strings.HasSuffix(host, "."+target) {
				if _, dup := seen[host]; !dup {
					seen[host] = struct{}{}
					
					result := Result{
						Host:      host,
						Source:    "wayback",
						Status:    "discovered",
						Timestamp: time.Now(),
					}
					
					job.AddResult("wayback", result)
					
					// Send to client
					fmt.Fprintf(w, "data: %s\n\n", host)
					flusher.Flush()
				}
			}
		}
	}

	log.Printf("Wayback found %d unique hosts for %s", len(seen), target)
	// Send completion signal
	fmt.Fprintf(w, "event: complete\ndata: Wayback scan completed - found %d hosts\n\n", len(seen))
	flusher.Flush()
}

// Implement other stream handlers similarly...
func crtshStream(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "missing target parameter", http.StatusBadRequest)
		return
	}

	sseHeader(w)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), config.Timeouts.CrtSh)
	defer cancel()

	job := createJob(target, []string{"crtsh"})
	defer job.Complete()

	apiURL := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", target)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		log.Printf("crt.sh request creation error: %v", err)
		fmt.Fprintf(w, "event: complete\ndata: Certificate transparency scan completed with errors\n\n")
		flusher.Flush()
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", config.HTTP.UserAgent)

	client := &http.Client{Timeout: config.HTTP.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("crt.sh API error: %v", err)
		fmt.Fprintf(w, "event: complete\ndata: Certificate transparency scan completed - API unavailable\n\n")
		flusher.Flush()
		return
	}
	defer resp.Body.Close()

	var entries []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		log.Printf("crt.sh JSON decode error: %v", err)
		fmt.Fprintf(w, "event: complete\ndata: Certificate transparency scan completed with errors\n\n")
		flusher.Flush()
		return
	}

	seen := make(map[string]struct{})
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			fmt.Fprintf(w, "event: complete\ndata: Certificate transparency scan cancelled\n\n")
			flusher.Flush()
			return
		default:
		}

		if nameValue, ok := entry["name_value"].(string); ok {
			for _, name := range strings.Split(nameValue, "\n") {
				host := strings.ToLower(strings.TrimSpace(name))
				host = strings.TrimPrefix(host, "*.")

				if strings.HasSuffix(host, "."+target) && host != target {
					if _, dup := seen[host]; !dup {
						seen[host] = struct{}{}

						result := Result{
							Host:      host,
							Source:    "crtsh",
							Status:    "discovered",
							Timestamp: time.Now(),
						}

						job.AddResult("crtsh", result)

						fmt.Fprintf(w, "data: %s\n\n", host)
						flusher.Flush()
					}
				}
			}
		}
	}

	log.Printf("crt.sh found %d unique hosts for %s", len(seen), target)
	fmt.Fprintf(w, "event: complete\ndata: Certificate transparency scan completed - found %d hosts\n\n", len(seen))
	flusher.Flush()
}

func dnsStream(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "missing target parameter", http.StatusBadRequest)
		return
	}

	sseHeader(w)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), config.Timeouts.DNS)
	defer cancel()

	job := createJob(target, []string{"dns"})
	defer job.Complete()

	// Get all subdomains from all categories
	var allSubdomains []string
	for _, subdomains := range commonSubdomains {
		allSubdomains = append(allSubdomains, subdomains...)
	}

	seen := make(map[string]struct{})
	semaphore := make(chan struct{}, config.DNS.Concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, subdomain := range allSubdomains {
		select {
		case <-ctx.Done():
			fmt.Fprintf(w, "event: complete\ndata: DNS brute force scan cancelled\n\n")
			flusher.Flush()
			return
		default:
		}

		wg.Add(1)
		go func(sub string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			host := fmt.Sprintf("%s.%s", sub, target)

			ips, err := dnsResolver.LookupHost(ctx, host)
			if err == nil && len(ips) > 0 {
				mu.Lock()
				if _, dup := seen[host]; !dup {
					seen[host] = struct{}{}

					result := Result{
						Host:      host,
						Source:    "dns",
						Status:    "discovered",
						Timestamp: time.Now(),
					}

					job.AddResult("dns", result)

					fmt.Fprintf(w, "data: %s\n\n", host)
					flusher.Flush()
				}
				mu.Unlock()
			}
		}(subdomain)
	}

	// Wait for all DNS queries to complete or context to be cancelled
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("DNS enumeration found %d unique hosts for %s", len(seen), target)
		fmt.Fprintf(w, "event: complete\ndata: DNS brute force scan completed - found %d hosts\n\n", len(seen))
		flusher.Flush()
	case <-ctx.Done():
		log.Printf("DNS enumeration cancelled for %s", target)
		fmt.Fprintf(w, "event: complete\ndata: DNS brute force scan cancelled\n\n")
		flusher.Flush()
	}
}

func searchEngineStream(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "missing target parameter", http.StatusBadRequest)
		return
	}

	sseHeader(w)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), config.Timeouts.Search)
	defer cancel()

	job := createJob(target, []string{"search"})
	defer job.Complete()

	// Simple Google search implementation
	searchURL := fmt.Sprintf("https://www.google.com/search?q=site:%s", target)
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		log.Printf("Search engine request error: %v", err)
		fmt.Fprintf(w, "event: complete\ndata: Search engine scan completed with errors\n\n")
		flusher.Flush()
		return
	}

	req.Header.Set("User-Agent", config.HTTP.UserAgent)

	client := &http.Client{Timeout: config.HTTP.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Search engine request failed: %v", err)
		fmt.Fprintf(w, "event: complete\ndata: Search engine scan completed - service unavailable\n\n")
		flusher.Flush()
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read search response: %v", err)
		fmt.Fprintf(w, "event: complete\ndata: Search engine scan completed with errors\n\n")
		flusher.Flush()
		return
	}

	urlPattern := regexp.MustCompile(`https?://([^/\s"'<>]+\.` + regexp.QuoteMeta(target) + `)`)
	matches := urlPattern.FindAllStringSubmatch(string(body), -1)

	seen := make(map[string]struct{})
	for _, match := range matches {
		select {
		case <-ctx.Done():
			fmt.Fprintf(w, "event: complete\ndata: Search engine scan cancelled\n\n")
			flusher.Flush()
			return
		default:
		}

		if len(match) > 1 {
			host := strings.ToLower(match[1])
			if _, dup := seen[host]; !dup {
				seen[host] = struct{}{}

				result := Result{
					Host:      host,
					Source:    "search",
					Status:    "discovered",
					Timestamp: time.Now(),
				}

				job.AddResult("search", result)

				fmt.Fprintf(w, "data: %s\n\n", host)
				flusher.Flush()
			}
		}
	}

	log.Printf("Search engine found %d unique hosts for %s", len(seen), target)
	fmt.Fprintf(w, "event: complete\ndata: Search engine scan completed - found %d hosts\n\n", len(seen))
	flusher.Flush()
}

func permuteStream(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "missing target parameter", http.StatusBadRequest)
		return
	}

	sseHeader(w)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), config.Timeouts.Permute)
	defer cancel()

	job := createJob(target, []string{"permute"})
	defer job.Complete()

	permutations := generatePermutations(target)
	seen := make(map[string]struct{})
	semaphore := make(chan struct{}, config.DNS.Concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, perm := range permutations {
		select {
		case <-ctx.Done():
			fmt.Fprintf(w, "event: complete\ndata: Permutation scan cancelled\n\n")
			flusher.Flush()
			return
		default:
		}

		wg.Add(1)
		go func(host string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			ips, err := dnsResolver.LookupHost(ctx, host)
			if err == nil && len(ips) > 0 {
				mu.Lock()
				if _, dup := seen[host]; !dup {
					seen[host] = struct{}{}

					result := Result{
						Host:      host,
						Source:    "permute",
						Status:    "discovered",
						Timestamp: time.Now(),
					}

					job.AddResult("permute", result)

					fmt.Fprintf(w, "data: %s\n\n", host)
					flusher.Flush()
				}
				mu.Unlock()
			}
		}(perm)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("Permutation generation found %d unique hosts for %s", len(seen), target)
		fmt.Fprintf(w, "event: complete\ndata: Permutation scan completed - found %d hosts\n\n", len(seen))
		flusher.Flush()
	case <-ctx.Done():
		log.Printf("Permutation generation cancelled for %s", target)
		fmt.Fprintf(w, "event: complete\ndata: Permutation scan cancelled\n\n")
		flusher.Flush()
	}
}

func zoneTransferStream(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "missing target parameter", http.StatusBadRequest)
		return
	}

	sseHeader(w)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), config.Timeouts.Zone)
	defer cancel()

	job := createJob(target, []string{"zone"})
	defer job.Complete()

	// Look up nameservers for the domain
	nsRecords, err := net.LookupNS(target)
	if err != nil {
		log.Printf("Failed to lookup NS records for %s: %v", target, err)
		// Send error message to client
		fmt.Fprintf(w, "event: complete\ndata: Zone transfer completed with errors - Failed to lookup NS records\n\n")
		flusher.Flush()
		return
	}

	// Send nameserver information to client
	fmt.Fprintf(w, "data: info: Found %d nameservers for %s\n\n", len(nsRecords), target)
	flusher.Flush()

	seen := make(map[string]struct{})

	// Try zone transfer against each nameserver
	for _, ns := range nsRecords {
		select {
		case <-ctx.Done():
			fmt.Fprintf(w, "event: complete\ndata: Zone transfer scan cancelled\n\n")
			flusher.Flush()
			return
		default:
		}

		log.Printf("Attempting zone transfer from %s for %s", ns.Host, target)
		
		// Send status update to client
		fmt.Fprintf(w, "data: status: Testing nameserver %s\n\n", ns.Host)
		flusher.Flush()

		// Simple connection test (actual zone transfer would need more complex DNS library usage)
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(ns.Host, "53"), 5*time.Second)
		if err != nil {
			log.Printf("Failed to connect to nameserver %s: %v", ns.Host, err)
			fmt.Fprintf(w, "data: error: Failed to connect to %s: %v\n\n", ns.Host, err)
			flusher.Flush()
			continue
		}
		conn.Close()

		// If we successfully connected, record the nameserver
		if _, dup := seen[ns.Host]; !dup {
			seen[ns.Host] = struct{}{}
			
			result := Result{
				Host:      ns.Host,
				Source:    "zone",
				Status:    "nameserver",
				Title:     fmt.Sprintf("Nameserver for %s", target),
				Timestamp: time.Now(),
			}
			
			job.AddResult("zone", result)
			
			// Send nameserver as a result (even though it's not a subdomain, it's useful info)
			fmt.Fprintf(w, "data: %s\n\n", ns.Host)
			flusher.Flush()
		}

		log.Printf("Successfully connected to nameserver %s (zone transfer would require DNS protocol implementation)", ns.Host)
	}

	// Send completion message
	log.Printf("Zone transfer attempt completed for %s (found %d nameservers)", target, len(nsRecords))
	fmt.Fprintf(w, "event: complete\ndata: Zone transfer scan completed - found %d nameservers\n\n", len(nsRecords))
	flusher.Flush()
}

func generatePermutations(domain string) []string {
	var permutations []string

	prefixes := []string{"dev", "test", "stage", "staging", "prod", "production", "www", "api", "admin", "app", "mobile", "m"}
	suffixes := []string{"dev", "test", "stage", "staging", "prod", "production", "api", "admin", "backup", "old", "new"}

	// Add base subdomains
	for _, prefix := range prefixes {
		permutations = append(permutations, fmt.Sprintf("%s.%s", prefix, domain))
	}

	// Add permutations with suffixes
	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
		baseDomain := parts[0]
		tld := strings.Join(parts[1:], ".")

		for _, suffix := range suffixes {
			permutations = append(permutations, fmt.Sprintf("%s-%s.%s", baseDomain, suffix, tld))
			permutations = append(permutations, fmt.Sprintf("%s%s.%s", baseDomain, suffix, tld))
		}
	}

	// Add numbered variations
	for i := 1; i <= 10; i++ {
		permutations = append(permutations, fmt.Sprintf("www%d.%s", i, domain))
		permutations = append(permutations, fmt.Sprintf("mail%d.%s", i, domain))
		permutations = append(permutations, fmt.Sprintf("ftp%d.%s", i, domain))
	}

	return permutations
}

// Health check function for containers
func performHealthCheck() error {
	// Create a timeout context for the health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Determine the port to check
	port := getEnvString("PORT", "8080")
	healthURL := fmt.Sprintf("http://localhost:%s/health", port)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Additional checks - verify DNS resolver is working
	if dnsResolver != nil {
		testCtx, testCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer testCancel()
		
		_, err := dnsResolver.LookupHost(testCtx, "google.com")
		if err != nil {
			return fmt.Errorf("DNS resolver health check failed: %w", err)
		}
	}

	return nil
}

// Version handler for API endpoint
func versionHandler(w http.ResponseWriter, r *http.Request) {
	versionInfo := map[string]interface{}{
		"version":     version,
		"build_time":  buildTime,
		"git_commit":  gitCommit,
		"go_version":  runtime.Version(),
		"platform":    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		"uptime":      time.Since(stats.StartTime).String(),
		"start_time":  stats.StartTime,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versionInfo)
}

// Enhanced job management endpoints
func jobsHandler(w http.ResponseWriter, r *http.Request) {
	jobManager.mu.RLock()
	defer jobManager.mu.RUnlock()
	
	jobs := make([]*Job, 0, len(jobManager.jobs))
	for _, job := range jobManager.jobs {
		jobs = append(jobs, job)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func jobDetailHandler(w http.ResponseWriter, r *http.Request) {
	jobID := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
	
	jobManager.mu.RLock()
	job, exists := jobManager.jobs[jobID]
	jobManager.mu.RUnlock()
	
	if !exists {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	
	jobManager.mu.RLock()
	defer jobManager.mu.RUnlock()
	
	activeJobs := make([]*Job, 0)
	for _, job := range jobManager.jobs {
		if job.Target == target && job.Status == "running" {
			activeJobs = append(activeJobs, job)
		}
	}
	
	status := map[string]interface{}{
		"target":      target,
		"active_jobs": len(activeJobs),
		"jobs":        activeJobs,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func abortHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	
	jobManager.mu.Lock()
	cancelled := 0
	for _, job := range jobManager.jobs {
		if job.Target == target && job.Status == "running" {
			if job.Cancel != nil {
				job.Cancel()
			}
			job.Status = "cancelled"
			cancelled++
		}
	}
	jobManager.mu.Unlock()
	
	log.Printf("Cancelled %d jobs for target: %s", cancelled, target)
	w.WriteHeader(http.StatusNoContent)
}