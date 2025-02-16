package main

import (
    "encoding/json"
    "net/http"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
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

