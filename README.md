# Sentiment API

A simple REST API for analyzing text sentiment, built with Go.

## What it does

Send any text and get back whether it's positive, negative, or neutral along with a confidence score.

## API Endpoints

### Analyze Sentiment
**POST** `/analyze`

```json
{
  "text": "I really enjoyed this movie!"
}
```

Response:
```json
{
  "text": "I really enjoyed this movie!",
  "sentiment": "positive",
  "score": 1.0
}
```

### Health Check  
**GET** `/health`

Returns server status.

## Running Locally

```bash
go run main.go
```

Server starts on port 8080.

## Example Usage

```bash
curl -X POST http://localhost:8080/analyze \
  -H "Content-Type: application/json" \
  -d '{"text": "This is amazing!"}'
```

## Tech Stack

- Go
- Standard library only
- Docker ready

## Deployment

Configured for cloud deployment with Docker support.