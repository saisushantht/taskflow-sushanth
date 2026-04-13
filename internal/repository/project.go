package repository

import (
	"database/sql"
	"taskflow/internal/model"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(name, description, ownerID string) (*model.Project, error) {
	p := &model.Project{}
	query := `
		INSERT INTO projects (name, description, owner_id)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, owner_id, created_at`
	err := r.db.QueryRow(query, name, description, ownerID).
		Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProjectRepository) ListByUser(userID string) ([]model.Project, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.owner_id = $1 OR t.assignee_id = $1
		ORDER BY p.created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *ProjectRepository) FindByID(id string) (*model.Project, error) {
	p := &model.Project{}
	query := `SELECT id, name, description, owner_id, created_at FROM projects WHERE id = $1`
	err := r.db.QueryRow(query, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProjectRepository) Update(id, name, description string) (*model.Project, error) {
	p := &model.Project{}
	query := `
		UPDATE projects SET name = $1, description = $2
		WHERE id = $3
		RETURNING id, name, description, owner_id, created_at`
	err := r.db.QueryRow(query, name, description, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProjectRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM projects WHERE id = $1`, id)
	return err
}
