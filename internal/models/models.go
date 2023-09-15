package models

import (
	"database/sql"
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

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
