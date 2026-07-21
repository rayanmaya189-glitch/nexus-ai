package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aeroxe/nexus-backend/internal/application/usecases"
	"github.com/aeroxe/nexus-backend/internal/config"
	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/aeroxe/nexus-backend/internal/infrastructure/database"
	"github.com/aeroxe/nexus-backend/internal/interfaces/rest"
	"github.com/aeroxe/nexus-backend/internal/middleware"
	"github.com/aeroxe/nexus-backend/pkg/logger"
)

func main() {
	svcLogger := logger.New("rag-service")
	svcLogger.Info("Starting RAG Service")

	cfg, err := config.LoadConfig("")
	if err != nil {
		svcLogger.Fatal(fmt.Sprintf("Failed to load config: %v", err))
	}

	db, err := database.NewDB(
		getEnv("DB_HOST", cfg.Database.Host),
		getEnvInt("DB_PORT", cfg.Database.Port),
		getEnv("DB_USER", cfg.Database.User),
		getEnv("DB_PASSWORD", cfg.Database.Password),
		getEnv("DB_NAME", cfg.Database.DBName),
		getEnv("DB_SSLMODE", cfg.Database.SSLMode),
	)
	if err != nil {
		svcLogger.Fatal(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	defer db.Close()
	svcLogger.Info("Connected to PostgreSQL database")

	docRepo := database.NewPostgresDocumentRepository(db.Pool)
	chunkRepo := database.NewPostgresDocumentChunkRepository(db.Pool)
	docSetRepo := database.NewPostgresDocumentSetRepository(db.Pool)

	docUseCase := usecases.NewDocumentUseCase(docRepo, chunkRepo, docSetRepo)
	docHandler := rest.NewDocumentHandler(docUseCase)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "healthy", "service": "rag-service", "version": "1.0.0",
		})
	})

	mux.HandleFunc("/api/v1/rag/documents", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			docHandler.ListDocuments(w, r)
		case http.MethodPost:
			docHandler.CreateDocument(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET or POST only")
		}
	})

	mux.HandleFunc("/api/v1/rag/documents/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			docHandler.GetDocument(w, r)
		case http.MethodPut:
			docHandler.UpdateDocument(w, r)
		case http.MethodDelete:
			docHandler.DeleteDocument(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET, PUT or DELETE only")
		}
	})

	mux.HandleFunc("/api/v1/rag/ingest", ingestHandler(docRepo, chunkRepo))
	mux.HandleFunc("/api/v1/rag/query", queryHandler)

	mux.HandleFunc("/api/v1/rag/sets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			docHandler.ListDocumentSets(w, r)
		case http.MethodPost:
			docHandler.CreateDocumentSet(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET or POST only")
		}
	})

	handler := middleware.RequestIDMiddleware(mux)

	port := getEnv("PORT", "8093")
	addr := fmt.Sprintf(":%s", port)
	svcLogger.Info(fmt.Sprintf("RAG Service listening on %s", addr))

	if err := http.ListenAndServe(addr, handler); err != nil {
		svcLogger.Fatal(fmt.Sprintf("Server failed: %v", err))
	}
}

func ingestHandler(docRepo *database.PostgresDocumentRepository, chunkRepo *database.PostgresDocumentChunkRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
			return
		}

		var req struct {
			DocumentID int64 `json:"document_id"`
			ChunkSize  int   `json:"chunk_size"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}

		doc, err := docRepo.FindByID(r.Context(), req.DocumentID)
		if err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Document not found")
			return
		}

		chunkSize := req.ChunkSize
		if chunkSize <= 0 {
			chunkSize = 500
		}

		content := doc.Content
		var chunks []*entities.DocumentChunk
		for i := 0; i < len(content); i += chunkSize {
			end := i + chunkSize
			if end > len(content) {
				end = len(content)
			}
			chunk := &entities.DocumentChunk{
				DocumentID: doc.ID,
				TenantID:   doc.TenantID,
				Content:    content[i:end],
				ChunkIndex: len(chunks),
				Metadata:   fmt.Sprintf(`{"document_title":"%s","chunk_index":%d}`, doc.Title, len(chunks)),
			}
			chunks = append(chunks, chunk)
		}

		if err := chunkRepo.CreateBatch(r.Context(), chunks); err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}

		doc.ChunkCount = len(chunks)
		docRepo.Update(r.Context(), doc)

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]interface{}{
				"document_id": doc.ID,
				"chunk_count": len(chunks),
				"status":      "ingested",
			},
		})
	}
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var req struct {
		Query    string `json:"query"`
		TenantID int64  `json:"tenant_id"`
		TopK     int    `json:"top_k"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if req.Query == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "query is required")
		return
	}

	topK := req.TopK
	if topK <= 0 {
		topK = 5
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"query":   req.Query,
			"results": []interface{}{},
			"top_k":   topK,
			"message": "RAG query processed - semantic search requires embeddings",
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error": map[string]string{"code": code, "message": message},
	})
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		var n int
		fmt.Sscanf(val, "%d", &n)
		if n > 0 {
			return n
		}
	}
	return defaultVal
}
