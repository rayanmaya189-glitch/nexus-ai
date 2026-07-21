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

type AuthHandler struct {
	useCase *usecases.AuthUseCase
}

func NewAuthHandler(useCase *usecases.AuthUseCase) *AuthHandler {
	return &AuthHandler{useCase: useCase}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var cmd commands.LoginCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	result, err := h.useCase.Login(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	httputil.WriteSuccess(w, result)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var cmd commands.RegisterCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	user, err := h.useCase.Register(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	httputil.WriteCreated(w, user)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var cmd commands.RefreshTokenCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	result, err := h.useCase.RefreshToken(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	httputil.WriteSuccess(w, result)
}

func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := strconv.ParseInt(r.URL.Query().Get("tenant_id"), 10, 64)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	if page <= 0 { page = 1 }
	if perPage <= 0 { perPage = 20 }

	q := queries.ListUsersQuery{TenantID: tenantID, Page: page, PerPage: perPage}
	users, total, err := h.useCase.ListUsers(r.Context(), q)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	httputil.WritePaginated(w, users, page, perPage, total)
}

func (h *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	q := queries.GetUserQuery{ID: id}
	user, err := h.useCase.GetUser(r.Context(), q)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, user)
}

func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var cmd commands.CreateUserCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	user, err := h.useCase.CreateUser(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteCreated(w, user)
}

func (h *AuthHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var cmd commands.UpdateUserCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	user, err := h.useCase.UpdateUser(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, user)
}

func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/v1/users/"):]
	id, _ := strconv.ParseInt(idStr, 10, 64)

	if err := h.useCase.DeleteUser(r.Context(), id); err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, map[string]bool{"deleted": true})
}

func (h *AuthHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page <= 0 { page = 1 }
	if perPage <= 0 { perPage = 20 }

	q := queries.ListTenantsQuery{Page: page, PerPage: perPage}
	tenants, total, err := h.useCase.ListTenants(r.Context(), q)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WritePaginated(w, tenants, page, perPage, total)
}

func (h *AuthHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var cmd commands.CreateTenantCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	tenant, err := h.useCase.CreateTenant(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteCreated(w, tenant)
}

func (h *AuthHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	var cmd commands.UpdateTenantCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	tenant, err := h.useCase.UpdateTenant(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, tenant)
}

func (h *AuthHandler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/v1/tenants/"):]
	id, _ := strconv.ParseInt(idStr, 10, 64)

	if err := h.useCase.DeleteTenant(r.Context(), id); err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, map[string]bool{"deleted": true})
}

func (h *AuthHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := strconv.ParseInt(r.URL.Query().Get("tenant_id"), 10, 64)
	q := queries.ListRolesQuery{TenantID: tenantID}
	roles, err := h.useCase.ListRoles(r.Context(), q)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, roles)
}

func (h *AuthHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var cmd commands.CreateRoleCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	role, err := h.useCase.CreateRole(r.Context(), cmd)
	if err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteCreated(w, role)
}

func (h *AuthHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/v1/roles/"):]
	id, _ := strconv.ParseInt(idStr, 10, 64)

	if err := h.useCase.DeleteRole(r.Context(), id); err != nil {
		if apiErr, ok := err.(*nexuserrors.APIError); ok {
			apiErr.WriteJSON(w)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.WriteSuccess(w, map[string]bool{"deleted": true})
}
