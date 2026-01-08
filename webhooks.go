package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aott33/chirpy/internal/auth"
	"github.com/google/uuid"
)

type webhookUserUpgradeRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) webhooksHandler (w http.ResponseWriter, r *http.Request) {
	var userUpgradeReq webhookUserUpgradeRequest

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if apiKey != cfg.polkaKey {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&userUpgradeReq)
	if err != nil {
    	w.WriteHeader(http.StatusBadRequest)
		return
	}

	if userUpgradeReq.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID, err := uuid.Parse(userUpgradeReq.Data.UserID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = cfg.dbQueries.UpgradeUserRedChirpy(r.Context(), userID)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)	
}
