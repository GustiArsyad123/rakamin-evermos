package category

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, dbConn *sql.DB) {
	repo := NewRepo(dbConn)
	uc := NewUsecase(repo)

	// Public: list and get
	r.GET("/api/v1/categories", makeListHandler(uc))
	r.GET("/api/v1/categories/:id", makeGetHandler(uc))

	// Admin-only management
	r.POST("/api/v1/categories", middleware.GinJWTAuth(), middleware.GinRequireRole("admin"), makeCreateHandler(uc))
	r.PUT("/api/v1/categories/:id", middleware.GinJWTAuth(), middleware.GinRequireRole("admin"), makeUpdateHandler(uc))
	r.DELETE("/api/v1/categories/:id", middleware.GinJWTAuth(), middleware.GinRequireRole("admin"), makeDeleteHandler(uc))
}

func makeCreateHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name string `json:"name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		id, err := uc.Create(req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": id})
	}
}

func makeUpdateHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		var req struct {
			Name string `json:"name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if err := uc.Update(id, req.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func makeDeleteHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		if err := uc.Delete(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func makeListHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		filters := map[string]string{}
		if v := c.Query("search"); v != "" {
			filters["search"] = v
		}

		page := 1
		limit := 10
		if v := c.Query("page"); v != "" {
			if pi, err := strconv.Atoi(v); err == nil && pi > 0 {
				page = pi
			}
		}
		if v := c.Query("limit"); v != "" {
			if li, err := strconv.Atoi(v); err == nil && li > 0 {
				limit = li
			}
		}

		data, total, err := uc.List(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resp := gin.H{
			"data":       data,
			"pagination": gin.H{"page": page, "limit": limit, "total": total},
		}
		c.JSON(http.StatusOK, resp)
	}
}

func makeGetHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		cat, err := uc.Get(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if cat == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, cat)
	}
}
