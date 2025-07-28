package main

import (
    "bufio"
    "context"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net"
    "net/http"
    "net/url"
    "os"
    "regexp"
    "strings"
    "sync"
    "time"
)

var (
    hostRe  = regexp.MustCompile(`https?://([^/]+)`)
    titleRe = regexp.MustCompile(`(?is)<title>(.*?)</title>`)

    // Map of running jobs, protected by a mutex
    jobs   = map[string]context.CancelFunc{}
    jobsMu sync.Mutex
    
    // Common subdomains for permutation generation
    commonSubdomains = []string{
        "www", "mail", "ftp", "localhost", "webmail", "smtp", "pop", "ns1", "webdisk", "ns2",
        "cpanel", "whm", "autodiscover", "autoconfig", "m", "imap", "test", "ns", "blog",
        "pop3", "dev", "www2", "admin", "forum", "news", "vpn", "ns3", "mail2", "new",
        "mysql", "old", "www1", "email", "img", "www3", "help", "shop", "sql", "secure",
        "beta", "pic", "mail3", "share", "web", "api", "img1", "www4", "www5", "admin2",
        "admins", "administrator", "email2", "asp", "backup", "sec", "mx", "static", "www6",
        "upload", "support", "www7", "ex", "www8", "web1", "www9", "www10", "mail4", "mx1",
        "lab", "file", "git", "app", "apps", "stage", "staging", "prod", "production",
        "demo", "docs", "portal", "status", "cdn", "assets", "media", "images", "js",
        "css", "wiki", "chat", "live", "video", "store", "download", "downloads", "update",
        "updates", "api1", "api2", "v1", "v2", "mobile", "m1", "m2", "analytics", "stats",
        "dashboard", "panel", "control", "manage", "manager", "login", "signin", "sso",
    }
    
    // Search engines patterns
    searchEngines = map[string]string{
        "google":    "site:%s",
        "bing":      "site:%s",
        "yahoo":     "site:%s",
        "duckduckgo": "site:%s",
    }
)

type crtEntry struct {
    NameValue string `json:"name_value"`
}

type result struct {
    Host   string
    Tried  string
    Status string
    Title  string
    Err    string
    Source string
}

type probeResponse struct {
    Status string `json:"status"`
    Title  string `json:"title"`
    Error  string `json:"error"`
}

type scanStatus struct {
    Active    bool                `json:"active"`
    Sources   map[string]bool     `json:"sources"`
    Stats     map[string]int      `json:"stats"`
    StartTime time.Time           `json:"start_time"`
}

func main() {
    mux := http.NewServeMux()

    // Serve static UI
    mux.Handle("/", http.FileServer(http.Dir("./public/")))

    // SSE streams  
    mux.HandleFunc("/api/wayback/stream", waybackStream)
    mux.HandleFunc("/api/crtsh/stream", crtshStream)
    mux.HandleFunc("/api/dns/stream", dnsStream)
    mux.HandleFunc("/api/search/stream", searchEngineStream)
    mux.HandleFunc("/api/permute/stream", permuteStream)
    mux.HandleFunc("/api/zone/stream", zoneTransferStream)

    // Probe handler
    mux.HandleFunc("/api/probe", probeHandler)

    // Control handlers
    mux.HandleFunc("/api/abort", abortHandler)
    mux.HandleFunc("/api/status", statusHandler)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Enhanced Subdomain Enumeration Tool listening on port %s…", port)
    log.Fatal(http.ListenAndServe(":"+port, mux))
}

