package file

import (
	"database/sql"
)

type Repository interface {
	// For future use - could track uploaded files in DB
}

type mysqlRepo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repository {
	return &mysqlRepo{db: db}
}
