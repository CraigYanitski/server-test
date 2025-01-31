package main

import (
	"fmt"
	"net/http"
    "sync/atomic"
)

type apiConfig struct {
    fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    // increment server hits
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg.fileserverHits.Add(1)
        next.ServeHTTP(w, r)
    })
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(200)
    hits := cfg.fileserverHits.Load()
    w.Write([]byte(fmt.Sprintf("Hits: %v", hits)))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
    // reset server hits
    cfg.fileserverHits.Store(0)
    w.WriteHeader(200)
    w.Write([]byte("Hits successfully reset!"))
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(200)
    w.Write([]byte("OK"))
}

func main() {
    // Initialise multiplexer
    mux := http.NewServeMux()

    // Initialise api config
    apiCfg := apiConfig{}

    // Define handlers
    appFileServer := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

    // Define multiplexer handle
    mux.Handle("/app/", apiCfg.middlewareMetricsInc(appFileServer))
    mux.HandleFunc("GET /healthz", handlerHealthz)
    mux.HandleFunc("GET /metrics", apiCfg.handlerMetrics)
    mux.HandleFunc("POST /reset", http.HandlerFunc(apiCfg.handlerReset))

    // Define server
    port := "8080"
    server := http.Server{
        Addr:    ":" + port, 
        Handler: mux,
    }

    // Start server
    fmt.Printf("Serving files from / on port: %v\n", port)
    err := server.ListenAndServe()
    if err != nil {
        panic(err)
    }
}

