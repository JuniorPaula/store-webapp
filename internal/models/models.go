package models

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
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
	IsRecurring    bool   `json:"is_recurring"`
	PlanID         string `json:"plan_id"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Order represents a customer order in the database.
type Order struct {
	ID            int         `json:"id"`
	WidgetID      int         `json:"widget_id"`
	TransactionID int         `json:"transaction_id"`
	CustomerID    int         `json:"customer_id"`
	StatusID      int         `json:"status_id"`
	Quantity      int         `json:"quantity"`
	Amount        int         `json:"amount"`
	Widget        Widget      `json:"widget"`
	Transaction   Transaction `json:"transaction"`
	Customer      Customer    `json:"customer"`

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
	Lastfour            string `json:"last_four"`
	ExpiryMonth         int    `json:"expiry_month"`
	ExpiryYear          int    `json:"expiry_year"`
	PaymentIntent       string `json:"payment_intent"`
	PaymentMethod       string `json:"payment_method"`
	BankReturnCode      string `json:"bank_return_code"`
	TransactionStatusID int    `json:"transaction_status_id"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// User is a type to represent a user in the database.
type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Customer is a type to represent a customer in the database.
type Customer struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`

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
			id, name, description, inventory_level, price, coalesce(image, ''), is_recurring, plan_id, created_at, updated_at 
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
		&widget.IsRecurring,
		&widget.PlanID,
		&widget.CreatedAt,
		&widget.UpdatedAt,
	)
	if err != nil {
		log.Println(err)
		return widget, err
	}

	return widget, nil
}

// InsertTransaction inserts an net txn into the database and return its id.
func (m *DBModel) InsertTransaction(txn Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		insert into transactions
			(amount, currency, last_four, bank_return_code, expiry_month, expiry_year, payment_intent, payment_method, transaction_status_id, created_at, updated_at)
		values	
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := m.DB.ExecContext(ctx, stmt,
		txn.Amount,
		txn.Currency,
		txn.Lastfour,
		txn.BankReturnCode,
		txn.ExpiryMonth,
		txn.ExpiryYear,
		txn.PaymentIntent,
		txn.PaymentMethod,
		txn.TransactionStatusID,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return int(id), nil
}

// InsertOrder inserts an order into the database and return its id.
func (m *DBModel) InsertOrder(order Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		insert into orders
			(widget_id, transaction_id, status_id, quantity, customer_id, amount, created_at, updated_at)
		values	
			(?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := m.DB.ExecContext(ctx, stmt,
		order.WidgetID,
		order.TransactionID,
		order.StatusID,
		order.Quantity,
		order.CustomerID,
		order.Amount,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return int(id), nil
}

// InsertCustomer inserts a customer into the database and return its id.
func (m *DBModel) InsertCustomer(c Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		insert into customers
			(first_name, last_name, email, created_at, updated_at)
		values
			(?, ?, ?, ?, ?)
	`

	result, err := m.DB.ExecContext(ctx, stmt,
		c.FirstName,
		c.LastName,
		c.Email,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		log.Println(err)
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return int(id), nil
}

// GetUserByEmail returns a user by email address.
func (m *DBModel) GetUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	email = strings.ToLower(email)
	var u User

	row := m.DB.QueryRowContext(ctx, `
		select
			id, first_name, last_name, email, password, created_at, updated_at
		from
			users
		where
			email = ?`, email,
	)

	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return u, err
	}

	return u, nil
}

// Authenticate authenticates a user.
func (m *DBModel) Authenticate(email, password string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, `select id, password from users where email = ?`, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, errors.New("incorrect password")
	} else if err != nil {
		return 0, err
	}

	return id, nil
}

// UpdatePassword updates a user's password.
func (m *DBModel) UpdatePassword(hash string, user User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		update users
		set password = ?
		where id = ?
	`
	_, err := m.DB.ExecContext(ctx, stmt, hash, user.ID)
	if err != nil {
		return err
	}

	return nil
}

// GetAllOrders returns all orders.
func (m *DBModel) GetAllOrders() ([]*Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		select
			o.id, o.widget_id, o.transaction_id, o.customer_id, o.status_id, o.quantity, o.amount,
			w.id, w.name, t.id, t.amount, t.currency, t.last_four, t.expiry_month, t.expiry_year, t.payment_intent, t.payment_method, t.bank_return_code, c.id, c.first_name, c.last_name, c.email,
			o.created_at, o.updated_at
		from
			orders o
		left join
			widgets w on (o.widget_id = w.id)
		left join
			transactions t on (o.transaction_id = t.id)
		left join
			customers c on (o.customer_id = c.id)
		where
			w.is_recurring = 0
		order by
			o.created_at desc
	`

	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order

	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.ID,
			&o.WidgetID,
			&o.TransactionID,
			&o.CustomerID,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.Widget.ID,
			&o.Widget.Name,
			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.Lastfour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.PaymentMethod,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
			&o.CreatedAt,
			&o.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, &o)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// GetAllOrdersPaginated returns a slice of a subset of orders
