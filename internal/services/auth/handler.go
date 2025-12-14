package auth

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"time"

	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/example/ms-ecommerce/internal/pkg/models"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, dbConn *sql.DB) {
	repo := NewRepo(dbConn)
	uc := NewUsecase(repo)
	r.POST("/api/v1/auth/register", makeRegisterHandler(uc))
	log.Printf("registered POST /api/v1/auth/register")
	r.POST("/api/v1/auth/login", makeLoginHandler(uc))
	log.Printf("registered POST /api/v1/auth/login")
	r.POST("/api/v1/auth/forgot-password", makeForgotHandler(uc))
	log.Printf("registered POST /api/v1/auth/forgot-password")
	r.POST("/api/v1/auth/reset-password", makeResetHandler(uc))
	log.Printf("registered POST /api/v1/auth/reset-password")
	r.POST("/api/v1/auth/sso/google", makeSSOHandler(uc))
	log.Printf("registered POST /api/v1/auth/sso/google")
	r.POST("/api/v1/auth/refresh", makeRefreshHandler(uc))
	log.Printf("registered POST /api/v1/auth/refresh")
	r.POST("/api/v1/auth/logout", makeRevokeHandler(uc))
	log.Printf("registered POST /api/v1/auth/logout")
	// revoke all tokens (self) or admin revoke for a user
	r.POST("/api/v1/auth/logout/all", middleware.GinJWTAuth(), makeRevokeAllHandler(uc))
	log.Printf("registered POST /api/v1/auth/logout/all")
	// Admin-only list users (register both trailing and non-trailing variants)
	r.GET("/api/v1/auth/users", middleware.GinJWTAuth(), middleware.GinRequireRole("admin"), makeListUsersHandler(uc))
	r.GET("/api/v1/auth/users/", middleware.GinJWTAuth(), middleware.GinRequireRole("admin"), makeListUsersHandler(uc))
	log.Printf("registered GET /api/v1/auth/users and /api/v1/auth/users/")
	// Get user by id: requires JWT; allowed if admin or owner
	r.GET("/api/v1/auth/users/:id", middleware.GinJWTAuth(), makeGetUserHandler(uc))
	log.Printf("registered GET /api/v1/auth/users/:id")
	// Update user (owner or admin). Admin may change role.
	r.PUT("/api/v1/auth/users/:id", middleware.GinJWTAuth(), makeUpdateUserHandler(uc))
	log.Printf("registered PUT /api/v1/auth/users/:id")
}

func makeRegisterHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Phone    string `json:"phone"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if req.Name == "" || req.Email == "" || req.Phone == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name, email, phone, and password are required"})
			return
		}
		u := &models.User{Name: req.Name, Email: req.Email, Phone: req.Phone, Password: req.Password}
		token, userID, err := uc.Register(u)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		refresh, refreshExp, err := uc.IssueRefreshToken(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		expiresAt := time.Now().UTC().Add(1 * time.Hour)
		c.JSON(http.StatusOK, map[string]interface{}{
			"token":              token,
			"expires_in":         3600,
			"expires_at":         expiresAt.Format(time.RFC3339),
			"refresh_token":      refresh,
			"refresh_expires_at": refreshExp.Format(time.RFC3339),
		})
	}
}

func makeForgotHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		token, err := uc.ForgotPassword(req.Email)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// In real app send email; for now return token so it can be used in tests/Postman
		c.JSON(http.StatusOK, map[string]interface{}{"reset_token": token, "expires_in": 3600})
	}
}

func makeResetHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Token    string `json:"token"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Token == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if err := uc.ResetPassword(req.Token, req.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func makeSSOHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			IDToken string `json:"id_token"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.IDToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		token, userID, err := uc.SSOLogin(req.IDToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		refresh, refreshExp, err := uc.IssueRefreshToken(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		expiresAt := time.Now().UTC().Add(1 * time.Hour)
		c.JSON(http.StatusOK, map[string]interface{}{
			"token":              token,
			"expires_in":         3600,
			"expires_at":         expiresAt.Format(time.RFC3339),
			"refresh_token":      refresh,
			"refresh_expires_at": refreshExp.Format(time.RFC3339),
		})
	}
}

func makeLoginHandler(uc Usecase) gin.HandlerFunc {
	type loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type loginResponse struct {
		Token            string `json:"token"`
		ExpiresIn        int    `json:"expires_in"`
		ExpiresAt        string `json:"expires_at"`
		RefreshToken     string `json:"refresh_token"`
		RefreshExpiresAt string `json:"refresh_expires_at"`
	}

	return func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		token, userID, err := uc.Login(req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		refreshToken, refreshExp, err := uc.IssueRefreshToken(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue refresh token"})
			return
		}

		expiresIn := 3600
		expiresAt := time.Now().UTC().Add(time.Duration(expiresIn) * time.Second)

		resp := loginResponse{
			Token:            token,
			ExpiresIn:        expiresIn,
			ExpiresAt:        expiresAt.Format(time.RFC3339),
			RefreshToken:     refreshToken,
			RefreshExpiresAt: refreshExp.UTC().Format(time.RFC3339),
		}

		c.JSON(http.StatusOK, resp)
	}
}

func makeListUsersHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// parse pagination
		page := 1
		limit := 10
		// enforce a reasonable max limit
		const maxLimit = 100
		if p := c.Query("page"); p != "" {
			if v, err := strconv.Atoi(p); err == nil && v > 0 {
				page = v
			}
		}
		if l := c.Query("limit"); l != "" {
			if v, err := strconv.Atoi(l); err == nil && v > 0 {
				limit = v
			}
		}
		if limit > maxLimit {
			limit = maxLimit
		}
		search := c.Query("search")
		users, total, err := uc.ListUsers(page, limit, search)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resp := map[string]interface{}{
			"data": users,
			"pagination": map[string]interface{}{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		}
		c.JSON(http.StatusOK, resp)
	}
}

func makeGetUserHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		// permission: admin or owner
		reqUID, ok := middleware.GinGetUserID(c)
		role, _ := middleware.GinGetRole(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if role != "admin" && reqUID != id {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		u, err := uc.GetUserByID(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if u == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, u)
	}
}

func makeUpdateUserHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		reqUID, ok := middleware.GinGetUserID(c)
		role, _ := middleware.GinGetRole(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		var body struct {
			Name  *string `json:"name"`
			Phone *string `json:"phone"`
			Role  *string `json:"role"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		// enforce authorization in usecase
		if err := uc.UpdateUser(reqUID, id, role, body.Name, body.Phone, body.Role); err != nil {
			if err.Error() == "forbidden" {
				c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func makeRefreshHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		token, refresh, refreshExp, err := uc.Refresh(req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		expiresAt := time.Now().UTC().Add(1 * time.Hour)
		c.JSON(http.StatusOK, map[string]interface{}{
			"token":              token,
			"expires_in":         3600,
			"expires_at":         expiresAt.Format(time.RFC3339),
			"refresh_token":      refresh,
			"refresh_expires_at": refreshExp.Format(time.RFC3339),
		})
	}
}

func makeRevokeHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if err := uc.RevokeRefreshToken(req.RefreshToken); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func makeRevokeAllHandler(uc Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID *int64 `json:"user_id,omitempty"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		// must be authenticated
		requesterID, ok := middleware.GinGetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		role, _ := middleware.GinGetRole(c)

		var target int64
		if req.UserID != nil {
			// admin may revoke for any user
			if role != "admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
				return
			}
			target = *req.UserID
		} else {
			target = requesterID
		}

		if err := uc.RevokeAllRefreshTokens(target); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
