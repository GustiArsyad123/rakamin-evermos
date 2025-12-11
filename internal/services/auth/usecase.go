package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	jwtpkg "github.com/example/ms-ecommerce/internal/pkg/jwt"
	"github.com/example/ms-ecommerce/internal/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type Usecase interface {
	Register(user *models.User) (string, error)
	Login(email, password string) (string, error)
	ForgotPassword(email string) (string, error)
	ResetPassword(token, newPassword string) error
	SSOLogin(idToken string) (string, error)
}

type authUsecase struct {
	repo Repository
}

func NewUsecase(r Repository) Usecase {
	return &authUsecase{repo: r}
}

func (u *authUsecase) Register(user *models.User) (string, error) {
	// uniqueness checks
	existing, _ := u.repo.GetUserByEmail(user.Email)
	if existing != nil {
		return "", errors.New("email already registered")
	}
	// hash password
	h, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	user.Password = string(h)
	id, err := u.repo.CreateUser(user)
	if err != nil {
		return "", err
	}
	token, err := jwtpkg.GenerateToken(id, user.Role)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (u *authUsecase) Login(email, password string) (string, error) {
	user, err := u.repo.GetUserByEmail(email)
	if err != nil || user == nil {
		return "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	token, err := jwtpkg.GenerateToken(user.ID, user.Role)
	if err != nil {
		return "", err
	}
	return token, nil
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
func (u *authUsecase) SSOLogin(idToken string) (string, error) {
	if idToken == "" {
		return "", errors.New("missing id token")
	}
	// Verify token with Google tokeninfo endpoint
	endpoint := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken)
	resp, err := http.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New("invalid id token")
	}
	var info struct {
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Aud           string `json:"aud"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	if info.Email == "" {
		return "", errors.New("invalid token payload")
	}

	// Try to find existing user
	user, err := u.repo.GetUserByEmail(info.Email)
	if err != nil {
		return "", err
	}
	if user != nil {
		// return token for existing user
		return jwtpkg.GenerateToken(user.ID, user.Role)
	}

	// Create a new user. Generate a random password (hashed) since login uses JWT.
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	rawPwd := hex.EncodeToString(b)
	h, err := bcrypt.GenerateFromPassword([]byte(rawPwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	newUser := &models.User{
		Name:     info.Name,
		Email:    info.Email,
		Password: string(h),
		Role:     "user",
	}
	id, err := u.repo.CreateUser(newUser)
	if err != nil {
		return "", err
	}
	token, err := jwtpkg.GenerateToken(id, newUser.Role)
	if err != nil {
		return "", err
	}
	return token, nil
}
