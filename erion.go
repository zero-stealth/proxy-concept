package main

import (
    "log"
    "io"
    "net/http"
)

type ProxyServer struct {
    IP   string
    Port string
}

func (ps *ProxyServer) Init() {
    log.Println("[+] Starting proxy server on", ps.IP, ":", ps.Port)

    // create a new http server and register a handler
    server := &http.Server{Addr: ps.IP + ":" + ps.Port, Handler: http.HandlerFunc(ps.ProxyHandler)}

    // start the server
    log.Fatal(server.ListenAndServe())
}

func (ps *ProxyServer) ProxyHandler(w http.ResponseWriter, r *http.Request) {
    // extract the destination URL from the request
    destURL := r.URL.Scheme + "://" + r.URL.Host + r.URL.Path
    if r.URL.RawQuery != "" {
        destURL += "?" + r.URL.RawQuery
    }

    // create a new request to send to the remote server
    req, err := http.NewRequest(r.Method, destURL, r.Body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // modify the request headers to include information about the original client
    req.Header.Set("X-Forwarded-For", r.RemoteAddr)
    req.Header.Set("User-Agent", r.UserAgent())

    // send the modified request to the remote server and write the response back to the client
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    defer resp.Body.Close()

    // copy response headers to the client
    for k, vv := range resp.Header {
        for _, v := range vv {
            w.Header().Add(k, v)
        }
    }

    w.WriteHeader(resp.StatusCode)
    _, err = io.Copy(w, resp.Body)
    if err != nil {
        log.Printf("[-] error copying response body: %v", err)
    }
}

func main() {
    proxyServer := &ProxyServer{IP: "0.0.0.0", Port: "8080"}
    proxyServer.Init()
}
