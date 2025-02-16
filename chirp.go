package main

import "strings"

type Chirp struct {
    Body string `json:"body"`
}
type CleanChirp struct {
    Body string `json:"cleaned_body"`
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

