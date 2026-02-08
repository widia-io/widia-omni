package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/widia-io/widia-omni/internal/domain"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type FinanceHandler struct {
	financeSvc *service.FinanceService
}

func NewFinanceHandler(financeSvc *service.FinanceService) *FinanceHandler {
	return &FinanceHandler{financeSvc: financeSvc}
}

func (h *FinanceHandler) financeGate(w http.ResponseWriter, r *http.Request) (uuid.UUID, *domain.EntitlementLimits, bool) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return uuid.Nil, nil, false
	}
	limits, ok := middleware.GetEntitlements(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "entitlements not loaded")
		return uuid.Nil, nil, false
	}
	if !limits.FinanceEnabled {
		writeError(w, http.StatusForbidden, "finance not available on your plan")
		return uuid.Nil, nil, false
	}
	return wsID, limits, true
}

// --- Summary ---

func (h *FinanceHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	wsID, _, ok := h.financeGate(w, r)
	if !ok {
		return
	}

	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	summary, err := h.financeSvc.GetSummary(r.Context(), wsID, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get summary")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// --- Transactions ---

func (h *FinanceHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	wsID, _, ok := h.financeGate(w, r)
	if !ok {
		return
	}

	var f service.TransactionFilters
	if v := r.URL.Query().Get("type"); v != "" {
		f.Type = &v
	}
	if v := r.URL.Query().Get("category_id"); v != "" {
		id, err := uuid.Parse(v)
		if err == nil {
			f.CategoryID = &id
		}
	}
	if v := r.URL.Query().Get("area_id"); v != "" {
		id, err := uuid.Parse(v)
		if err == nil {
			f.AreaID = &id
		}
	}
	if v := r.URL.Query().Get("date_from"); v != "" {
		f.DateFrom = &v
	}
	if v := r.URL.Query().Get("date_to"); v != "" {
		f.DateTo = &v
	}
	if v := r.URL.Query().Get("tag"); v != "" {
		f.Tag = &v
	}
	f.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	f.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))

	txns, err := h.financeSvc.ListTransactions(r.Context(), wsID, f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list transactions")
		return
	}
	writeJSON(w, http.StatusOK, txns)
}

func (h *FinanceHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	wsID, limits, ok := h.financeGate(w, r)
	if !ok {
		return
	}

	var req service.CreateTransactionRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	txn, err := h.financeSvc.CreateTransaction(r.Context(), wsID, limits, req)
	if err != nil {
		if err.Error() == "monthly transaction limit reached" {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create transaction")
		return
	}
	writeJSON(w, http.StatusCreated, txn)
}

func (h *FinanceHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	wsID, _, ok := h.financeGate(w, r)
	if !ok {
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	var req service.UpdateTransactionRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	txn, err := h.financeSvc.UpdateTransaction(r.Context(), wsID, id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update transaction")
		return
	}
	writeJSON(w, http.StatusOK, txn)
}

func (h *FinanceHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	wsID, _, ok := h.financeGate(w, r)
	if !ok {
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	if err := h.financeSvc.DeleteTransaction(r.Context(), wsID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete transaction")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Categories ---

func (h *FinanceHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	wsID, _, ok := h.financeGate(w, r)
	if !ok {
		return
	}

	cats, err := h.financeSvc.ListCategories(r.Context(), wsID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}
	writeJSON(w, http.StatusOK, cats)
}

func (h *FinanceHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	wsID, _, ok := h.financeGate(w, r)
	if !ok {
		return
	}

	var req service.CreateCategoryRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.financeSvc.CreateCategory(r.Context(), wsID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create category")
		return
	}
	writeJSON(w, http.StatusCreated, cat)
}

func (h *FinanceHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	wsID, _, ok := h.financeGate(w, r)
	if !ok {
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	var req service.UpdateCategoryRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.financeSvc.UpdateCategory(r.Context(), wsID, id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update category")
		return
	}
	writeJSON(w, http.StatusOK, cat)
}

func (h *FinanceHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	wsID, _, ok := h.financeGate(w, r)
	if !ok {
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	if err := h.financeSvc.DeleteCategory(r.Context(), wsID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete category")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Budgets ---

func (h *FinanceHandler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	wsID, _, ok := h.financeGate(w, r)
	if !ok {
		return
	}

	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	budgets, err := h.financeSvc.ListBudgets(r.Context(), wsID, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list budgets")
		return
	}
	writeJSON(w, http.StatusOK, budgets)
}

func (h *FinanceHandler) UpsertBudget(w http.ResponseWriter, r *http.Request) {
	wsID, _, ok := h.financeGate(w, r)
	if !ok {
		return
	}

	var req service.UpsertBudgetRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	budget, err := h.financeSvc.UpsertBudget(r.Context(), wsID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to upsert budget")
		return
	}
	writeJSON(w, http.StatusCreated, budget)
}

func (h *FinanceHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	wsID, _, ok := h.financeGate(w, r)
	if !ok {
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget id")
		return
	}

	if err := h.financeSvc.DeleteBudget(r.Context(), wsID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete budget")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
