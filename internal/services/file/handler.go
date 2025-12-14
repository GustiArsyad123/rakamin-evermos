package file

import (
	"net/http"

	"database/sql"

	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, dbConn *sql.DB) {
	repo := NewRepo(dbConn)
	uc := NewUsecase(repo)
	r.POST("/api/v1/files/upload", middleware.GinJWTAuth(), makeUploadHandler(uc))
	// Serve uploaded files
	r.StaticFS("/uploads", http.Dir("uploads"))
}

func makeUploadHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from JWT
		uid, ok := middleware.GinGetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// Parse multipart form (max 10MB)
		if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form data"})
			return
		}

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
			return
		}
		defer file.Close()

		// Upload file
		fileURL, err := uc.UploadFile(uid, file, header)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Return success response
		c.JSON(http.StatusOK, map[string]interface{}{
			"url":      fileURL,
			"filename": header.Filename,
			"size":     header.Size,
		})
	}
}
