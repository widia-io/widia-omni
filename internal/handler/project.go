package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type ProjectHandler struct {
	projectSvc *service.ProjectService
}

func NewProjectHandler(projectSvc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectSvc: projectSvc}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	var f service.ProjectFilters
	if v := r.URL.Query().Get("area_id"); v != "" {
		id, err := uuid.Parse(v)
		if err == nil {
			f.AreaID = &id
		}
	}
	if v := r.URL.Query().Get("goal_id"); v != "" {
		id, err := uuid.Parse(v)
		if err == nil {
			f.GoalID = &id
		}
	}
	if v := r.URL.Query().Get("status"); v != "" {
		f.Status = &v
	}
	if r.URL.Query().Get("include_archived") == "true" {
		f.IncludeArchived = true
	}

	projects, err := h.projectSvc.List(r.Context(), wsID, f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list projects")
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}
	limits, ok := middleware.GetEntitlements(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "entitlements not loaded")
		return
	}

	var req service.CreateProjectRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	project, err := h.projectSvc.Create(r.Context(), wsID, limits, req)
	if err != nil {
		if err.Error() == "project limit reached" {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create project")
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	project, err := h.projectSvc.GetByID(r.Context(), wsID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get project")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var req service.UpdateProjectRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	project, err := h.projectSvc.Update(r.Context(), wsID, id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update project")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	if err := h.projectSvc.Delete(r.Context(), wsID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete project")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *ProjectHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var req struct {
		SortOrder int `json:"sort_order"`
	}
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.projectSvc.Reorder(r.Context(), wsID, id, req.SortOrder); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder project")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "reordered"})
}

func (h *ProjectHandler) Archive(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	project, err := h.projectSvc.Archive(r.Context(), wsID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to archive project")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Unarchive(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	project, err := h.projectSvc.Unarchive(r.Context(), wsID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to unarchive project")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

// --- Project Sections ---

func (h *ProjectHandler) ListSections(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	sections, err := h.projectSvc.ListSections(r.Context(), wsID, projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list sections")
		return
	}
	writeJSON(w, http.StatusOK, sections)
}

func (h *ProjectHandler) CreateSection(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var req service.CreateProjectSectionRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	section, err := h.projectSvc.CreateSection(r.Context(), wsID, projectID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create section")
		return
	}
	writeJSON(w, http.StatusCreated, section)
}

func (h *ProjectHandler) UpdateSection(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	sectionID, err := uuid.Parse(chi.URLParam(r, "sectionId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid section id")
		return
	}

	var req service.UpdateProjectSectionRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	section, err := h.projectSvc.UpdateSection(r.Context(), wsID, projectID, sectionID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update section")
		return
	}
	writeJSON(w, http.StatusOK, section)
}

func (h *ProjectHandler) DeleteSection(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	sectionID, err := uuid.Parse(chi.URLParam(r, "sectionId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid section id")
		return
	}

	if err := h.projectSvc.DeleteSection(r.Context(), wsID, projectID, sectionID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete section")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *ProjectHandler) ReorderSection(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	sectionID, err := uuid.Parse(chi.URLParam(r, "sectionId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid section id")
		return
	}

	var req struct {
		Position int `json:"position"`
	}
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.projectSvc.ReorderSection(r.Context(), wsID, projectID, sectionID, req.Position); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder section")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "reordered"})
}
