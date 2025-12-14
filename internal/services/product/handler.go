package product

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"database/sql"

	"github.com/example/ms-ecommerce/internal/pkg/cache"
	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/example/ms-ecommerce/internal/pkg/models"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, dbConn *sql.DB, productCache *cache.ProductCache) {
	repo := NewRepo(dbConn, productCache)
	uc := NewUsecase(repo)
	// create product requires authentication
	r.POST("/api/v1/products", middleware.GinJWTAuth(), makeCreateHandler(uc))
	r.GET("/api/v1/products", middleware.GinJWTAuth(), makeListHandler(uc))
	r.GET("/api/v1/products/:id", middleware.GinJWTAuth(), makeGetHandler(uc))
	r.PUT("/api/v1/products/:id", middleware.GinJWTAuth(), makeUpdateHandler(uc))
	r.DELETE("/api/v1/products/:id", middleware.GinJWTAuth(), makeDeleteHandler(uc))
}

func makeCreateHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// user id from context (set by middleware)
		uid, ok := middleware.GinGetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		role, _ := middleware.GinGetRole(c)

		if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form"})
			return
		}
		name := c.Request.FormValue("name")
		desc := c.Request.FormValue("description")
		priceStr := c.Request.FormValue("price")
		stockStr := c.Request.FormValue("stock")
		catStr := c.Request.FormValue("category_id")
		if name == "" || priceStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields"})
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
		file, fh, err := c.Request.FormFile("image")
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
		id, err := uc.CreateProduct(uid, role, p)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Fetch the created product to return complete data
		createdProduct, err := uc.GetProduct(uid, role, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "product created but failed to retrieve"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":          createdProduct.ID,
			"name":        createdProduct.Name,
			"description": createdProduct.Description,
			"price":       createdProduct.Price,
			"stock":       createdProduct.Stock,
			"category_id": createdProduct.CategoryID,
			"image":       createdProduct.ImageURL,
			"created_at":  createdProduct.CreatedAt,
		})
	}
}

func makeListHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// user id from context (set by middleware)
		uid, ok := middleware.GinGetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		role, _ := middleware.GinGetRole(c)

		filters := map[string]string{}
		if v := c.Query("search"); v != "" {
			filters["search"] = v
		}
		if v := c.Query("category_id"); v != "" {
			filters["category_id"] = v
		}
		if v := c.Query("min_price"); v != "" {
			filters["min_price"] = v
		}
		if v := c.Query("max_price"); v != "" {
			filters["max_price"] = v
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

		data, total, err := uc.ListProducts(uid, role, filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resp := map[string]interface{}{
			"data":       data,
			"pagination": map[string]interface{}{"page": page, "limit": limit, "total": total},
		}
		c.JSON(http.StatusOK, resp)
	}
}

func makeGetHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// user id from context (set by middleware)
		uid, ok := middleware.GinGetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		role, _ := middleware.GinGetRole(c)

		idStr := c.Param("id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		p, err := uc.GetProduct(uid, role, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if p == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, p)
	}
}

func makeUpdateHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// user id from context (set by middleware)
		uid, ok := middleware.GinGetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		role, _ := middleware.GinGetRole(c)

		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		var req struct {
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Price       float64 `json:"price"`
			Stock       int     `json:"stock"`
			CategoryID  *int64  `json:"category_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
			return
		}

		err := uc.UpdateProduct(uid, role, id, req.Name, req.Description, req.Price, req.Stock, req.CategoryID)
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
		// user id from context (set by middleware)
		uid, ok := middleware.GinGetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		role, _ := middleware.GinGetRole(c)

		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		err := uc.DeleteProduct(uid, role, id)
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
