package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type HealthRespones struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

type ProfileResponse struct {
	NameMessage string `json:"name"`
	LangMessage string `json:"lang"`
}

func healthHandler(w http.ResponseWriter, r *http.Request)  {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type","application/json")

	res := HealthRespones{
		Status: "OK",
		Time:	time.Now().Format(time.RFC3339),	
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Println("faild to encode respones:", err)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")

	if name == "" {
		name = "guest"
	}

	w.Header().Set("Content-Type", "application/json")

	res := HelloResponse{
		Message: "Hello, " + name + "!",
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Println("failed to encode respones:", err)
	}
}

func ProfileHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")

	lang := r.URL.Query().Get("lang")

	if name == " " {
		name = "guest"
	}

	if lang == " " {
		lang = "empty"
	}

	w.Header().Set("Content-Type", "application/json")

	res := ProfileResponse{
		NameMessage: name,
		LangMessage: lang,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Println("failed to encode response:",err)
	}

}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/profile",ProfileHandle)

	log.Println("server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}