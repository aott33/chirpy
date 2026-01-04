package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type createUserParams struct {
	Email		string 	`json:"email"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var params createUserParams

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userCreated, err := cfg.dbQueries.CreateUser(r.Context(), params.Email)
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := User{
		ID: userCreated.ID,
		CreatedAt: userCreated.CreatedAt,
		UpdatedAt: userCreated.UpdatedAt,
		Email: userCreated.Email,
	}

	dat, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(dat)	
}
