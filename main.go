package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/CraigYanitski/server-test/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
    fileserverHits  atomic.Int32
    dbQueries       *database.Queries
    platform        string
    secret          string
}

func main() {
    // Load environment variables
    godotenv.Load()
    dbURL := os.Getenv("DB_URL")
    if dbURL == "" {
        log.Fatal("DB_RUL must be set")
    }
    platform := os.Getenv("PLATFORM")
    if platform == "" {
        log.Fatal("PLATFORM must be set")
    }
    secret := os.Getenv("SECRET")
    if secret == "" {
        log.Fatal("SECRET must be set")
    }
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatalf("error opening database: %s", err)
    }
    dbQueries := database.New(db)

    // Create API config with DB queries
    apiCfg := apiConfig{
        fileserverHits: atomic.Int32{},
        dbQueries:      dbQueries,
        platform:       platform,
        secret:         secret,
    }

    // Initialise multiplexer
    mux := http.NewServeMux()

    // Initialise api config
    // Define server
    const fsPath = "."
    const port = "8080"
    server := http.Server{
        Addr:    ":" + port, 
        Handler: mux,
    }

    // Define handlers
    appFileServer := http.StripPrefix("/app", http.FileServer(http.Dir(fsPath)))

    // Define multiplexer handle
    mux.Handle("/app/", apiCfg.middlewareMetricsInc(appFileServer))
    
    // API status
    mux.HandleFunc("GET /api/healthz", handlerHealthz)

    // API users
    mux.HandleFunc("POST /api/users", http.HandlerFunc(apiCfg.handlerCreateUser))
    mux.HandleFunc("PUT /api/users", http.HandlerFunc(apiCfg.handlerUpdateUser))
    mux.HandleFunc("POST /api/login", http.HandlerFunc(apiCfg.handlerLogin))
    mux.HandleFunc("POST /api/refresh", http.HandlerFunc(apiCfg.handlerRefresh))
    mux.HandleFunc("POST /api/revoke", http.HandlerFunc(apiCfg.handlerRevoke))

    // Polka webhook
    mux.HandleFunc("POST /api/polka/webhooks", http.HandlerFunc(apiCfg.handlerUpgradeUserToRed))

    // API chirps
    mux.HandleFunc("POST /api/chirps", http.HandlerFunc(apiCfg.handlerCreateChirp))
    mux.HandleFunc("GET /api/chirps/{chirp_id}", http.HandlerFunc(apiCfg.handlerGetChirps))
    mux.HandleFunc("DELETE /api/chirps/{chirp_id}", http.HandlerFunc(apiCfg.handlerDeleteChirp))
    
    // Admin stuff
    mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
    mux.HandleFunc("POST /admin/reset", http.HandlerFunc(apiCfg.handlerReset))

    // Start server
    fmt.Printf("Serving files from / on port: %v\n", port)
    log.Fatal(server.ListenAndServe())
}

