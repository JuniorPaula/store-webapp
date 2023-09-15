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

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (m *DBModel) GetWidget(ID int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var widget Widget

	row := m.DB.QueryRowContext(ctx, `select id, name from widgets where id = ?`, ID)
	err := row.Scan(&widget.ID, &widget.Name)
	if err != nil {
		log.Println(err)
		return widget, err
	}

	return widget, nil
}
