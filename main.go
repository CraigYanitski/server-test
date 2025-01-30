package main

import (
	"fmt"
	"net/http"
)

func main() {
    // Initialise multiplexer
    mux := http.NewServeMux()

    // Define multiplexer handle
    mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        w.WriteHeader(200)
        w.Write([]byte("OK"))
    })

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

