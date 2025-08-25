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

type BatchRequest struct {
	Texts []string `json:"texts"`
}

type BatchResponse struct {
	Results []SentimentResponse `json:"results"`
	Summary BatchSummary        `json:"summary"`
}

type BatchSummary struct {
	Total    int `json:"total"`
	Positive int `json:"positive"`
	Negative int `json:"negative"`
	Neutral  int `json:"neutral"`
}

type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content ContentResponse `json:"content"`
}

type ContentResponse struct {
	Parts []PartResponse `json:"parts"`
}

type PartResponse struct {
	Text string `json:"text"`
}

func analyzeSentimentAI(text string) (string, float64, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Printf("DEBUG: No GEMINI_API_KEY found, using fallback analysis")
		return analyzeSentimentSimple(text), 0.8, nil
	}
	
	log.Printf("DEBUG: Using Gemini Pro API for text: %.50s...", text)

	prompt := fmt.Sprintf(`Analyze the sentiment of this text and respond with ONLY one word: "positive", "negative", or "neutral"

Text: "%s"

Response:`, text)

	reqBody := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return analyzeSentimentSimple(text), 0.5, nil
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("DEBUG: Gemini API request error: %v", err)
		return analyzeSentimentSimple(text), 0.5, nil
	}
	
	log.Printf("DEBUG: Gemini API response status: %d", resp.StatusCode)
	
	if resp.StatusCode != 200 {
		log.Printf("DEBUG: Gemini API error status %d, using fallback", resp.StatusCode)
		return analyzeSentimentSimple(text), 0.5, nil
	}
	defer resp.Body.Close()

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		log.Printf("DEBUG: Gemini JSON decode error: %v", err)
		return analyzeSentimentSimple(text), 0.5, nil
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		aiResponse := geminiResp.Candidates[0].Content.Parts[0].Text
		log.Printf("DEBUG: Gemini AI response: %s", aiResponse)
		
		sentiment := strings.ToLower(strings.TrimSpace(aiResponse))
		if sentiment == "positive" || sentiment == "negative" || sentiment == "neutral" {
			log.Printf("DEBUG: Using Gemini result: %s with score 0.95", sentiment)
			return sentiment, 0.95, nil
		}
		
		log.Printf("DEBUG: Unexpected Gemini response format: %s", sentiment)
	} else {
		log.Printf("DEBUG: Empty Gemini response, using fallback")
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

func handleBatchSentiment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	var req BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if len(req.Texts) == 0 {
		http.Error(w, "Texts array is required", http.StatusBadRequest)
		return
	}
	
	if len(req.Texts) > 50 {
		http.Error(w, "Maximum 50 texts allowed per batch", http.StatusBadRequest)
		return
	}
	
	var results []SentimentResponse
	summary := BatchSummary{Total: len(req.Texts)}
	
	for _, text := range req.Texts {
		sentiment, score, _ := analyzeSentimentAI(text)
		
		result := SentimentResponse{
			Text:      text,
			Sentiment: sentiment,
			Score:     score,
		}
		
		results = append(results, result)
		
		switch sentiment {
		case "positive":
			summary.Positive++
		case "negative":
			summary.Negative++
		default:
			summary.Neutral++
		}
	}
	
	response := BatchResponse{
		Results: results,
		Summary: summary,
	}
	
	json.NewEncoder(w).Encode(response)
}

func handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}
	
	var req BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	var results []SentimentResponse
	for _, text := range req.Texts {
		sentiment, score, _ := analyzeSentimentAI(text)
		results = append(results, SentimentResponse{
			Text:      text,
			Sentiment: sentiment,
			Score:     score,
		})
	}
	
	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=sentiment_analysis.csv")
		
		fmt.Fprintf(w, "Text,Sentiment,Score\n")
		for _, result := range results {
			fmt.Fprintf(w, "\"%s\",%s,%.2f\n", result.Text, result.Sentiment, result.Score)
		}
		
	default: // json
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=sentiment_analysis.json")
		json.NewEncoder(w).Encode(results)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"version": "1.0",
		"features": "single-analysis,batch-analysis,data-export",
	})
}

func main() {
	// API endpoints first
	http.HandleFunc("/analyze", handleSentiment)
	http.HandleFunc("/analyze/batch", handleBatchSentiment)
	http.HandleFunc("/export", handleExport)
	http.HandleFunc("/health", handleHealth)
	
	// Serve static files
	http.Handle("/", http.FileServer(http.Dir("./public/")))
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}