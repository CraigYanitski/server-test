package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/CraigYanitski/server-test/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
    ID         uuid.UUID  `json:"id"`
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
    Body       string     `json:"body"`
    UserID     uuid.UUID  `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
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

    // validate chirp length and sanitize
    var cleanChirp string
    if len(chp.Body) > 140 {
        respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
        return
    } else {
        cleanChirp = CleanChirpBody(chp.Body)
    }

    // create chirp
    params := database.CreateChirpParams{
        Body: cleanChirp, 
        UserID: chp.UserID,
    }
    chirp, err := cfg.dbQueries.CreateChirp(r.Context(), params)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "error creating chirp", err)
    }
    respondWithJSON(w, http.StatusCreated, Chirp(chirp))
    return
}

func CleanChirpBody(body string) string {
    clean_body := []string{}
    badWords := map[string]struct{} {
        "kerfuffle": {},
        "sharbert":  {},
        "fornax":    {},
    }
    for _, word := range strings.Fields(body) {
        if _, ok := badWords[strings.ToLower(word)]; ok {
                clean_body = append(clean_body, "****")
            } else {
                clean_body = append(clean_body, word)
            }
        }
    return strings.Join(clean_body, " ")
}

