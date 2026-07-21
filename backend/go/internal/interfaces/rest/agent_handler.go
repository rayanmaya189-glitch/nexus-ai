package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aeroxe/nexus-backend/internal/application/commands"
	"github.com/aeroxe/nexus-backend/internal/application/queries"
	"github.com/aeroxe/nexus-backend/internal/application/usecases"
	nexuserrors "github.com/aeroxe/nexus-backend/pkg/errors"
	"github.com/aeroxe/nexus-backend/pkg/httputil"
)

type AgentHandler struct {
	useCase *usecases.AgentUseCase
}

func NewAgentHandler(useCase *usecases.AgentUseCase) *AgentHandler {
	return &AgentHandler{useCase: useCase}
}

func (h *AgentHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	q := queries.GetAgentQuery{ID: id}
	agent, err := h.useCase.GetAgent(r.Context(), q)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, agent)
}

func (h *AgentHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := strconv.ParseInt(r.URL.Query().Get("tenant_id"), 10, 64)
	agentType := r.URL.Query().Get("agent_type")
	q := queries.ListAgentsQuery{TenantID: tenantID, AgentType: agentType}
	agents, err := h.useCase.ListAgents(r.Context(), q)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, agents)
}

func (h *AgentHandler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	var cmd commands.CreateAgentCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	agent, err := h.useCase.CreateAgent(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteCreated(w, agent)
}

func (h *AgentHandler) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	var cmd commands.UpdateAgentCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	agent, err := h.useCase.UpdateAgent(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, agent)
}

func (h *AgentHandler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/v1/agents/"):]
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.useCase.DeleteAgent(r.Context(), id); err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, map[string]bool{"deleted": true})
}

func (h *AgentHandler) GetExecution(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	q := queries.GetAgentExecutionQuery{ID: id}
	exec, err := h.useCase.GetExecution(r.Context(), q)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, exec)
}

func (h *AgentHandler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	agentID, _ := strconv.ParseInt(r.URL.Query().Get("agent_id"), 10, 64)
	q := queries.ListAgentExecutionsQuery{AgentID: agentID}
	execs, err := h.useCase.ListExecutions(r.Context(), q)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, execs)
}
