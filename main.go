package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/aott33/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits	atomic.Int32
	dbQueries		database.Queries
}

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

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}

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


func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Error opening database: %v", err)
		return
	}

	dbQueries := database.New(db)

	mux := http.NewServeMux()

	apiCfg := &apiConfig{}
	apiCfg.dbQueries = *dbQueries

	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServerHandler))

	mux.HandleFunc("GET /api/healthz", healthHandler)

	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)

	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

	server := &http.Server{Addr: ":8080", Handler: mux}

	fmt.Printf("Server starting on port 8080\n")

	server.ListenAndServe()
}
