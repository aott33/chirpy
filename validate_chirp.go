package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	_ "github.com/lib/pq"
)

type chirpParams struct {
	Body	string `json:"body"`
}

type errorResponse struct {
	Error	string `json:"error"`
}

type cleanBodyResponse struct {
	CleanedBody	string `json:"cleaned_body"`
}

type validateResponse struct {
	Valid	bool `json:"valid"`
}	

var badWords = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

var bleepStr = "****"

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var params chirpParams

	err := decoder.Decode(&params)
	if err != nil {	
		writeJSON(w, http.StatusBadRequest, errorResponse {
			Error: "Something went wrong",
		})	
		return
	}

	if len(params.Body) > 140 {
		writeJSON(w, http.StatusBadRequest, errorResponse {
			Error: "Chirp is too long",
		})
		return
	}

	cleanedMsg := checkMsg(params.Body)

	writeJSON(w, http.StatusOK, cleanBodyResponse {
		CleanedBody: cleanedMsg,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	dat, err := json.Marshal(v)
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(dat)
}

func checkMsg(msg string) string { 
	lowerMsg := strings.ToLower(msg)
	
	strSlice := strings.Split(lowerMsg, " ")
	originalSlice := strings.Split(msg, " ")

	for i := range strSlice {
		if slices.Contains(badWords, strSlice[i]) {
			strSlice[i] = bleepStr
		} else {
			strSlice[i] = originalSlice[i]
		}
	}

	cleanedMsg := strings.Join(strSlice, " ")

	return cleanedMsg
}
