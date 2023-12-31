package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

const (
	ScopeAuthentication = "authentication"
)

// Token represents a token for a specific user authentication.
type Token struct {
	Plaintext string    `json:"token"`
	UserID    int64     `json:"-"`
	Hash      []byte    `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

// GenerateToken creates a new Token for a specific user ID and with a given
// scope and expiry time.
func GenerateToken(userID int, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: int64(userID),
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

// InsertToken inserts a new token into the database.
func (m *DBModel) InsertToken(t *Token, u User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// delete existing tokens for this user and scope.
	stmt := `
		delete from tokens
		where user_id = ?
	`
	_, err := m.DB.ExecContext(ctx, stmt, t.UserID)
	if err != nil {
		return err
	}

	stmt = `
		insert into tokens (user_id, name, email, token_hash, expiry, created_at, updated_at)
		values (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = m.DB.ExecContext(ctx, stmt,
		t.UserID,
		u.FirstName,
		u.Email,
		t.Hash,
		t.Expiry,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}

// GetUserByToken gets a user by a specific token.
func (m *DBModel) GetUserByToken(token string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(token))
	var user User

	query := `
		select 
			u.id, u.first_name, u.last_name, u.email
		from 
			tokens t
		left join 
			users u on (t.user_id = u.id)
		where 
			t.token_hash = ?
			and t.expiry > ?
	`

	row := m.DB.QueryRowContext(ctx, query, tokenHash[:], time.Now())
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
