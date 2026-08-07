package repository

import (
	"database/sql"
	"errors"
	"strings"

	"minimal-service/internal/model"
)

// ErrNotFound is returned when a user does not exist.
var ErrNotFound = errors.New("user not found")

// ErrAlreadyExists is returned when a unique constraint is violated (e.g. username taken).
var ErrAlreadyExists = errors.New("user already exists")

// UserRepository abstracts database operations on users.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new repository backed by the given connection.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user and writes the assigned id back into u.
func (r *UserRepository) Create(u *model.User) error {
	err := r.db.QueryRow(
		`INSERT INTO users (username, password_hash, email, phone, status, confirmation_token)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		u.Username, u.PasswordHash, u.Email, u.Phone, u.Status, u.ConfirmationToken,
	).Scan(&u.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

// GetByID returns a user by ID.
func (r *UserRepository) GetByID(id int) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(
		`SELECT id, username, password_hash, email, phone, status, confirmation_token
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email, &u.Phone, &u.Status, &u.ConfirmationToken)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}

// GetByUsername returns a user by username (used during login).
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(
		`SELECT id, username, password_hash, email, phone, status, confirmation_token
		 FROM users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email, &u.Phone, &u.Status, &u.ConfirmationToken)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}

// GetByConfirmationToken returns a user by confirmation token.
func (r *UserRepository) GetByConfirmationToken(token string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(
		`SELECT id, username, password_hash, email, phone, status, confirmation_token
		 FROM users WHERE confirmation_token = $1`, token,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email, &u.Phone, &u.Status, &u.ConfirmationToken)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}

// ConfirmUser activates the account and clears the confirmation token.
func (r *UserRepository) ConfirmUser(id int) error {
	res, err := r.db.Exec(
		`UPDATE users
		 SET status=$1, confirmation_token=NULL, updated_at=CURRENT_TIMESTAMP
		 WHERE id=$2 AND status=$3`,
		model.UserStatusActive, id, model.UserStatusConfirmPending,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateProfile modifies the editable profile fields of an existing user.
// Returns ErrNotFound if no rows were affected.
func (r *UserRepository) UpdateProfile(id int, p *model.UpdateProfileDTO) error {
	res, err := r.db.Exec(
		`UPDATE users
		 SET email=$1, phone=$2
		 WHERE id=$3`,
		p.Email, p.Phone, id,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// isUniqueViolation tries to detect a Postgres unique_violation (SQLSTATE 23505)
// without depending on a specific driver error type.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "23505")
}
