package main

import (
    "bufio"
    "context"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "io"
    "log"
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
}

type probeResponse struct {
    Status string `json:"status"`
    Title  string `json:"title"`
    Error  string `json:"error"`
}

func main() {
    mux := http.NewServeMux()

    // Serve static UI
    mux.Handle("/", http.FileServer(http.Dir("./public/")))

    // SSE streams  
    mux.HandleFunc("/api/wayback/stream", waybackStream)
    mux.HandleFunc("/api/crtsh/stream", crtshStream)

    // Probe handler
    mux.HandleFunc("/api/probe", probeHandler)

    // Abort handler
    mux.HandleFunc("/api/abort", abortHandler)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Listening on port %s…", port)
    log.Fatal(http.ListenAndServe(":"+port, mux))
}

func sseHeader(w http.ResponseWriter) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")
}

func probeHandler(w http.ResponseWriter, r *http.Request) {
    targetURL := r.URL.Query().Get("url")
    if targetURL == "" {
        http.Error(w, "missing url parameter", http.StatusBadRequest)
        return
    }

    // Parse URL to validate it
    _, err := url.Parse(targetURL)
    if err != nil {
        writeProbeError(w, "invalid URL", err)
        return
    }

    // Create HTTP client with timeout and custom transport
    client := &http.Client{
        Timeout: 10 * time.Second,
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true, // Skip certificate verification for self-signed certs
            },
        },
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            if len(via) >= 3 {
                return fmt.Errorf("too many redirects")
            }
            return nil
        },
    }

    // Make the request
    req, err := http.NewRequest("GET", targetURL, nil)
    if err != nil {
        writeProbeError(w, "failed to create request", err)
        return
    }

    req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; SubdomainScanner/1.0)")
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

    resp, err := client.Do(req)
    if err != nil {
        writeProbeError(w, "connection failed", err)
        return
    }
    defer resp.Body.Close()

    // Read response body (limit to 1MB to prevent abuse)
    body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
    if err != nil {
        writeProbeError(w, "failed to read response", err)
        return
    }

    // Extract title
    title := "No title"
    if matches := titleRe.FindStringSubmatch(string(body)); len(matches) > 1 {
        title = strings.TrimSpace(matches[1])
        // Clean up title (remove extra whitespace, newlines)
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
    // Cancel both wayback and crtsh streams for this target
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

    // Create cancellable context
    ctx, cancel := context.WithCancel(r.Context())
    jobKey := target + "_wayback"
    
    jobsMu.Lock()
    // Cancel any existing job for this target first
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

    // Channel for results
    resultCh := make(chan string, 100)
    
    // Start fetching data in background
    go func() {
        defer close(resultCh)
        
        apiURL := fmt.Sprintf(
            "https://web.archive.org/cdx/search/cdx?url=*.%s/*&output=text&fl=original&collapse=urlkey",
            target,
        )
        req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
        resp, err := http.DefaultClient.Do(req)
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

    // Stream results to client
    for host := range resultCh {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        // Send to client immediately
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

    ctx, cancel := context.WithCancel(r.Context())
    jobKey := target + "_crtsh"
    
    jobsMu.Lock()
    // Cancel any existing job for this target first
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

    // Track processed entries to avoid duplicates
    processedEntries := make(map[string]struct{})

    // Channel for results
    resultCh := make(chan string, 100)
    
    // Start fetching data in background
    go func() {
        defer close(resultCh)
        
        apiURL := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", target)
        req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
        req.Header.Set("Accept", "application/json")
        req.Header.Set("User-Agent", "curl/7.64.1")

        client := &http.Client{Timeout: 30 * time.Second}
        resp, err := client.Do(req)
        if err != nil {
            log.Printf("crt.sh API error: %v", err)
            return
        }
        defer resp.Body.Close()

        // Check if we actually got JSON
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
        if len(entries) == 0 {
            log.Printf("crt.sh returned no entries for domain: %s", target)
            return
        }
        
        log.Printf("crt.sh found %d certificate entries for %s", len(entries), target)
        
        for _, e := range entries {
            select {
            case <-ctx.Done():
                return
            default:
            }
            
            // Create a unique key for this entry to avoid processing duplicates
            entryKey := e.NameValue
            if _, processed := processedEntries[entryKey]; processed {
                continue
            }
            processedEntries[entryKey] = struct{}{}
            
            for _, name := range strings.Split(e.NameValue, "\n") {
                host := strings.ToLower(strings.TrimSpace(name))
                // Remove any wildcard characters
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

    // Stream results to client
    for host := range resultCh {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        // Send to client immediately
        fmt.Fprintf(w, "data: %s\n\n", host)
        flusher.Flush()
    }
}