package cache

import (
	"fmt"
	"time"

	"github.com/example/ms-ecommerce/internal/pkg/db"
	"github.com/example/ms-ecommerce/internal/pkg/models"
)

// ProductCache provides caching for product operations
type ProductCache struct {
	cache *db.RedisCache
}

// NewProductCache creates a new product cache instance
func NewProductCache(cache *db.RedisCache) *ProductCache {
	return &ProductCache{cache: cache}
}

// GetProductsCacheKey generates cache key for product list
func (c *ProductCache) GetProductsCacheKey(filters map[string]string, page, limit int) string {
	key := fmt.Sprintf("products:list:%d:%d", page, limit)
	for k, v := range filters {
		if v != "" {
			key += fmt.Sprintf(":%s=%s", k, v)
		}
	}
	return key
}

// GetProductCacheKey generates cache key for single product
func (c *ProductCache) GetProductCacheKey(id int64) string {
	return fmt.Sprintf("products:id:%d", id)
}

// SetProducts caches product list with filters
func (c *ProductCache) SetProducts(key string, products []*models.Product, total int, expiration time.Duration) error {
	cacheData := struct {
		Products []*models.Product `json:"products"`
		Total    int               `json:"total"`
	}{
		Products: products,
		Total:    total,
	}
	return c.cache.SetJSON(key, cacheData, expiration)
}

// GetProducts retrieves cached product list
func (c *ProductCache) GetProducts(key string) ([]*models.Product, int, error) {
	var cacheData struct {
		Products []*models.Product `json:"products"`
		Total    int               `json:"total"`
	}
	err := c.cache.GetJSON(key, &cacheData)
	if err != nil {
		return nil, 0, err
	}
	return cacheData.Products, cacheData.Total, nil
}

// SetProduct caches single product
func (c *ProductCache) SetProduct(key string, product *models.Product, expiration time.Duration) error {
	return c.cache.SetJSON(key, product, expiration)
}

// GetProduct retrieves cached product
func (c *ProductCache) GetProduct(key string) (*models.Product, error) {
	var product models.Product
	err := c.cache.GetJSON(key, &product)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// DeleteProduct removes product from cache
func (c *ProductCache) DeleteProduct(key string) error {
	return c.cache.Delete(key)
}

// InvalidateProductsCache invalidates all product-related cache
func (c *ProductCache) InvalidateProductsCache() error {
	// In a production system, you might want to use Redis SCAN or KEYS
	// For simplicity, we'll flush all cache (not recommended for production)
	return c.cache.FlushAll()
}
