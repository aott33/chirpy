package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits	atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	htmlStr := `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`
	result := fmt.Sprintf(htmlStr, cfg.fileserverHits.Load())
	w.Write([]byte(result))	
}

func (cfg *apiConfig) resetHanlder(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body	string `json:"body"`
	}

	type returnError struct {
		Error	string `json:"error"`
	}

	type returnValid struct {
		Valid	bool `json:"valid"`
	}
	
	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)

	params := parameters{}

	respBodyError := returnError{}
	
	respBodyValid := returnValid{}

	err := decoder.Decode(&params)

	if err != nil {	
		respBodyError.Error = fmt.Sprintf("Something went wrong: %s", err)
		dat, err := json.Marshal(respBodyError)
		if err != nil {
			fmt.Printf("Something went wrong: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}	
		w.WriteHeader(http.StatusBadRequest)
		w.Write(dat)
		return
	}

	if len(params.Body) > 400 {
		respBodyError.Error = "Chirp is too long"
		dat, err := json.Marshal(respBodyError)
		if err != nil {
			fmt.Printf("Something went wrong: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(dat)
		return
	}

	respBodyValid.Valid = true
	dat, err := json.Marshal(respBodyValid)
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
	w.Write(dat)
}


func main() {
	mux := http.NewServeMux()

	apiCfg := &apiConfig{}

	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServerHandler))

	mux.HandleFunc("GET /api/healthz", healthHandler)

	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)

	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	mux.HandleFunc("POST /admin/reset", apiCfg.resetHanlder)

	server := &http.Server{Addr: ":8080", Handler: mux}

	fmt.Printf("Server starting on port 8080\n")

	server.ListenAndServe()
}
