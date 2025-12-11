package auth

import (
	"database/sql"
	"time"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

type Repository interface {
	CreateUser(user *models.User) (int64, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdatePassword(userID int64, hashed string) error
	GetUserByID(id int64) (*models.User, error)
	// ListUsers returns a page of users and the total count. If search is non-empty,
	// it filters by name or email containing the search term.
	ListUsers(page, limit int, search string) ([]*models.User, int, error)

	// Refresh token operations
	CreateRefreshToken(userID int64, tokenHash string, expiresAt time.Time) error
	GetRefreshToken(tokenHash string) (userID int64, expiresAt time.Time, found bool, err error)
	DeleteRefreshToken(tokenHash string) error
	DeleteRefreshTokensByUser(userID int64) error
}

type mysqlRepo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repository {
	return &mysqlRepo{db: db}
}

func (r *mysqlRepo) CreateUser(user *models.User) (int64, error) {
	res, err := r.db.Exec("INSERT INTO users (name,email,phone,password,role) VALUES (?,?,?,?,?)", user.Name, user.Email, user.Phone, user.Password, user.Role)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	// create store automatically
	_, _ = r.db.Exec("INSERT INTO stores (user_id,name) VALUES (?,?)", id, user.Name+"'s store")
	return id, nil
}

func (r *mysqlRepo) GetUserByEmail(email string) (*models.User, error) {
	u := &models.User{}
	row := r.db.QueryRow("SELECT id,name,email,phone,password,role,created_at FROM users WHERE email = ?", email)
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.Password, &u.Role, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *mysqlRepo) UpdatePassword(userID int64, hashed string) error {
	_, err := r.db.Exec("UPDATE users SET password = ? WHERE id = ?", hashed, userID)
	return err
}

func (r *mysqlRepo) GetUserByID(id int64) (*models.User, error) {
	u := &models.User{}
	row := r.db.QueryRow("SELECT id,name,email,phone,role,created_at FROM users WHERE id = ?", id)
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.Role, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *mysqlRepo) ListUsers(page, limit int, search string) ([]*models.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	// total count
	var total int
	if search == "" {
		if err := r.db.QueryRow("SELECT COUNT(1) FROM users").Scan(&total); err != nil {
			return nil, 0, err
		}
	} else {
		like := "%" + search + "%"
		if err := r.db.QueryRow("SELECT COUNT(1) FROM users WHERE name LIKE ? OR email LIKE ?", like, like).Scan(&total); err != nil {
			return nil, 0, err
		}
	}

	var rows *sql.Rows
	var err error
	if search == "" {
		rows, err = r.db.Query("SELECT id,name,email,phone,role,created_at FROM users ORDER BY id ASC LIMIT ? OFFSET ?", limit, offset)
	} else {
		like := "%" + search + "%"
		rows, err = r.db.Query("SELECT id,name,email,phone,role,created_at FROM users WHERE name LIKE ? OR email LIKE ? ORDER BY id ASC LIMIT ? OFFSET ?", like, like, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []*models.User{}
	for rows.Next() {
		u := &models.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.Role, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, u)
	}
	return out, total, nil
}

func (r *mysqlRepo) CreateRefreshToken(userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec("INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES (?,?,?)", userID, tokenHash, expiresAt)
	return err
}

func (r *mysqlRepo) GetRefreshToken(tokenHash string) (int64, time.Time, bool, error) {
	var userID int64
	var expiresAt time.Time
	row := r.db.QueryRow("SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = ?", tokenHash)
	if err := row.Scan(&userID, &expiresAt); err != nil {
		if err == sql.ErrNoRows {
			return 0, time.Time{}, false, nil
		}
		return 0, time.Time{}, false, err
	}
	return userID, expiresAt, true, nil
}

func (r *mysqlRepo) DeleteRefreshToken(tokenHash string) error {
	_, err := r.db.Exec("DELETE FROM refresh_tokens WHERE token_hash = ?", tokenHash)
	return err
}

func (r *mysqlRepo) DeleteRefreshTokensByUser(userID int64) error {
	_, err := r.db.Exec("DELETE FROM refresh_tokens WHERE user_id = ?", userID)
	return err
}
