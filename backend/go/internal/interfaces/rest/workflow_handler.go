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

type WorkflowHandler struct {
	useCase *usecases.WorkflowUseCase
}

func NewWorkflowHandler(useCase *usecases.WorkflowUseCase) *WorkflowHandler {
	return &WorkflowHandler{useCase: useCase}
}

func (h *WorkflowHandler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	q := queries.GetWorkflowQuery{ID: id}
	wf, err := h.useCase.GetWorkflow(r.Context(), q)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, wf)
}

func (h *WorkflowHandler) ListWorkflows(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := strconv.ParseInt(r.URL.Query().Get("tenant_id"), 10, 64)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	q := queries.ListWorkflowsQuery{TenantID: tenantID, Page: page, PerPage: perPage}
	wfs, err := h.useCase.ListWorkflows(r.Context(), q)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, wfs)
}

func (h *WorkflowHandler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	var cmd commands.CreateWorkflowCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	wf, err := h.useCase.CreateWorkflow(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteCreated(w, wf)
}

func (h *WorkflowHandler) UpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	var cmd commands.UpdateWorkflowCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	wf, err := h.useCase.UpdateWorkflow(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, wf)
}

func (h *WorkflowHandler) DeleteWorkflow(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/v1/workflows/"):]
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.useCase.DeleteWorkflow(r.Context(), id); err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, map[string]bool{"deleted": true})
}

func (h *WorkflowHandler) GetWorkflowSteps(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("workflow_id"), 10, 64)
	steps, err := h.useCase.GetWorkflowSteps(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, steps)
}
