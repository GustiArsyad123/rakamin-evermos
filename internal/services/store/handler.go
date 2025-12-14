package store

import (
	"net/http"
	"strconv"

	"database/sql"

	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, dbConn *sql.DB) {
	repo := NewRepo(dbConn)
	uc := NewUsecase(repo)
	r.POST("/api/v1/stores", middleware.GinJWTAuth(), makeCreateHandler(uc))
	r.GET("/api/v1/stores/:id", middleware.GinJWTAuth(), makeGetHandler(uc))
	r.PUT("/api/v1/stores/:id", middleware.GinJWTAuth(), makeUpdateHandler(uc))
	r.DELETE("/api/v1/stores/:id", middleware.GinJWTAuth(), makeDeleteHandler(uc))
}

func makeCreateHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := middleware.GinGetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		var req struct {
			Name string `json:"name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		id, err := uc.CreateStore(uid, req.Name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, map[string]interface{}{"id": id})
	}
}

func makeGetHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := middleware.GinGetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		role, _ := middleware.GinGetRole(c)
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		s, err := uc.GetStore(uid, id, role)
		if err != nil {
			if err.Error() == "forbidden" {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		if s == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, s)
	}
}

func makeUpdateHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := middleware.GinGetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		role, _ := middleware.GinGetRole(c)
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var req struct {
			Name string `json:"name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if err := uc.UpdateStore(uid, id, role, req.Name); err != nil {
			if err.Error() == "forbidden" {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			} else if err.Error() == "not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func makeDeleteHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := middleware.GinGetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		role, _ := middleware.GinGetRole(c)
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		if err := uc.DeleteStore(uid, id, role); err != nil {
			if err.Error() == "forbidden" {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			} else if err.Error() == "not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.Status(http.StatusNoContent)
	}
}