func (m *DBModel) GetAllOrdersPaginated(pageSize, page int) ([]*Order, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	offset := (page - 1) * pageSize

	stmt := `
		select
			o.id, o.widget_id, o.transaction_id, o.customer_id, o.status_id, o.quantity, o.amount,
			w.id, w.name, t.id, t.amount, t.currency, t.last_four, t.expiry_month, t.expiry_year, t.payment_intent, t.payment_method, t.bank_return_code, c.id, c.first_name, c.last_name, c.email,
			o.created_at, o.updated_at
		from
			orders o
		left join
			widgets w on (o.widget_id = w.id)
		left join
			transactions t on (o.transaction_id = t.id)
		left join
			customers c on (o.customer_id = c.id)
		where
			w.is_recurring = 0
		order by
			o.created_at desc
		limit ? offset ?
	`

	rows, err := m.DB.QueryContext(ctx, stmt, pageSize, offset)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	var orders []*Order

	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.ID,
			&o.WidgetID,
			&o.TransactionID,
			&o.CustomerID,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.Widget.ID,
			&o.Widget.Name,
			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.Lastfour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.PaymentMethod,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
			&o.CreatedAt,
			&o.UpdatedAt,
		)
		if err != nil {
			return nil, 0, 0, err
		}

		orders = append(orders, &o)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, 0, err
	}

	stmt = `
		select count(o.id)
		from orders o
		left join widgets w on (o.widget_id = w.id)
		where 
		w.is_recurring = 0 
	`
	var totalRecords int
	countRow := m.DB.QueryRowContext(ctx, stmt)
	err = countRow.Scan(&totalRecords)
	if err != nil {
		return nil, 0, 0, err
	}

	lastPage := totalRecords / pageSize

	return orders, lastPage, totalRecords, nil
}

// GetOrderByID returns an order by id.
func (m *DBModel) GetOrderByID(id int) (Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		select
			o.id, o.widget_id, o.transaction_id, o.customer_id, o.status_id, o.quantity, o.amount,
			w.id, w.name, t.id, t.amount, t.currency, t.last_four, t.expiry_month, t.expiry_year, t.payment_intent, t.payment_method, t.bank_return_code, c.id, c.first_name, c.last_name, c.email,
			o.created_at, o.updated_at
		from
			orders o
		left join
			widgets w on (o.widget_id = w.id)
		left join
			transactions t on (o.transaction_id = t.id)
		left join
			customers c on (o.customer_id = c.id)
		where
			o.id = ?
	`

	var o Order

	row := m.DB.QueryRowContext(ctx, stmt, id)
	err := row.Scan(
		&o.ID,
		&o.WidgetID,
		&o.TransactionID,
		&o.CustomerID,
		&o.StatusID,
		&o.Quantity,
		&o.Amount,
		&o.Widget.ID,
		&o.Widget.Name,
		&o.Transaction.ID,
		&o.Transaction.Amount,
		&o.Transaction.Currency,
		&o.Transaction.Lastfour,
		&o.Transaction.ExpiryMonth,
		&o.Transaction.ExpiryYear,
		&o.Transaction.PaymentIntent,
		&o.Transaction.PaymentMethod,
		&o.Transaction.BankReturnCode,
		&o.Customer.ID,
		&o.Customer.FirstName,
		&o.Customer.LastName,
		&o.Customer.Email,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		return o, err
	}

	return o, nil
}

// GetAllSubscriptions returns all subscriptions.
func (m *DBModel) GetAllSubscriptions() ([]*Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		select
			o.id, o.widget_id, o.transaction_id, o.customer_id, o.status_id, o.quantity, o.amount,
			w.id, w.name, t.id, t.amount, t.currency, t.last_four, t.expiry_month, t.expiry_year, t.payment_intent, t.payment_method, t.bank_return_code, c.id, c.first_name, c.last_name, c.email,
			o.created_at, o.updated_at
		from
			orders o
		left join
			widgets w on (o.widget_id = w.id)
		left join
			transactions t on (o.transaction_id = t.id)
		left join
			customers c on (o.customer_id = c.id)
		where
			w.is_recurring = 1
		order by
			o.created_at desc
	`

	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order

	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.ID,
			&o.WidgetID,
			&o.TransactionID,
			&o.CustomerID,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.Widget.ID,
			&o.Widget.Name,
			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.Lastfour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.PaymentMethod,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
			&o.CreatedAt,
			&o.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, &o)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// GetSubscriptionByID returns a subscription by id.
func (m *DBModel) GetSubscriptionByID(id int) (Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		select
			o.id, o.widget_id, o.transaction_id, o.customer_id, o.status_id, o.quantity, o.amount,
			w.id, w.name, t.id, t.amount, t.currency, t.last_four, t.expiry_month, t.expiry_year, t.payment_intent, t.payment_method, t.bank_return_code, c.id, c.first_name, c.last_name, c.email,
			o.created_at, o.updated_at
		from
			orders o
		left join
			widgets w on (o.widget_id = w.id)
		left join
			transactions t on (o.transaction_id = t.id)
		left join
			customers c on (o.customer_id = c.id)
		where
			o.id = ?
	`

	var o Order

	row := m.DB.QueryRowContext(ctx, stmt, id)
	err := row.Scan(
		&o.ID,
		&o.WidgetID,
		&o.TransactionID,
		&o.CustomerID,
		&o.StatusID,
		&o.Quantity,
		&o.Amount,
		&o.Widget.ID,
		&o.Widget.Name,
		&o.Transaction.ID,
		&o.Transaction.Amount,
		&o.Transaction.Currency,
		&o.Transaction.Lastfour,
		&o.Transaction.ExpiryMonth,
		&o.Transaction.ExpiryYear,
		&o.Transaction.PaymentIntent,
		&o.Transaction.PaymentMethod,
		&o.Transaction.BankReturnCode,
		&o.Customer.ID,
		&o.Customer.FirstName,
		&o.Customer.LastName,
		&o.Customer.Email,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		return o, err
	}

	return o, nil
}

// UpdateOrderStatus updates an order status.
func (m *DBModel) UpdateOrderStatus(id, statusID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update orders set status_id = ? where id = ?`

	_, err := m.DB.ExecContext(ctx, stmt, statusID, id)
	if err != nil {
		return err
	}

	return nil
}
