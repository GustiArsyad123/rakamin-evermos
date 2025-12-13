package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	jwtpkg "github.com/example/ms-ecommerce/internal/pkg/jwt"
	"github.com/example/ms-ecommerce/internal/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type Usecase interface {
	Register(user *models.User) (string, int64, error)
	Login(email, password string) (string, int64, error)
	ForgotPassword(email string) (string, error)
	ResetPassword(token, newPassword string) error
	SSOLogin(idToken string) (string, int64, error)
	GetUserByID(id int64) (*models.User, error)
	UpdateUser(requesterID, id int64, requesterRole string, name, phone, role *string) error
	ListUsers(page, limit int, search string) ([]*models.User, int, error)
	IssueRefreshToken(userID int64) (string, time.Time, error)
	Refresh(refreshToken string) (string, string, time.Time, error)
	RevokeRefreshToken(refreshToken string) error
}

type authUsecase struct {
	repo Repository
}

func NewUsecase(r Repository) Usecase {
	return &authUsecase{repo: r}
}

func (u *authUsecase) Register(user *models.User) (string, int64, error) {
	// uniqueness checks
	existing, _ := u.repo.GetUserByEmail(user.Email)
	if existing != nil {
		return "", 0, errors.New("email already registered")
	}
	existing, _ = u.repo.GetUserByPhone(user.Phone)
	if existing != nil {
		return "", 0, errors.New("phone number already registered")
	}
	// hash password
	h, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", 0, err
	}
	user.Password = string(h)
	id, err := u.repo.CreateUser(user)
	if err != nil {
		return "", 0, err
	}
	// create store automatically
	storeName := user.Name + "'s Store"
	err = u.repo.CreateStore(id, storeName)
	if err != nil {
		// If store creation fails, we should ideally delete the user, but for simplicity, log error
		// In production, use transactions
		return "", 0, errors.New("user created but failed to create store: " + err.Error())
	}
	token, err := jwtpkg.GenerateToken(id, user.Role)
	if err != nil {
		return "", 0, err
	}
	return token, id, nil
}

func (u *authUsecase) Login(email, password string) (string, int64, error) {
	user, err := u.repo.GetUserByEmail(email)
	if err != nil || user == nil {
		return "", 0, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", 0, errors.New("invalid credentials")
	}
	token, err := jwtpkg.GenerateToken(user.ID, user.Role)
	if err != nil {
		return "", 0, err
	}
	return token, user.ID, nil
}

func (u *authUsecase) ForgotPassword(email string) (string, error) {
	user, err := u.repo.GetUserByEmail(email)
	if err != nil || user == nil {
		return "", errors.New("user not found")
	}
	// generate short-lived reset token
	token, err := jwtpkg.GenerateResetToken(user.ID)
	if err != nil {
		return "", err
	}
	// In production you'd email this token; for now return it in response
	return token, nil
}

func (u *authUsecase) ResetPassword(token, newPassword string) error {
	uid, err := jwtpkg.ParseResetToken(token)
	if err != nil {
		return err
	}
	// hash password
	h, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return u.repo.UpdatePassword(uid, string(h))
}

