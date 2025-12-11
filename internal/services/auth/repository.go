package auth

import (
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetUserByID(id int64) (*User, error) {
	var u User
	row := r.db.QueryRow("SELECT id, name, email, phone, role, created_at FROM users WHERE id = ?", id)
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.Role, &u.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetAllUsers() ([]*User, error) {
	rows, err := r.db.Query("SELECT id, name, email, phone, role, created_at FROM users ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.Role, &u.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// GetUserByEmail is needed for login/register, so we'll add it here.
func (r *Repository) GetUserByEmail(email string) (*User, string, error) {
	var u User
	var passwordHash string
	row := r.db.QueryRow("SELECT id, name, email, phone, role, password, created_at FROM users WHERE email = ?", email)
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.Role, &passwordHash, &u.CreatedAt)
	if err != nil {
		return nil, "", err
	}
	return &u, passwordHash, nil
}

// CreateUser is needed for register, so we'll add it here.
func (r *Repository) CreateUser(name, email, phone, passwordHash string) (int64, error) {
	res, err := r.db.Exec("INSERT INTO users (name, email, phone, password) VALUES (?, ?, ?, ?)", name, email, phone, passwordHash)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdatePassword updates a user's password hash.
func (r *Repository) UpdatePassword(userID int64, passwordHash string) error {
	_, err := r.db.Exec("UPDATE users SET password = ? WHERE id = ?", passwordHash, userID)
	if err != nil {
		return fmt.Errorf("could not update password for user %d: %w", userID, err)
	}
	return nil
}
