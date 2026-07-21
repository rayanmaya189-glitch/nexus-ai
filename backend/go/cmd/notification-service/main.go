package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Notification struct {
	ID        string                 `json:"id"`
	TenantID  int64                  `json:"tenant_id"`
	UserID    int64                  `json:"user_id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Channel   string                 `json:"channel"`
	Priority  string                 `json:"priority"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Read      bool                   `json:"read"`
	CreatedAt string                 `json:"created_at"`
}

type NotificationPreference struct {
	UserID           int64    `json:"user_id"`
	Channels         []string `json:"channels"`
	Types            []string `json:"types"`
	QuietHoursStart  *int     `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd    *int     `json:"quiet_hours_end,omitempty"`
	Enabled          bool     `json:"enabled"`
}

type NotificationStore struct {
	notifications []Notification
	preferences   map[int64]*NotificationPreference
	mu            sync.RWMutex
}

var store = &NotificationStore{
	notifications: make([]Notification, 0),
	preferences:   make(map[int64]*NotificationPreference),
}

func init() {
	store.notifications = append(store.notifications,
		Notification{ID: "notif-001", TenantID: 1, UserID: 1, Type: "system", Title: "Welcome", Message: "Welcome to AeroXe Nexus AI", Channel: "in_app", Priority: "normal", Read: false, CreatedAt: time.Now().Add(-1 * time.Hour).Format(time.RFC3339)},
		Notification{ID: "notif-002", TenantID: 1, UserID: 1, Type: "security", Title: "New Login", Message: "New login from IP 192.168.1.100", Channel: "email", Priority: "high", Read: false, CreatedAt: time.Now().Add(-30 * time.Minute).Format(time.RFC3339)},
	)
}

func main() {
	log.Println("Starting Notification Service")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/notifications", listHandler)
	mux.HandleFunc("/api/v1/notifications/create", createHandler)
	mux.HandleFunc("/api/v1/notifications/read/", readHandler)
	mux.HandleFunc("/api/v1/notifications/preferences", preferencesHandler)

	port := getEnv("PORT", "8087")
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Notification Service listening on %s", addr)

	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "healthy", "service": "notification-service", "version": "1.0.0",
	})
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": store.notifications})
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	var notif Notification
	if err := json.NewDecoder(r.Body).Decode(&notif); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	notif.ID = fmt.Sprintf("notif-%d", time.Now().UnixNano())
	notif.CreatedAt = time.Now().Format(time.RFC3339)
	notif.Read = false

	store.mu.Lock()
	store.notifications = append(store.notifications, notif)
	store.mu.Unlock()

	writeJSON(w, http.StatusCreated, map[string]interface{}{"data": notif})
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/notifications/read/"):]

	store.mu.Lock()
	defer store.mu.Unlock()

	for i, n := range store.notifications {
		if n.ID == id {
			store.notifications[i].Read = true
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": store.notifications[i]})
			return
		}
	}

	writeError(w, http.StatusNotFound, "NOT_FOUND", "Notification not found")
}

func preferencesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		store.mu.RLock()
		defer store.mu.RUnlock()
		prefs := make([]*NotificationPreference, 0)
		for _, p := range store.preferences {
			prefs = append(prefs, p)
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": prefs})
		return
	}

	if r.Method == http.MethodPost {
		var pref NotificationPreference
		if err := json.NewDecoder(r.Body).Decode(&pref); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
			return
		}
		pref.Enabled = true
		store.mu.Lock()
		store.preferences[pref.UserID] = &pref
		store.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": pref})
		return
	}

	writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET or POST only")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		next.ServeHTTP(w, r)
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

func getEnv(key, def string) string { return def }
