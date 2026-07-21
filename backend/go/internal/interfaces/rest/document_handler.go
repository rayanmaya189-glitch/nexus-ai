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

type DocumentHandler struct {
	useCase *usecases.DocumentUseCase
}

func NewDocumentHandler(useCase *usecases.DocumentUseCase) *DocumentHandler {
	return &DocumentHandler{useCase: useCase}
}

func (h *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	q := queries.GetDocumentQuery{ID: id}
	doc, err := h.useCase.GetDocument(r.Context(), q)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, doc)
}

func (h *DocumentHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := strconv.ParseInt(r.URL.Query().Get("tenant_id"), 10, 64)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	q := queries.ListDocumentsQuery{TenantID: tenantID, Page: page, PerPage: perPage}
	docs, err := h.useCase.ListDocuments(r.Context(), q)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, docs)
}

func (h *DocumentHandler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	var cmd commands.CreateDocumentCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	doc, err := h.useCase.CreateDocument(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteCreated(w, doc)
}

func (h *DocumentHandler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	var cmd commands.UpdateDocumentCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	doc, err := h.useCase.UpdateDocument(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, doc)
}

func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/v1/rag/documents/"):]
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.useCase.DeleteDocument(r.Context(), id); err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, map[string]bool{"deleted": true})
}

func (h *DocumentHandler) CreateDocumentSet(w http.ResponseWriter, r *http.Request) {
	var cmd commands.CreateDocumentSetCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	set, err := h.useCase.CreateDocumentSet(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteCreated(w, set)
}

func (h *DocumentHandler) ListDocumentSets(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := strconv.ParseInt(r.URL.Query().Get("tenant_id"), 10, 64)
	q := queries.ListDocumentSetsQuery{TenantID: tenantID}
	sets, err := h.useCase.ListDocumentSets(r.Context(), q)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, sets)
}
