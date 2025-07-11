package main

import (
  "fmt"
  "io"
  "log"
  "mime"
  "net/http"
  "net/url"
  "os"
)

func handler(w http.ResponseWriter, r *http.Request) {
  message := os.Getenv("MESSAGE")
  instanceId := os.Getenv("CLOUDFLARE_DEPLOYMENT_ID")
  fmt.Fprintf(w, "Hi, I'm a container and this is my message: \"%s\", my instance ID is: %s", message, instanceId)
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
  panic("This is a panic")
}

func fetchImageHandler(w http.ResponseWriter, r *http.Request) {
  raw := r.URL.Query().Get("url")
  if raw == "" {
    http.Error(w, "url parameter is required", http.StatusBadRequest)
    return
  }
  parsed, err := url.Parse(raw)
  if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
    http.Error(w, "invalid url", http.StatusBadRequest)
    return
  }

  resp, err := http.Get(parsed.String())
  if err != nil {
    http.Error(w, "failed to fetch image: "+err.Error(), http.StatusBadGateway)
    return
  }
  defer resp.Body.Close()

  w.WriteHeader(resp.StatusCode)

  ct := resp.Header.Get("Content-Type")
  if ct != "" {
    w.Header().Set("Content-Type", ct)
    if exts, _ := mime.ExtensionsByType(ct); len(exts) > 0 {
      w.Header().Set("Content-Disposition", "inline; filename=\"image"+exts[0]+"\"")
    }
  } else {
    w.Header().Set("Content-Type", "application/octet-stream")
  }

  if _, err := io.Copy(w, resp.Body); err != nil {
    log.Println("error streaming image:", err)
  }
}

func main() {
  http.HandleFunc("/", handler)
  http.HandleFunc("/container", handler)
  http.HandleFunc("/error", errorHandler)
  http.HandleFunc("/fetch-image", fetchImageHandler)
  log.Fatal(http.ListenAndServe(":8080", nil))
}