func sseHeader(w http.ResponseWriter) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
    target := r.URL.Query().Get("target")
    
    jobsMu.Lock()
    defer jobsMu.Unlock()
    
    status := scanStatus{
        Active:  false,
        Sources: make(map[string]bool),
        Stats:   make(map[string]int),
    }
    
    // Check which sources are active for this target
    sources := []string{"wayback", "crtsh", "dns", "search", "permute", "zone"}
    for _, source := range sources {
        jobKey := target + "_" + source
        if _, exists := jobs[jobKey]; exists {
            status.Active = true
            status.Sources[source] = true
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(status)
}

func probeHandler(w http.ResponseWriter, r *http.Request) {
    targetURL := r.URL.Query().Get("url")
    if targetURL == "" {
        http.Error(w, "missing url parameter", http.StatusBadRequest)
        return
    }

    _, err := url.Parse(targetURL)
    if err != nil {
        writeProbeError(w, "invalid URL", err)
        return
    }

    client := &http.Client{
        Timeout: 10 * time.Second,
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true,
            },
        },
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            if len(via) >= 3 {
                return fmt.Errorf("too many redirects")
            }
            return nil
        },
    }

    req, err := http.NewRequest("GET", targetURL, nil)
    if err != nil {
        writeProbeError(w, "failed to create request", err)
        return
    }

    req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; SubdomainScanner/2.0)")
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

    resp, err := client.Do(req)
    if err != nil {
        writeProbeError(w, "connection failed", err)
        return
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
    if err != nil {
        writeProbeError(w, "failed to read response", err)
        return
    }

    title := "No title"
    if matches := titleRe.FindStringSubmatch(string(body)); len(matches) > 1 {
        title = strings.TrimSpace(matches[1])
        title = strings.ReplaceAll(title, "\n", " ")
        title = strings.ReplaceAll(title, "\r", " ")
        title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")
        if len(title) > 100 {
            title = title[:100] + "..."
        }
    }

    response := probeResponse{
        Status: fmt.Sprintf("%d", resp.StatusCode),
        Title:  title,
        Error:  "",
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func writeProbeError(w http.ResponseWriter, message string, err error) {
    response := probeResponse{
        Status: "0",
        Title:  "Connection failed",
        Error:  fmt.Sprintf("%s: %v", message, err),
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func abortHandler(w http.ResponseWriter, r *http.Request) {
    target := r.URL.Query().Get("target")
    jobsMu.Lock()
    cancelled := 0
    for key, cancel := range jobs {
        if strings.Contains(key, target) {
            cancel()
            delete(jobs, key)
            cancelled++
        }
    }
    jobsMu.Unlock()
    
    if cancelled > 0 {
        log.Printf("Cancelled %d jobs for target: %s", cancelled, target)
    }
    
    w.WriteHeader(http.StatusNoContent)
}

func waybackStream(w http.ResponseWriter, r *http.Request) {
    target := r.URL.Query().Get("target")
    if target == "" {
        http.Error(w, "missing target", http.StatusBadRequest)
        return
    }
    
    sseHeader(w)
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "streaming not supported", http.StatusInternalServerError)
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
    defer cancel()
    
    jobKey := target + "_wayback"
    
    jobsMu.Lock()
    if existingCancel, exists := jobs[jobKey]; exists {
        existingCancel()
    }
    jobs[jobKey] = cancel
    jobsMu.Unlock()

    defer func() {
        jobsMu.Lock()
        delete(jobs, jobKey)
        jobsMu.Unlock()
        log.Printf("Wayback stream ended for %s", target)
    }()

    resultCh := make(chan string, 100)
    
    go func() {
        defer close(resultCh)
        
        apiURL := fmt.Sprintf(
            "https://web.archive.org/cdx/search/cdx?url=*.%s/*&output=text&fl=original&collapse=urlkey",
            target,
        )
        req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
        client := &http.Client{Timeout: 30 * time.Second}
        resp, err := client.Do(req)
        if err != nil {
            log.Printf("Wayback API error: %v", err)
            return
        }
        defer resp.Body.Close()

        scanner := bufio.NewScanner(resp.Body)
        seen := map[string]struct{}{}
        
        for scanner.Scan() {
            select {
            case <-ctx.Done():
                return
            default:
            }
            
            line := scanner.Text()
            if m := hostRe.FindStringSubmatch(line); m != nil {
                host := strings.ToLower(m[1])
                if strings.HasSuffix(host, "."+target) {
                    if _, dup := seen[host]; !dup {
                        seen[host] = struct{}{}
                        select {
                        case resultCh <- host:
                        case <-ctx.Done():
                            return
                        }
                    }
                }
            }
        }
        
        log.Printf("Wayback found %d unique hosts for %s", len(seen), target)
    }()

    for host := range resultCh {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        fmt.Fprintf(w, "data: %s\n\n", host)
        flusher.Flush()
    }
}

func crtshStream(w http.ResponseWriter, r *http.Request) {
    target := r.URL.Query().Get("target")
    if target == "" {
        http.Error(w, "missing target", http.StatusBadRequest)
        return
    }
    
    sseHeader(w)
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "streaming not supported", http.StatusInternalServerError)
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
    defer cancel()
    
    jobKey := target + "_crtsh"
    
    jobsMu.Lock()
    if existingCancel, exists := jobs[jobKey]; exists {
        existingCancel()
    }
    jobs[jobKey] = cancel
    jobsMu.Unlock()

    defer func() {
        jobsMu.Lock()
        delete(jobs, jobKey)
        jobsMu.Unlock()
        log.Printf("crt.sh stream ended for %s", target)
    }()

    resultCh := make(chan string, 100)
    
    go func() {
        defer close(resultCh)
        
        apiURL := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", target)
        req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
        req.Header.Set("Accept", "application/json")
        req.Header.Set("User-Agent", "curl/7.64.1")

        client := &http.Client{Timeout: 60 * time.Second}
        resp, err := client.Do(req)
        if err != nil {
            log.Printf("crt.sh API error: %v", err)
            return
        }
        defer resp.Body.Close()

        contentType := resp.Header.Get("Content-Type")
        if !strings.Contains(contentType, "json") {
            log.Printf("crt.sh returned non-JSON content type: %s", contentType)
            return
        }

        var entries []crtEntry
        if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
            log.Printf("crt.sh JSON decode error: %v", err)
            return
        }

        seen := map[string]struct{}{}
        
        for _, e := range entries {
            select {
            case <-ctx.Done():
                return
            default:
            }
            
            for _, name := range strings.Split(e.NameValue, "\n") {
                host := strings.ToLower(strings.TrimSpace(name))
                host = strings.TrimPrefix(host, "*.")
                
                if strings.HasSuffix(host, "."+target) && host != target {
                    if _, dup := seen[host]; !dup {
                        seen[host] = struct{}{}
                        select {
                        case resultCh <- host:
                        case <-ctx.Done():
                            return
                        }
                    }
                }
            }
        }
        
        log.Printf("crt.sh found %d unique hosts for %s", len(seen), target)
    }()

    for host := range resultCh {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        fmt.Fprintf(w, "data: %s\n\n", host)
        flusher.Flush()
    }
}

func dnsStream(w http.ResponseWriter, r *http.Request) {
    target := r.URL.Query().Get("target")
    if target == "" {
        http.Error(w, "missing target", http.StatusBadRequest)
        return
    }
    
    sseHeader(w)
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "streaming not supported", http.StatusInternalServerError)
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
    defer cancel()
    
    jobKey := target + "_dns"
    
    jobsMu.Lock()
    if existingCancel, exists := jobs[jobKey]; exists {
        existingCancel()
    }
    jobs[jobKey] = cancel
    jobsMu.Unlock()

    defer func() {
        jobsMu.Lock()
        delete(jobs, jobKey)
        jobsMu.Unlock()
        log.Printf("DNS enumeration ended for %s", target)
    }()

    resultCh := make(chan string, 100)
    
    go func() {
        defer close(resultCh)
        
        // DNS brute force with common subdomains
        seen := map[string]struct{}{}
        semaphore := make(chan struct{}, 50) // Limit concurrent DNS queries
        
        var wg sync.WaitGroup
        
        for _, subdomain := range commonSubdomains {
            select {
            case <-ctx.Done():
                return
            default:
            }
            
            wg.Add(1)
            go func(sub string) {
                defer wg.Done()
                
                semaphore <- struct{}{} // Acquire
                defer func() { <-semaphore }() // Release
                
                host := fmt.Sprintf("%s.%s", sub, target)
                
                // Try DNS resolution
                resolver := &net.Resolver{
                    PreferGo: true,
                    Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
                        d := net.Dialer{
                            Timeout: time.Second * 2,
                        }
                        return d.DialContext(ctx, network, "8.8.8.8:53")
                    },
                }
                
                ips, err := resolver.LookupIPAddr(ctx, host)
                if err == nil && len(ips) > 0 {
                    if _, dup := seen[host]; !dup {
                        seen[host] = struct{}{}
                        select {
                        case resultCh <- host:
                        case <-ctx.Done():
                            return
                        }
                    }
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
        case <-ctx.Done():
            log.Printf("DNS enumeration cancelled for %s", target)
        }
    }()

    for host := range resultCh {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        fmt.Fprintf(w, "data: %s\n\n", host)
        flusher.Flush()
    }
}

func searchEngineStream(w http.ResponseWriter, r *http.Request) {
    target := r.URL.Query().Get("target")
    if target == "" {
        http.Error(w, "missing target", http.StatusBadRequest)
        return
    }
    
    sseHeader(w)
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "streaming not supported", http.StatusInternalServerError)
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
    defer cancel()
    
    jobKey := target + "_search"
    
    jobsMu.Lock()
    if existingCancel, exists := jobs[jobKey]; exists {
        existingCancel()
    }
    jobs[jobKey] = cancel
    jobsMu.Unlock()

    defer func() {
        jobsMu.Lock()
        delete(jobs, jobKey)
        jobsMu.Unlock()
        log.Printf("Search engine enumeration ended for %s", target)
    }()

    resultCh := make(chan string, 100)
    
    go func() {
        defer close(resultCh)
        
        seen := map[string]struct{}{}
        
        // Simple Google search scraping (basic implementation)
        // Note: This is a simplified version - in production you'd want more sophisticated scraping
        searchURL := fmt.Sprintf("https://www.google.com/search?q=site:%s", target)
        
        client := &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                TLSClientConfig: &tls.Config{
                    InsecureSkipVerify: true,
                },
            },
        }
        
        req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
        if err != nil {
            log.Printf("Search engine request error: %v", err)
            return
        }
        
        req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
        
        resp, err := client.Do(req)
        if err != nil {
            log.Printf("Search engine request failed: %v", err)
            return
        }
        defer resp.Body.Close()
        
        body, err := io.ReadAll(resp.Body)
        if err != nil {
            log.Printf("Failed to read search response: %v", err)
            return
        }
        
        // Extract domains from search results using regex
        urlPattern := regexp.MustCompile(`https?://([^/\s"'<>]+\.` + regexp.QuoteMeta(target) + `)`)
        matches := urlPattern.FindAllStringSubmatch(string(body), -1)
        
        for _, match := range matches {
            if len(match) > 1 {
                host := strings.ToLower(match[1])
                if _, dup := seen[host]; !dup {
                    seen[host] = struct{}{}
                    select {
                    case resultCh <- host:
                    case <-ctx.Done():
                        return
                    }
                }
            }
        }
        
        log.Printf("Search engine found %d unique hosts for %s", len(seen), target)
    }()

    for host := range resultCh {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        fmt.Fprintf(w, "data: %s\n\n", host)
        flusher.Flush()
    }
}

