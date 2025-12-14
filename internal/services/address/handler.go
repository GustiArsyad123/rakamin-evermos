package address

import (
	"net/http"
	"strconv"

	"database/sql"

	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/example/ms-ecommerce/internal/pkg/models"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, dbConn *sql.DB) {
	repo := NewRepo(dbConn)
	uc := NewUsecase(repo)
	r.POST("/api/v1/addresses", middleware.GinJWTAuth(), makeCreateHandler(uc))
	r.GET("/api/v1/addresses", middleware.GinJWTAuth(), makeListHandler(uc))
	r.GET("/api/v1/addresses/:id", middleware.GinJWTAuth(), makeGetHandler(uc))
	r.PUT("/api/v1/addresses/:id", middleware.GinJWTAuth(), makeUpdateHandler(uc))
	r.DELETE("/api/v1/addresses/:id", middleware.GinJWTAuth(), makeDeleteHandler(uc))
}

func makeCreateHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := middleware.GinGetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		var req struct {
			Label      string `json:"label"`
			Address    string `json:"address"`
			City       string `json:"city"`
			PostalCode string `json:"postal_code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		a := &models.Address{Label: req.Label, Address: req.Address, City: req.City, PostalCode: req.PostalCode}
		id, err := uc.CreateAddress(uid, a)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, map[string]interface{}{"id": id})
	}
}

func makeListHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := middleware.GinGetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		filters := map[string]string{}
		if v := c.Query("label"); v != "" {
			filters["label"] = v
		}
		if v := c.Query("city"); v != "" {
			filters["city"] = v
		}
		if v := c.Query("postal_code"); v != "" {
			filters["postal_code"] = v
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

		data, total, err := uc.ListAddresses(uid, filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, map[string]interface{}{"data": data, "pagination": map[string]interface{}{"page": page, "limit": limit, "total": total}})
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
		a, err := uc.GetAddress(uid, id, role)
		if err != nil {
			if err.Error() == "forbidden" {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		if a == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, a)
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
			Label      string `json:"label"`
			Address    string `json:"address"`
			City       string `json:"city"`
			PostalCode string `json:"postal_code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		err = uc.UpdateAddress(uid, id, role, req.Label, req.Address, req.City, req.PostalCode)
		if err != nil {
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
		err = uc.DeleteAddress(uid, id, role)
		if err != nil {
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
