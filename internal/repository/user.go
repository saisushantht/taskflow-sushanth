package repository

import (
	"database/sql"
	"taskflow/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(name, email, hashedPassword string) (*model.User, error) {
	user := &model.User{}
	query := `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at`

	err := r.db.QueryRow(query, name, email, hashedPassword).
		Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, name, email, password, created_at FROM users WHERE email = $1`
	err := r.db.QueryRow(query, email).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByID(id string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, name, email, created_at FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).
		Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}
