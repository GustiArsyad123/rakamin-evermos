package models

import "time"

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Product struct {
	ID          int64     `json:"id"`
	StoreID     int64     `json:"store_id"`
	CategoryID  *int64    `json:"category_id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
}

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Transaction struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	StoreID   int64     `json:"store_id"`
	Total     float64   `json:"total"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type ProductLog struct {
	ID            int64     `json:"id"`
	TransactionID int64     `json:"transaction_id"`
	ProductID     int64     `json:"product_id"`
	ProductName   string    `json:"product_name"`
	ProductPrice  float64   `json:"product_price"`
	Quantity      int       `json:"quantity"`
	CreatedAt     time.Time `json:"created_at"`
}
