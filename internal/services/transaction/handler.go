package transaction

import (
	"log"
	"net/http"
	"strconv"

	"database/sql"

	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/example/ms-ecommerce/internal/pkg/payment"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, dbConn *sql.DB) {
	repo := NewRepo(dbConn)
	provider := payment.NewProviderFromEnv()
	uc := NewUsecase(repo, dbConn, provider)
	// transactions require auth
	r.POST("/api/v1/transactions", middleware.GinJWTAuth(), makeCreateHandler(uc))
	r.GET("/api/v1/transactions", middleware.GinJWTAuth(), makeListHandler(uc))
	r.GET("/api/v1/transactions/:id", middleware.GinJWTAuth(), makeGetHandler(uc))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})
}

func makeCreateHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := middleware.GinGetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		var req struct {
			AddressID     int64     `json:"address_id"`
			Items         []ItemReq `json:"items"`
			PaymentMethod string    `json:"payment_method"` // e.g. "card" or provider-specific
			PaymentToken  string    `json:"payment_token"`  // token / nonce from frontend
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
			return
		}
		if req.AddressID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "address_id is required"})
			return
		}
		// If payment details provided, attempt to create and charge
		var id int64
		var err error
		if req.PaymentToken != "" {
			id, err = uc.CreateAndCharge(uid, req.AddressID, req.Items, req.PaymentMethod, req.PaymentToken)
		} else {
			id, err = uc.Create(uid, req.AddressID, req.Items)
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, map[string]interface{}{"id": id})
	}
}

func makeListHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := middleware.GinGetUserID(c)
		log.Printf("Transaction list handler: uid=%v, ok=%v", uid, ok)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.String(200, "ok")
		return
		role, _ := middleware.GinGetRole(c)
		filters := map[string]string{}
		if v := c.Query("status"); v != "" {
			filters["status"] = v
		}
		if v := c.Query("store_id"); v != "" {
			filters["store_id"] = v
		}
		if v := c.Query("min_total"); v != "" {
			filters["min_total"] = v
		}
		if v := c.Query("max_total"); v != "" {
			filters["max_total"] = v
		}

		page := 1
		limit := 10
		if v := c.Query("page"); v != "" {
			if pi, err := strconv.Atoi(v); err == nil {
				page = pi
			}
		}
		if v := c.Query("limit"); v != "" {
			if li, err := strconv.Atoi(v); err == nil {
				limit = li
			}
		}
		data, total, err := uc.List(uid, role, filters, page, limit)
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
		id, _ := strconv.ParseInt(idStr, 10, 64)
		t, logs, err := uc.Get(uid, id, role)
		if err != nil {
			if err.Error() == "forbidden" {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		if t == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, map[string]interface{}{"transaction": t, "logs": logs})
	}
}
