package models

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// DBModel wraps the connection values.
type DBModel struct {
	DB *sql.DB
}

// Models wraps the DBModel.
type Models struct {
	DB DBModel
}

// NewModels creates a new Models type.
func NewModels(db *sql.DB) Models {
	return Models{
		DB: DBModel{DB: db},
	}
}

// Widget represents a widget for sale in the database.
type Widget struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	InventoryLevel int    `json:"inventory_level"`
	Price          int    `json:"price"`
	Image          string `json:"image"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Order represents a customer order in the database.
type Order struct {
	ID            int `json:"id"`
	WidgetID      int `json:"widget_id"`
	TransactionID int `json:"transaction_id"`
	StatusID      int `json:"status_id"`
	Quantity      int `json:"quantity"`
	Amount        int `json:"amount"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Statuses represents the status of an order in the database.
type Statuses struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Transaction represents a transaction in the database.
type TransactionStatuses struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Transaction is a type to represent a transaction in the database.
type Transaction struct {
	ID                  int    `json:"id"`
	Amount              int    `json:"amount"`
	Currency            string `json:"currency"`
	Lastfor             string `json:"last_for"`
	BankReturnCode      string `json:"bank_return_code"`
	TransactionStatusID int    `json:"transaction_status_id"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// User is a type to represent a user in the database.
type User struct {
	ID       int    `json:"id"`
	FirsName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email    string `json:"email"`
	Password string `json:"password"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// OrderStatus represents the status of an order in the database.
func (m *DBModel) GetWidget(ID int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var widget Widget

	row := m.DB.QueryRowContext(ctx, `
		select 
			id, name, description, inventory_level, price, coalesce(image, ''), created_at, updated_at 
		from 
			widgets 
		where id = ?`, ID,
	)

	err := row.Scan(
		&widget.ID,
		&widget.Name,
		&widget.Description,
		&widget.InventoryLevel,
		&widget.Price,
		&widget.Image,
		&widget.CreatedAt,
		&widget.UpdatedAt,
	)
	if err != nil {
		log.Println(err)
		return widget, err
	}

	return widget, nil
}
