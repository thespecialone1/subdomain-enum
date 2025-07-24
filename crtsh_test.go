package main

import (
    "fmt"
    "io"
    "net/http"
    "time"
)

func main() {
    // Test the crt.sh API directly
    domain := "example.com" // Change this to your test domain
    
    apiURL := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)
    fmt.Printf("Testing URL: %s\n", apiURL)
    
    client := &http.Client{Timeout: 30 * time.Second}
    req, _ := http.NewRequest("GET", apiURL, nil)
    req.Header.Set("Accept", "application/json")
    req.Header.Set("User-Agent", "curl/7.64.1")
    
    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer resp.Body.Close()
    
    fmt.Printf("Status: %s\n", resp.Status)
    fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("Read error: %v\n", err)
        return
    }
    
    fmt.Printf("Response length: %d bytes\n", len(body))
    fmt.Printf("First 500 characters:\n%s\n", string(body[:min(500, len(body))]))
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}