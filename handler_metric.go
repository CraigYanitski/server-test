package main

import (
    "fmt"
    "net/http"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    // increment server hits
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg.fileserverHits.Add(1)
        next.ServeHTTP(w, r)
    })
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
    // check if there is admin access
    if !cfg.CheckAdmin(w) {
        return
    }
    // serve metrics HTML
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(200)
    hits := cfg.fileserverHits.Load()
    hitsHTML := fmt.Sprintf(
        "<html>" +
        "<body>" +
        "<h1>Welcome, Chirpy Admin</h1>" +
        "<p>Chirpy has been visited %d times!</p>" +
        "</body>" +
        "</html>",
        hits,
    )
    w.Write([]byte(hitsHTML))
}

