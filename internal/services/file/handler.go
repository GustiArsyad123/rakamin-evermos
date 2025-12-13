package file

import (
	"encoding/json"
	"net/http"

	"database/sql"

	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, dbConn *sql.DB) {
	repo := NewRepo(dbConn)
	uc := NewUsecase(repo)
	r.Handle("/api/v1/files/upload", middleware.JWTAuth(makeUploadHandler(uc))).Methods("POST")
	// Serve uploaded files
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads/"))))
}

func makeUploadHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from JWT
		uid, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse multipart form (max 10MB)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "invalid form data", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "no file provided", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Upload file
		fileURL, err := uc.UploadFile(uid, file, header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"url":      fileURL,
			"filename": header.Filename,
			"size":     header.Size,
		})
	}
}
