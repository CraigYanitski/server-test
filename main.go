package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
    // w.Write([]byte(fmt.Sprintf("Hits: %v", hits)))
    w.Write([]byte(hitsHTML))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
    // reset server hits
    cfg.fileserverHits.Store(0)
    w.WriteHeader(200)
    w.Write([]byte("Hits successfully reset!"))
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

type Chirp struct {
    Body string `json:"body"`
}
type CleanChirp struct {
    Body string `json:"cleaned_body"`
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
    // type chirpError struct {Error string `json:"error"`}

    decoder := json.NewDecoder(r.Body)
    chp := &Chirp{}
    err := decoder.Decode(chp)
    if err != nil {
        //fmt.Printf("error decoding a JSON: %s\n", err)
        //w.WriteHeader(500)
        respondWithError(w, http.StatusInternalServerError, "error decoding JSON", err)
        return
    }

    if len(chp.Body) > 140 {
        respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
    } else {
        cleanChirp := CleanChirpBody(chp)
        respondWithJSON(w, http.StatusOK, cleanChirp)
    }
}

func CleanChirpBody(chp *Chirp)  *CleanChirp {
    clean_body := []string{}
    for _, word := range strings.Fields(chp.Body) {
        if strings.Contains(strings.ToLower(word), "kerfuffle") ||
            strings.Contains(strings.ToLower(word), "sharbert") ||
            strings.Contains(strings.ToLower(word), "fornax") {
                clean_body = append(clean_body, "****")
            } else {
                clean_body = append(clean_body, word)
            }
        }
    return &CleanChirp{Body: strings.Join(clean_body, " ")}
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
        log.Printf("Responding with 5XX error: %s: %s", msg, err)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func main() {
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

    apiCfg := apiConfig{}

    // Define handlers
    appFileServer := http.StripPrefix("/app", http.FileServer(http.Dir(fsPath)))

    // Define multiplexer handle
    mux.Handle("/app/", apiCfg.middlewareMetricsInc(appFileServer))
    
    mux.HandleFunc("GET /api/healthz", handlerHealthz)
    mux.HandleFunc("POST /api/validate_chirp", http.HandlerFunc(validateChirp))
    
    mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
    mux.HandleFunc("POST /admin/reset", http.HandlerFunc(apiCfg.handlerReset))

    // Start server
    fmt.Printf("Serving files from / on port: %v\n", port)
    err := server.ListenAndServe()
    if err != nil {
        panic(err)
    }
}

