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

type AuditHandler struct {
	useCase *usecases.AuditUseCase
}

func NewAuditHandler(useCase *usecases.AuditUseCase) *AuditHandler {
	return &AuditHandler{useCase: useCase}
}

func (h *AuditHandler) CreateAuditLog(w http.ResponseWriter, r *http.Request) {
	var cmd commands.CreateAuditLogCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	log, err := h.useCase.CreateAuditLog(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteCreated(w, log)
}

func (h *AuditHandler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	q := queries.GetAuditLogQuery{ID: id}
	log, err := h.useCase.GetAuditLog(r.Context(), q)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, log)
}

func (h *AuditHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := strconv.ParseInt(r.URL.Query().Get("tenant_id"), 10, 64)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	q := queries.ListAuditLogsQuery{TenantID: tenantID, Page: page, PerPage: perPage}
	logs, total, err := h.useCase.ListAuditLogs(r.Context(), q)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WritePaginated(w, logs, page, perPage, total)
}
