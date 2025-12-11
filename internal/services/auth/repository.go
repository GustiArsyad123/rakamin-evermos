package auth

import (
	"database/sql"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

type Repository interface {
	CreateUser(user *models.User) (int64, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdatePassword(userID int64, hashed string) error
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