// SSOLogin accepts a Google ID token, validates it via Google's tokeninfo endpoint,
// and logs the user in (creating an account automatically if needed).
func (u *authUsecase) SSOLogin(idToken string) (string, int64, error) {
	if idToken == "" {
		return "", 0, errors.New("missing id token")
	}
	// Verify token with Google tokeninfo endpoint
	endpoint := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken)
	resp, err := http.Get(endpoint)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", 0, errors.New("invalid id token")
	}
	var info struct {
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Aud           string `json:"aud"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", 0, err
	}
	if info.Email == "" {
		return "", 0, errors.New("invalid token payload")
	}

	// Try to find existing user
	user, err := u.repo.GetUserByEmail(info.Email)
	if err != nil {
		return "", 0, err
	}
	if user != nil {
		// return token for existing user
		tk, err := jwtpkg.GenerateToken(user.ID, user.Role)
		if err != nil {
			return "", 0, err
		}
		return tk, user.ID, nil
	}

	// Create a new user. Generate a random password (hashed) since login uses JWT.
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", 0, err
	}
	rawPwd := hex.EncodeToString(b)
	h, err := bcrypt.GenerateFromPassword([]byte(rawPwd), bcrypt.DefaultCost)
	if err != nil {
		return "", 0, err
	}
	newUser := &models.User{
		Name:     info.Name,
		Email:    info.Email,
		Password: string(h),
		Role:     "user",
	}
	id, err := u.repo.CreateUser(newUser)
	if err != nil {
		return "", 0, err
	}
	token, err := jwtpkg.GenerateToken(id, newUser.Role)
	if err != nil {
		return "", 0, err
	}
	return token, id, nil
}

func (u *authUsecase) GetUserByID(id int64) (*models.User, error) {
	return u.repo.GetUserByID(id)
}

func (u *authUsecase) UpdateUser(requesterID, id int64, requesterRole string, name, phone, role *string) error {
	// Only owner or admin can update. Only admin can change role.
	if requesterRole != "admin" && requesterID != id {
		return errors.New("forbidden")
	}
	if role != nil && requesterRole != "admin" {
		return errors.New("forbidden")
	}
	return u.repo.UpdateUser(id, name, phone, role)
}

func (u *authUsecase) ListUsers(page, limit int, search string) ([]*models.User, int, error) {
	return u.repo.ListUsers(page, limit, search)
}

// helper to get refresh token expiry from env, default 30 days
func getRefreshExpirySeconds() int64 {
	s := os.Getenv("REFRESH_TOKEN_EXP_SECONDS")
	if s == "" {
		return 60 * 60 * 24 * 30 // 30 days
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil || v <= 0 {
		return 60 * 60 * 24 * 30
	}
	return v
}

// IssueRefreshToken creates and stores a new refresh token for a user and returns the plaintext token and expiry
func (u *authUsecase) IssueRefreshToken(userID int64) (string, time.Time, error) {
	// generate random 32-byte token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}
	token := hex.EncodeToString(b)
	// hash for storage
	h := sha256.Sum256([]byte(token))
	hashHex := hex.EncodeToString(h[:])
	expiresAt := time.Now().UTC().Add(time.Duration(getRefreshExpirySeconds()) * time.Second)
	if err := u.repo.CreateRefreshToken(userID, hashHex, expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

// Refresh rotates a refresh token: validates provided token, issues new access and refresh tokens
func (u *authUsecase) Refresh(refreshToken string) (string, string, time.Time, error) {
	if refreshToken == "" {
		return "", "", time.Time{}, errors.New("missing refresh token")
	}
	h := sha256.Sum256([]byte(refreshToken))
	hashHex := hex.EncodeToString(h[:])
	userID, expiresAt, found, err := u.repo.GetRefreshToken(hashHex)
	if err != nil {
		return "", "", time.Time{}, err
	}
	if !found {
		return "", "", time.Time{}, errors.New("invalid refresh token")
	}
	if time.Now().UTC().After(expiresAt) {
		// delete expired token
		_ = u.repo.DeleteRefreshToken(hashHex)
		return "", "", time.Time{}, errors.New("refresh token expired")
	}
	// issue new access token
	accessToken, err := jwtpkg.GenerateToken(userID, "user")
	if err != nil {
		return "", "", time.Time{}, err
	}
	// rotate refresh token: delete old and create new
	if err := u.repo.DeleteRefreshToken(hashHex); err != nil {
		return "", "", time.Time{}, err
	}
	newRefresh, newExpiresAt, err := u.IssueRefreshToken(userID)
	if err != nil {
		return "", "", time.Time{}, err
	}
	return accessToken, newRefresh, newExpiresAt, nil
}

// RevokeRefreshToken deletes a refresh token (logout)
func (u *authUsecase) RevokeRefreshToken(refreshToken string) error {
	if refreshToken == "" {
		return errors.New("missing refresh token")
	}
	h := sha256.Sum256([]byte(refreshToken))
	hashHex := hex.EncodeToString(h[:])
	return u.repo.DeleteRefreshToken(hashHex)
}