func permuteStream(w http.ResponseWriter, r *http.Request) {
    target := r.URL.Query().Get("target")
    if target == "" {
        http.Error(w, "missing target", http.StatusBadRequest)
        return
    }
    
    sseHeader(w)
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "streaming not supported", http.StatusInternalServerError)
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
    defer cancel()
    
    jobKey := target + "_permute"
    
    jobsMu.Lock()
    if existingCancel, exists := jobs[jobKey]; exists {
        existingCancel()
    }
    jobs[jobKey] = cancel
    jobsMu.Unlock()

    defer func() {
        jobsMu.Lock()
        delete(jobs, jobKey)
        jobsMu.Unlock()
        log.Printf("Permutation generation ended for %s", target)
    }()

    resultCh := make(chan string, 100)
    
    go func() {
        defer close(resultCh)
        
        seen := map[string]struct{}{}
        var seenMu sync.Mutex // Mutex to protect seen map
        semaphore := make(chan struct{}, 100) // Limit concurrent checks
        
        var wg sync.WaitGroup
        
        // Generate permutations and check if they resolve
        permutations := generatePermutations(target)
        
        for _, perm := range permutations {
            select {
            case <-ctx.Done():
                return
            default:
            }
            
            wg.Add(1)
            go func(host string) {
                defer wg.Done()
                
                semaphore <- struct{}{} // Acquire
                defer func() { <-semaphore }() // Release
                
                // Quick DNS check
                resolver := &net.Resolver{
                    PreferGo: true,
                    Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
                        d := net.Dialer{
                            Timeout: time.Second * 2,
                        }
                        return d.DialContext(ctx, network, "8.8.8.8:53")
                    },
                }
                
                ips, err := resolver.LookupIPAddr(ctx, host)
                if err == nil && len(ips) > 0 {
                    seenMu.Lock()
                    if _, dup := seen[host]; !dup {
                        seen[host] = struct{}{}
                        seenMu.Unlock()
                        select {
                        case resultCh <- host:
                        case <-ctx.Done():
                            return
                        }
                    } else {
                        seenMu.Unlock()
                    }
                }
            }(perm)
        }
        
        // Wait for all checks to complete or context to be cancelled
        done := make(chan struct{})
        go func() {
            wg.Wait()
            close(done)
        }()
        
        select {
        case <-done:
            seenMu.Lock()
            count := len(seen)
            seenMu.Unlock()
            log.Printf("Permutation generation found %d unique hosts for %s", count, target)
        case <-ctx.Done():
            log.Printf("Permutation generation cancelled for %s", target)
        }
    }()

    for host := range resultCh {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        fmt.Fprintf(w, "data: %s\n\n", host)
        flusher.Flush()
    }
}

