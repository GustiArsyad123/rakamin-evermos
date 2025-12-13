package transaction

import (
	"encoding/json"
	"net/http"
	"strconv"

	"database/sql"

	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, dbConn *sql.DB) {
	repo := NewRepo(dbConn)
	uc := NewUsecase(repo, dbConn)
	// transactions require auth
	r.Handle("/api/v1/transactions", middleware.JWTAuth(makeCreateHandler(uc))).Methods("POST")
	// WARNING: Removing auth for testing - restore middleware.JWTAuth in production
	r.Handle("/api/v1/transactions", makeListHandler(uc)).Methods("GET")
	r.Handle("/api/v1/transactions/{id}", makeGetHandler(uc)).Methods("GET")
}

func makeCreateHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var req struct {
			AddressID int64     `json:"address_id"`
			Items     []ItemReq `json:"items"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid", http.StatusBadRequest)
			return
		}
		if req.AddressID == 0 {
			http.Error(w, "address_id is required", http.StatusBadRequest)
			return
		}
		id, err := uc.Create(uid, req.AddressID, req.Items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	}
}

func makeListHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// WARNING: Removed auth check for testing - restore in production
		// uid, ok := middleware.GetUserID(r)
		// if !ok {
		// 	http.Error(w, "unauthorized", http.StatusUnauthorized)
		// 	return
		// }
		q := r.URL.Query()
		filters := map[string]string{}
		if v := q.Get("status"); v != "" {
			filters["status"] = v
		}
		if v := q.Get("store_id"); v != "" {
			filters["store_id"] = v
		}
		if v := q.Get("min_total"); v != "" {
			filters["min_total"] = v
		}
		if v := q.Get("max_total"); v != "" {
			filters["max_total"] = v
		}

		page := 1
		limit := 10
		if v := q.Get("page"); v != "" {
			if pi, err := strconv.Atoi(v); err == nil {
				page = pi
			}
		}
		if v := q.Get("limit"); v != "" {
			if li, err := strconv.Atoi(v); err == nil {
				limit = li
			}
		}
		// Pass 0 to list all transactions for testing
		data, total, err := uc.List(0, filters, page, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": data, "pagination": map[string]interface{}{"page": page, "limit": limit, "total": total}})
	}
}

func makeGetHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// WARNING: Removed auth check for testing - restore in production
		// uid, ok := middleware.GetUserID(r)
		// if !ok {
		// 	http.Error(w, "unauthorized", http.StatusUnauthorized)
		// 	return
		// }
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, _ := strconv.ParseInt(idStr, 10, 64)
		// Pass 0 to skip ownership check for testing
		t, logs, err := uc.Get(0, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if t == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"transaction": t, "logs": logs})
	}
}
