package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

type SentimentRequest struct {
	Text string `json:"text"`
}

type SentimentResponse struct {
	Text      string  `json:"text"`
	Sentiment string  `json:"sentiment"`
	Score     float64 `json:"score"`
}

func analyzeSentiment(text string) (string, float64) {
	text = strings.ToLower(text)
	
	positiveWords := []string{"good", "great", "excellent", "amazing", "wonderful", "love", "happy", "awesome", "fantastic"}
	negativeWords := []string{"bad", "terrible", "awful", "hate", "horrible", "sad", "angry", "worst", "disappointed"}
	
	positiveCount := 0
	negativeCount := 0
	
	for _, word := range positiveWords {
		if strings.Contains(text, word) {
			positiveCount++
		}
	}
	
	for _, word := range negativeWords {
		if strings.Contains(text, word) {
			negativeCount++
		}
	}
	
	if positiveCount > negativeCount {
		return "positive", float64(positiveCount) / float64(positiveCount+negativeCount)
	} else if negativeCount > positiveCount {
		return "negative", float64(negativeCount) / float64(positiveCount+negativeCount)
	}
	
	return "neutral", 0.5
}

func handleSentiment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	var req SentimentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if req.Text == "" {
		http.Error(w, "Text field required", http.StatusBadRequest)
		return
	}
	
	sentiment, score := analyzeSentiment(req.Text)
	
	response := SentimentResponse{
		Text:      req.Text,
		Sentiment: sentiment,
		Score:     score,
	}
	
	json.NewEncoder(w).Encode(response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	http.HandleFunc("/analyze", handleSentiment)
	http.HandleFunc("/health", handleHealth)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}