func zoneTransferStream(w http.ResponseWriter, r *http.Request) {
    target := r.URL.Query().Get("target")
    if target == "" {
        http.Error(w, "missing target", http.StatusBadRequest)
        return
    }
    
    sseHeader(w)
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "streaming not supported", http.StatusInternalServerError)
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
    defer cancel()
    
    jobKey := target + "_zone"
    
    jobsMu.Lock()
    if existingCancel, exists := jobs[jobKey]; exists {
        existingCancel()
    }
    jobs[jobKey] = cancel
    jobsMu.Unlock()

    defer func() {
        jobsMu.Lock()
        delete(jobs, jobKey)
        jobsMu.Unlock()
        log.Printf("Zone transfer attempt ended for %s", target)
    }()

    resultCh := make(chan string, 100)
    
    go func() {
        defer close(resultCh)
        
        // First, find NS records for the domain
        nsRecords, err := net.LookupNS(target)
        if err != nil {
            log.Printf("Failed to lookup NS records for %s: %v", target, err)
            return
        }
        
        
        // Try zone transfer against each nameserver
        for _, ns := range nsRecords {
            select {
            case <-ctx.Done():
                return
            default:
            }
            
            log.Printf("Attempting zone transfer from %s for %s", ns.Host, target)
            
            // Note: Zone transfers are typically restricted and will fail for most domains
            // This is more of a check to see if misconfigured DNS servers exist
            
            // Try to connect to the nameserver on port 53
            conn, err := net.DialTimeout("tcp", net.JoinHostPort(ns.Host, "53"), 5*time.Second)
            if err != nil {
                log.Printf("Failed to connect to nameserver %s: %v", ns.Host, err)
                continue
            }
            conn.Close()
            
            // If we can connect, it might be worth noting (though actual zone transfer would need DNS library)
            log.Printf("Successfully connected to nameserver %s (zone transfer would require DNS protocol implementation)", ns.Host)
        }
        
        log.Printf("Zone transfer attempt completed for %s (found %d nameservers)", target, len(nsRecords))
    }()

    // Send nameserver information as results
    for host := range resultCh {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        fmt.Fprintf(w, "data: %s\n\n", host)
        flusher.Flush()
    }
}

func generatePermutations(domain string) []string {
    var permutations []string
    
    // Add common prefixes and suffixes
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