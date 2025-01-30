package main

import (
	"fmt"
	"net/http"
)

func main() {
    // Initialise multiplexer
    mux := http.NewServeMux()

    // Define multiplexer handle
    mux.Handle("/", http.FileServer(http.Dir(".")))

    // Define server
    server := http.Server{Addr: ":8080", Handler: mux}

    // Start server
    err := server.ListenAndServe()
    if err != nil {
        panic(err)
    }
    fmt.Println("Server started...")
}
