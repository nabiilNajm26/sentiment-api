package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

func analyzeSentimentAI(text string) (string, float64, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Fallback to simple analysis
		return analyzeSentimentSimple(text), 0.8, nil
	}

	prompt := fmt.Sprintf(`Analyze the sentiment of this text and respond with ONLY one word: "positive", "negative", or "neutral"

Text: "%s"

Response:`, text)

	reqBody := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return analyzeSentimentSimple(text), 0.5, nil
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return analyzeSentimentSimple(text), 0.5, nil
	}
	defer resp.Body.Close()

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return analyzeSentimentSimple(text), 0.5, nil
	}

	if len(openAIResp.Choices) > 0 {
		sentiment := strings.ToLower(strings.TrimSpace(openAIResp.Choices[0].Message.Content))
		if sentiment == "positive" || sentiment == "negative" || sentiment == "neutral" {
			return sentiment, 0.95, nil
		}
	}

	return analyzeSentimentSimple(text), 0.5, nil
}

func analyzeSentimentSimple(text string) string {
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
		return "positive"
	} else if negativeCount > positiveCount {
		return "negative"
	}
	
	return "neutral"
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
	
	sentiment, score, _ := analyzeSentimentAI(req.Text)
	
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