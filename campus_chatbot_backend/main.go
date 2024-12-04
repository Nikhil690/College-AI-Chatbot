package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

type QueryRequest struct {
	Query string `json:"query"`
}

type QueryResponse struct {
	Response string `json:"response"`
	Source   string `json:"source"`
}

func PrettyPrintResponse(response QueryResponse) {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println(cyan("🤖 Query Response:"))
	fmt.Printf("%s: %s\n", green("Source"), response.Source)
	fmt.Printf("%s: %s\n", yellow("Response"), response.Response)
	fmt.Println("-------------------")
}

func SendQueryToPythonBackend(query string) (QueryResponse, error) {
	pythonBackendURL := os.Getenv("PYTHON_BACKEND_URL")
	if pythonBackendURL == "" {
		pythonBackendURL = "http://localhost:5000/query"
	}

	requestBody, err := json.Marshal(QueryRequest{Query: query})
	if err != nil {
		return QueryResponse{}, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(pythonBackendURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return QueryResponse{}, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return QueryResponse{}, fmt.Errorf("failed to read response: %v", err)
	}

	var queryResponse QueryResponse
	err = json.Unmarshal(body, &queryResponse)
	if err != nil {
		return QueryResponse{}, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	time.Sleep(5 * time.Second)

	return queryResponse, nil
}

func QueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var queryReq QueryRequest
	err := json.NewDecoder(r.Body).Decode(&queryReq)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received query: %s\n", queryReq.Query)

	response, err := SendQueryToPythonBackend(queryReq.Query)
	if err != nil {
		log.Printf("Error processing query: %v", err)
		http.Error(w, "Failed to process query", http.StatusInternalServerError)
		return
	}

	PrettyPrintResponse(response)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	color.Green("🟢 Health Check Successful")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Backend is healthy!"))
}

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		color.Cyan("Received request: %s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default configurations")
	}

	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = "127.0.0.1:8080"
	}

	r := mux.NewRouter()
	r.Use(LogRequest)
	r.HandleFunc("/query", QueryHandler).Methods("POST")
	r.HandleFunc("/health", HealthCheckHandler).Methods("GET")

	corsOptions := cors.Options{
		AllowedOrigins:   []string{"http://127.0.0.1:8081"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
		Debug:            true,
	}

	handler := cors.New(corsOptions).Handler(r)

	color.Cyan("🚀 Starting Go Backend on %s", serverAddress)
	log.Fatal(http.ListenAndServe(serverAddress, handler))
}

