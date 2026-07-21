package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/aeroxe/nexus-backend/internal/config"
	"github.com/aeroxe/nexus-backend/internal/domain/entities"
	"github.com/aeroxe/nexus-backend/internal/infrastructure/database"
	"github.com/aeroxe/nexus-backend/internal/middleware"
	"github.com/aeroxe/nexus-backend/pkg/logger"
)

func main() {
	svcLogger := logger.New("configuration-service")
	svcLogger.Info("Starting Configuration Service")

	cfg, err := config.LoadConfig("")
	if err != nil {
		svcLogger.Fatal(fmt.Sprintf("Failed to load config: %v", err))
	}

	dbHost := getEnv("DB_HOST", cfg.Database.Host)
	dbPort := getEnvInt("DB_PORT", cfg.Database.Port)
	dbUser := getEnv("DB_USER", cfg.Database.User)
	dbPass := getEnv("DB_PASSWORD", cfg.Database.Password)
	dbName := getEnv("DB_NAME", cfg.Database.DBName)
	dbSSL := getEnv("DB_SSLMODE", cfg.Database.SSLMode)

	db, err := database.NewDB(dbHost, dbPort, dbUser, dbPass, dbName, dbSSL)
	if err != nil {
		svcLogger.Fatal(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	defer db.Close()
	svcLogger.Info("Connected to PostgreSQL database")

	platformCfgRepo := database.NewPostgresPlatformConfigurationRepository(db.Pool)
	tenantCfgRepo := database.NewPostgresTenantConfigurationRepository(db.Pool)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "healthy", "service": "configuration-service", "version": "1.0.0",
		})
	})
	mux.HandleFunc("/api/v1/config", listConfigHandler(platformCfgRepo, tenantCfgRepo))
	mux.HandleFunc("/api/v1/config/create", createConfigHandler(platformCfgRepo))
	mux.HandleFunc("/api/v1/config/", configHandler(platformCfgRepo))
	mux.HandleFunc("/api/v1/tenant-config", listTenantConfigHandler(tenantCfgRepo))
	mux.HandleFunc("/api/v1/tenant-config/create", createTenantConfigHandler(tenantCfgRepo))

	handler := middleware.RequestIDMiddleware(mux)

	port := getEnv("PORT", "8088")
	addr := fmt.Sprintf(":%s", port)
	svcLogger.Info(fmt.Sprintf("Configuration Service listening on %s", addr))

	if err := http.ListenAndServe(addr, handler); err != nil {
		svcLogger.Fatal(fmt.Sprintf("Server failed: %v", err))
	}
}

func listConfigHandler(platformRepo *database.PostgresPlatformConfigurationRepository, tenantRepo *database.PostgresTenantConfigurationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		category := r.URL.Query().Get("category")
		tenantID := r.URL.Query().Get("tenant_id")

		if tenantID != "" {
			tid, err := strconv.ParseInt(tenantID, 10, 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid tenant_id")
				return
			}
			configs, err := tenantRepo.FindByTenantID(ctx, tid)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
				return
			}
			if configs == nil {
				configs = make([]*entities.TenantConfiguration, 0)
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": configs})
			return
		}

		var configs []*entities.PlatformConfiguration
		var err error

		if category != "" {
			configs, err = platformRepo.FindByCategory(ctx, category)
		} else {
			configs, err = platformRepo.List(ctx)
		}

		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}
		if configs == nil {
			configs = make([]*entities.PlatformConfiguration, 0)
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": configs})
	}
}

func createConfigHandler(repo *database.PostgresPlatformConfigurationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
			return
		}

		var cfg entities.PlatformConfiguration
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}

		if err := repo.Create(r.Context(), &cfg); err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{"data": cfg})
	}
}

func configHandler(repo *database.PostgresPlatformConfigurationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Path[len("/api/v1/config/"):]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid configuration ID")
			return
		}

		ctx := r.Context()

		cfg, err := repo.FindByID(ctx, id)
		if err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Configuration not found")
			return
		}

		if r.Method == http.MethodPut {
			var updated entities.PlatformConfiguration
			if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
				writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
				return
			}

			cfg.ConfigValue = updated.ConfigValue
			cfg.Description = updated.Description
			cfg.DataType = updated.DataType

			if err := repo.Update(ctx, cfg); err != nil {
				writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
				return
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{"data": cfg})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{"data": cfg})
	}
}

func listTenantConfigHandler(repo *database.PostgresTenantConfigurationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.URL.Query().Get("tenant_id")
		if tenantID == "" {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "tenant_id is required")
			return
		}

		tid, err := strconv.ParseInt(tenantID, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid tenant_id")
			return
		}

		configs, err := repo.FindByTenantID(r.Context(), tid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}
		if configs == nil {
			configs = make([]*entities.TenantConfiguration, 0)
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": configs})
	}
}

func createTenantConfigHandler(repo *database.PostgresTenantConfigurationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
			return
		}

		var cfg entities.TenantConfiguration
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}

		if err := repo.Create(r.Context(), &cfg); err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{"data": cfg})
	}
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

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
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
