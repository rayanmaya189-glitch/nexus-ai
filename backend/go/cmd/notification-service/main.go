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
	svcLogger := logger.New("notification-service")
	svcLogger.Info("Starting Notification Service")

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

	notifRepo := database.NewPostgresNotificationRepository(db.Pool)
	prefRepo := database.NewPostgresNotificationPreferenceRepository(db.Pool)
	tmplRepo := database.NewPostgresNotificationTemplateRepository(db.Pool)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "healthy", "service": "notification-service", "version": "1.0.0",
		})
	})
	mux.HandleFunc("/api/v1/notifications", listNotificationsHandler(notifRepo))
	mux.HandleFunc("/api/v1/notifications/create", createNotificationHandler(notifRepo))
	mux.HandleFunc("/api/v1/notifications/preferences", preferencesHandler(prefRepo))
	mux.HandleFunc("/api/v1/notifications/read/", readNotificationHandler(notifRepo))
	mux.HandleFunc("/api/v1/notifications/templates", listTemplatesHandler(tmplRepo))
	mux.HandleFunc("/api/v1/notifications/unread-count", unreadCountHandler(notifRepo))

	handler := middleware.RequestIDMiddleware(mux)

	port := getEnv("PORT", "8087")
	addr := fmt.Sprintf(":%s", port)
	svcLogger.Info(fmt.Sprintf("Notification Service listening on %s", addr))

	if err := http.ListenAndServe(addr, handler); err != nil {
		svcLogger.Fatal(fmt.Sprintf("Server failed: %v", err))
	}
}

func listNotificationsHandler(repo *database.PostgresNotificationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tenantID := r.URL.Query().Get("tenant_id")
		userID := r.URL.Query().Get("user_id")

		var notifications []*entities.Notification
		var err error

		switch {
		case userID != "":
			uid, e := strconv.ParseInt(userID, 10, 64)
			if e != nil {
				writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid user_id")
				return
			}
			notifications, err = repo.FindByUserID(ctx, uid)
		case tenantID != "":
			tid, e := strconv.ParseInt(tenantID, 10, 64)
			if e != nil {
				writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid tenant_id")
				return
			}
			notifications, err = repo.FindByTenantID(ctx, tid)
		default:
			notifications = make([]*entities.Notification, 0)
		}

		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}
		if notifications == nil {
			notifications = make([]*entities.Notification, 0)
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": notifications})
	}
}

func createNotificationHandler(repo *database.PostgresNotificationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
			return
		}

		var notif entities.Notification
		if err := json.NewDecoder(r.Body).Decode(&notif); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}
		notif.Status = "unread"
		notif.IsRead = false

		if err := repo.Create(r.Context(), &notif); err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{"data": notif})
	}
}

func readNotificationHandler(repo *database.PostgresNotificationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Path[len("/api/v1/notifications/read/"):]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid notification ID")
			return
		}

		if err := repo.MarkAsRead(r.Context(), id); err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Notification not found")
			return
		}

		notif, err := repo.FindByID(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Notification not found")
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{"data": notif})
	}
}

func preferencesHandler(repo *database.PostgresNotificationPreferenceRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method == http.MethodGet {
			tenantID := r.URL.Query().Get("tenant_id")
			userID := r.URL.Query().Get("user_id")

			if tenantID != "" && userID != "" {
				tid, _ := strconv.ParseInt(tenantID, 10, 64)
				uid, _ := strconv.ParseInt(userID, 10, 64)
				pref, err := repo.FindByTenantAndUser(ctx, tid, uid)
				if err != nil {
					writeJSON(w, http.StatusOK, map[string]interface{}{"data": map[string]interface{}{}})
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"data": pref})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": []interface{}{}})
			return
		}

		if r.Method == http.MethodPost {
			var pref entities.NotificationPreference
			if err := json.NewDecoder(r.Body).Decode(&pref); err != nil {
				writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
				return
			}
			pref.EmailEnabled = true
			pref.PushEnabled = true

			if err := repo.Create(ctx, &pref); err != nil {
				writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": pref})
			return
		}

		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET or POST only")
	}
}

func listTemplatesHandler(repo *database.PostgresNotificationTemplateRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.URL.Query().Get("tenant_id")
		if tenantID == "" {
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": []interface{}{}})
			return
		}

		tid, err := strconv.ParseInt(tenantID, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid tenant_id")
			return
		}

		templates, err := repo.FindByTenantID(r.Context(), tid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}
		if templates == nil {
			templates = make([]*entities.NotificationTemplate, 0)
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": templates})
	}
}

func unreadCountHandler(repo *database.PostgresNotificationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "user_id is required")
			return
		}

		uid, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid user_id")
			return
		}

		count, err := repo.GetUnreadCount(r.Context(), uid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{"data": map[string]interface{}{"count": count}})
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
