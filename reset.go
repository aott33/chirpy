package main

import (
	"net/http"

	_ "github.com/lib/pq"
)

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	cfg.fileserverHits.Store(0)
	cfg.dbQueries.DeleteUsers(r.Context())
	w.WriteHeader(http.StatusOK)
}
