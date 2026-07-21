package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type ServiceConfig struct {
	Name    string
	BaseURL string
}

var services = map[string]ServiceConfig{
	"identity":     {Name: "identity-service", BaseURL: getEnv("IDENTITY_SERVICE_URL", "http://localhost:8081")},
	"ai-gateway":   {Name: "ai-gateway", BaseURL: getEnv("AI_GATEWAY_URL", "http://localhost:8082")},
	"agent":        {Name: "agent-orchestrator", BaseURL: getEnv("AGENT_ORCHESTRATOR_URL", "http://localhost:8091")},
	"rag":          {Name: "rag-service", BaseURL: getEnv("RAG_SERVICE_URL", "http://localhost:8092")},
	"vision":       {Name: "vision-service", BaseURL: getEnv("VISION_SERVICE_URL", "http://localhost:8093")},
	"memory":       {Name: "memory-service", BaseURL: getEnv("MEMORY_SERVICE_URL", "http://localhost:8094")},
	"security":     {Name: "security-ai", BaseURL: getEnv("SECURITY_AI_URL", "http://localhost:8095")},
	"workflow":     {Name: "workflow-service", BaseURL: getEnv("WORKFLOW_SERVICE_URL", "http://localhost:8084")},
	"audit":        {Name: "audit-service", BaseURL: getEnv("AUDIT_SERVICE_URL", "http://localhost:8085")},
	"model":        {Name: "model-registry", BaseURL: getEnv("MODEL_REGISTRY_URL", "http://localhost:8086")},
	"notification": {Name: "notification-service", BaseURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8087")},
	"config":       {Name: "configuration-service", BaseURL: getEnv("CONFIGURATION_SERVICE_URL", "http://localhost:8088")},
	"ecosystem":    {Name: "ecosystem-service", BaseURL: getEnv("ECOSYSTEM_SERVICE_URL", "http://localhost:8089")},
	"sql-agent":    {Name: "sql-agent-service", BaseURL: getEnv("SQL_AGENT_SERVICE_URL", "http://localhost:8090")},
}

func main() {
	log.Println("Starting API Gateway")

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"api-gateway","version":"1.0.0","services":%d}`, len(services))
	})

	mux.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		serviceKey := resolveService(path)
		if serviceKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"error":{"code":"NOT_FOUND","message":"No service found for path: %s"}}`, path)
			return
		}

		svc := services[serviceKey]
		proxyToService(w, r, svc.BaseURL)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"service":"api-gateway","version":"1.0.0","available_services":[%s]}`, getServiceList())
	})

	handler := corsMiddleware(mux)

	port := getEnv("PORT", "8000")
	addr := fmt.Sprintf(":%s", port)
	log.Printf("API Gateway listening on %s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func resolveService(path string) string {
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.TrimPrefix(path, "/")

	if strings.HasPrefix(path, "auth") || strings.HasPrefix(path, "users") || strings.HasPrefix(path, "tenants") || strings.HasPrefix(path, "roles") {
		return "identity"
	}
	if strings.HasPrefix(path, "ai") {
		return "ai-gateway"
	}
	if strings.HasPrefix(path, "agents") {
		return "agent"
	}
	if strings.HasPrefix(path, "rag") {
		return "rag"
	}
	if strings.HasPrefix(path, "vision") {
		return "vision"
	}
	if strings.HasPrefix(path, "memory") {
		return "memory"
	}
	if strings.HasPrefix(path, "security") {
		return "security"
	}
	if strings.HasPrefix(path, "workflows") {
		return "workflow"
	}
	if strings.HasPrefix(path, "audit") {
		return "audit"
	}
	if strings.HasPrefix(path, "models") {
		return "model"
	}
	if strings.HasPrefix(path, "notifications") {
		return "notification"
	}
	if strings.HasPrefix(path, "config") {
		return "config"
	}
	if strings.HasPrefix(path, "integrations") || strings.HasPrefix(path, "mcp") {
		return "ecosystem"
	}
	if strings.HasPrefix(path, "sql") {
		return "sql-agent"
	}

	return ""
}

func proxyToService(w http.ResponseWriter, r *http.Request, targetURL string) {
	target, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Bad gateway", http.StatusBadGateway)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, `{"error":{"code":"BAD_GATEWAY","message":"Service unavailable"}}`)
	}

	proxy.ServeHTTP(w, r)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getServiceList() string {
	keys := make([]string, 0, len(services))
	for k := range services {
		keys = append(keys, fmt.Sprintf(`"%s"`, k))
	}
	return strings.Join(keys, ",")
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
