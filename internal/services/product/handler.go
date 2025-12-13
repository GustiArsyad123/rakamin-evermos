package product

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"database/sql"

	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/example/ms-ecommerce/internal/pkg/models"
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, dbConn *sql.DB) {
	repo := NewRepo(dbConn)
	uc := NewUsecase(repo)
	// create product requires authentication
	r.Handle("/api/v1/products", middleware.JWTAuth(makeCreateHandler(uc))).Methods("POST")
	r.Handle("/api/v1/products", middleware.JWTAuth(makeListHandler(uc))).Methods("GET")
	r.Handle("/api/v1/products/{id}", middleware.JWTAuth(makeGetHandler(uc))).Methods("GET")
}

func makeCreateHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// user id from context (set by middleware)
		uid, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		name := r.FormValue("name")
		desc := r.FormValue("description")
		priceStr := r.FormValue("price")
		stockStr := r.FormValue("stock")
		catStr := r.FormValue("category_id")
		if name == "" || priceStr == "" {
			http.Error(w, "missing fields", http.StatusBadRequest)
			return
		}
		price, _ := strconv.ParseFloat(priceStr, 64)
		stock, _ := strconv.Atoi(stockStr)
		var cat *int64
		if catStr != "" {
			v, _ := strconv.ParseInt(catStr, 10, 64)
			cat = &v
		}

		var imageURL string
		file, fh, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			os.MkdirAll("uploads", 0755)
			filename := fmt.Sprintf("%d_%s", (int64)(uid), fh.Filename)
			dst := filepath.Join("uploads", filename)
			out, err := os.Create(dst)
			if err == nil {
				defer out.Close()
				io.Copy(out, file)
				imageURL = dst
			}
		}

		p := &models.Product{Name: name, Description: desc, Price: price, Stock: stock, CategoryID: cat, ImageURL: imageURL}
		id, err := uc.CreateProduct(uid, p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	}
}

func makeListHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// user id from context (set by middleware)
		uid, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		q := r.URL.Query()
		filters := map[string]string{}
		if v := q.Get("search"); v != "" {
			filters["search"] = v
		}
		if v := q.Get("category_id"); v != "" {
			filters["category_id"] = v
		}
		if v := q.Get("min_price"); v != "" {
			filters["min_price"] = v
		}
		if v := q.Get("max_price"); v != "" {
			filters["max_price"] = v
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

		data, total, err := uc.ListProducts(uid, filters, page, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp := map[string]interface{}{
			"data":       data,
			"pagination": map[string]interface{}{"page": page, "limit": limit, "total": total},
		}
		json.NewEncoder(w).Encode(resp)
	}
}

func makeGetHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// user id from context (set by middleware)
		uid, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		idStr := vars["id"]
		id, _ := strconv.ParseInt(idStr, 10, 64)
		p, err := uc.GetProduct(uid, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if p == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(p)
	}
}
