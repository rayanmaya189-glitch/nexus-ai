package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

type WSMessage struct {
	Prompt    string `json:"prompt"`
	Model     string `json:"model"`
	SessionID string `json:"session_id"`
}

type WSChunk struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func handleWebSocket(ollamaURL string) websocket.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		log.Printf("[ws] client connected: %s", ws.RemoteAddr())

		for {
			var msg WSMessage
			err := websocket.JSON.Receive(ws, &msg)
			if err != nil {
				if err == io.EOF {
					log.Printf("[ws] client disconnected: %s", ws.RemoteAddr())
				} else {
					log.Printf("[ws] read error from %s: %v", ws.RemoteAddr(), err)
				}
				return
			}

			log.Printf("[ws] received message from %s: model=%s session=%s prompt_len=%d",
				ws.RemoteAddr(), msg.Model, msg.SessionID, len(msg.Prompt))

			if strings.TrimSpace(msg.Prompt) == "" {
				websocket.JSON.Send(ws, WSChunk{Response: "Error: empty prompt", Done: true})
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			streamOllama(ctx, ws, ollamaURL, msg)
			cancel()
		}
	})
}

func streamOllama(ctx context.Context, ws *websocket.Conn, ollamaURL string, msg WSMessage) {
	model := msg.Model
	if model == "" {
		model = "llama3.2:1b"
	}

	body := map[string]interface{}{
		"model":  model,
		"prompt": msg.Prompt,
		"stream": true,
	}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", ollamaURL+"/api/generate", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[ws] failed to create request: %v", err)
		websocket.JSON.Send(ws, WSChunk{Response: "Error creating request", Done: true})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ws] ollama request failed: %v", err)
		websocket.JSON.Send(ws, WSChunk{Response: "Error connecting to AI model", Done: true})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		log.Printf("[ws] ollama returned status %d: %s", resp.StatusCode, string(errBody))
		websocket.JSON.Send(ws, WSChunk{Response: fmt.Sprintf("AI model error (status %d)", resp.StatusCode), Done: true})
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		if ctx.Err() != nil {
			log.Printf("[ws] stream cancelled for session %s", msg.SessionID)
			return
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		var ollamaResp struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}
		if err := json.Unmarshal([]byte(line), &ollamaResp); err != nil {
			log.Printf("[ws] failed to parse ollama line: %v", err)
			continue
		}

		chunk := WSChunk{
			Response: ollamaResp.Response,
			Done:     ollamaResp.Done,
		}
		if err := websocket.JSON.Send(ws, chunk); err != nil {
			log.Printf("[ws] failed to send chunk to client: %v", err)
			return
		}

		if ollamaResp.Done {
			log.Printf("[ws] stream complete for session %s model=%s", msg.SessionID, model)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[ws] scanner error: %v", err)
	}
}

func wsCORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
