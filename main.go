package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
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

type Item struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type CreateItemRespones struct {
	Message string `json:"message"`
	Item    Item   `json:"item"`
}

type ImageRequestResponse struct {
	URL string `json:"url"`
	Width int `json:"width"`
	Format string `json:"format"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	res := HealthRespones{
		Status: "OK",
		Time:   time.Now().Format(time.RFC3339),
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
		log.Println("failed to encode response:", err)
	}

}

func createItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var item Item

	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	w.Header().Set("Contente-Type", "application/json")

	res := CreateItemRespones{
		Message: "item created",
		Item:    item,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Println("failed to encode response:", err)
	}

}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allow", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()

	originURL := query.Get("url")
	widthText := query.Get("w")
	format := query.Get("format")

	if originURL == " " {
		http.Error(w, "w is required", http.StatusBadRequest)
		return
	}

	if widthText == " " {
		http.Error(w, "w si required", http.StatusBadRequest)
		return
	}

	width, err := strconv.Atoi(widthText)
	if err != nil {
		http.Error(w, "w must be number", http.StatusBadRequest)
		return
	}

	if width <= 0 {
		http.Error(w, "w must be greater than 0", http.StatusBadRequest)
		return
	}
	
	if format == " " {
		format = "jpeg"
	}

	if format != "jpeg" && format != "webp" {
		http.Error(w, "format must be jpeg or webp", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(originURL)
	if err != nil {
		http.Error(w, "failed to fetch origin image", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "origin returned non-200 status", http.StatusBadGateway)
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == " " {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Image-Proxy-Width", strconv.Itoa(width))
	w.Header().Set("X-Image-Proxy-Format", format)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Println("failed to copy image response:", err)
	}
}



func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/profile", ProfileHandle)
	mux.HandleFunc("/items", createItemHandler)
	mux.HandleFunc("/image", imageHandler)

	log.Println("server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
