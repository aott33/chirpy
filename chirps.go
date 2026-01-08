package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/aott33/chirpy/internal/auth"
	"github.com/aott33/chirpy/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type chirpParams struct {
	Body		string		`json:"body"`
}

type errorResponse struct {
	Error		string		`json:"error"`
}

type chirpResponse struct {
	ID			uuid.UUID	`json:"id"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
	Body		string		`json:"body"`
	UserID		uuid.UUID	`json:"user_id"`
}

var badWords = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

var bleepStr = "****"

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {	
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	uuidVal, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {	
		w.WriteHeader(http.StatusUnauthorized)	
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {	
		w.WriteHeader(http.StatusInternalServerError)	
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}


	if uuidVal != chirp.UserID {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	
	err = cfg.dbQueries.DeleteChirpByID(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	
}

func (cfg *apiConfig) getChirpByIDHandler (w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {	
		writeJSON(w, http.StatusNotFound, errorResponse {
			Error: "Something went wrong",
		})	
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {	
		writeJSON(w, http.StatusNotFound, errorResponse {
			Error: "Something went wrong",
		})	
		return
	}

	chirpResponse := chirpResponse {
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}

	dat, err := json.Marshal(chirpResponse)
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
	
}

func (cfg *apiConfig) getChirpsHandler (w http.ResponseWriter, r *http.Request) {
	var chirpSlc []chirpResponse

	chirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {	
		writeJSON(w, http.StatusBadRequest, errorResponse {
			Error: "Something went wrong",
		})	
		return
	}

	for i := range chirps {
		chirp := chirps[i]
		tempChirp := chirpResponse{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		}
		chirpSlc = append(chirpSlc, tempChirp)
	}

	dat, err := json.Marshal(chirpSlc)
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

func (cfg *apiConfig) createChirpHandler (w http.ResponseWriter, r *http.Request) {
	var params chirpParams

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {	
		writeJSON(w, http.StatusUnauthorized, errorResponse {
			Error: fmt.Sprintf("Something went wrong: %v", err),
		})	
		return
	}

	uuidVal, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {	
		writeJSON(w, http.StatusUnauthorized, errorResponse {
			Error: fmt.Sprintf("Something went wrong: %v", err),
		})	
		return
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
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

	chirpCreated, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body: cleanedMsg,
		UserID: uuidVal,
	})
	if err != nil {	
		writeJSON(w, http.StatusBadRequest, errorResponse {
			Error: "Something went wrong",
		})	
		return
	}

	chirpResponse := chirpResponse {
		ID: chirpCreated.ID,
		CreatedAt: chirpCreated.CreatedAt,
		UpdatedAt: chirpCreated.UpdatedAt,
		Body: chirpCreated.Body,
		UserID: chirpCreated.UserID,
	}

	dat, err := json.Marshal(chirpResponse)
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(dat)		
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
