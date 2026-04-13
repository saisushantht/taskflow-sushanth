package repository

import (
	"database/sql"
	"fmt"
	"taskflow/internal/model"
	"time"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(title, description, status, priority, projectID string, assigneeID *string, dueDate *time.Time) (*model.Task, error) {
	t := &model.Task{}
	query := `
		INSERT INTO tasks (title, description, status, priority, project_id, assignee_id, due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, description, status, priority, project_id, assignee_id, due_date, created_at, updated_at`
	err := r.db.QueryRow(query, title, description, status, priority, projectID, assigneeID, dueDate).
		Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.ProjectID, &t.AssigneeID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TaskRepository) ListByProject(projectID, status, assigneeID string, page, limit int) ([]model.Task, int, error) {
	conditions := []interface{}{projectID}
	where := "WHERE project_id = $1"
	idx := 2

	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		conditions = append(conditions, status)
		idx++
	}
	if assigneeID != "" {
		where += fmt.Sprintf(" AND assignee_id = $%d", idx)
		conditions = append(conditions, assigneeID)
		idx++
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks %s", where)
	r.db.QueryRow(countQuery, conditions...).Scan(&total)

	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT id, title, description, status, priority, project_id, assignee_id, due_date, created_at, updated_at
		FROM tasks %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, idx, idx+1)
	conditions = append(conditions, limit, offset)

	rows, err := r.db.Query(query, conditions...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.ProjectID, &t.AssigneeID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, nil
}

func (r *TaskRepository) FindByID(id string) (*model.Task, error) {
	t := &model.Task{}
	query := `SELECT id, title, description, status, priority, project_id, assignee_id, due_date, created_at, updated_at FROM tasks WHERE id = $1`
	err := r.db.QueryRow(query, id).
		Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.ProjectID, &t.AssigneeID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TaskRepository) Update(id, title, description, status, priority string, assigneeID *string, dueDate *time.Time) (*model.Task, error) {
	t := &model.Task{}
	query := `
		UPDATE tasks
		SET title=$1, description=$2, status=$3, priority=$4, assignee_id=$5, due_date=$6, updated_at=NOW()
		WHERE id=$7
		RETURNING id, title, description, status, priority, project_id, assignee_id, due_date, created_at, updated_at`
	err := r.db.QueryRow(query, title, description, status, priority, assigneeID, dueDate, id).
		Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.ProjectID, &t.AssigneeID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TaskRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM tasks WHERE id = $1`, id)
	return err
}

func (r *TaskRepository) GetStats(projectID string) (map[string]interface{}, error) {
	stats := map[string]interface{}{}

	statusRows, err := r.db.Query(`
		SELECT status, COUNT(*) FROM tasks WHERE project_id = $1 GROUP BY status`, projectID)
	if err != nil {
		return nil, err
	}
	defer statusRows.Close()

	byStatus := map[string]int{}
	for statusRows.Next() {
		var s string
		var c int
		statusRows.Scan(&s, &c)
		byStatus[s] = c
	}

	assigneeRows, err := r.db.Query(`
		SELECT u.name, COUNT(t.id)
		FROM tasks t
		JOIN users u ON u.id = t.assignee_id
		WHERE t.project_id = $1 AND t.assignee_id IS NOT NULL
		GROUP BY u.name`, projectID)
	if err != nil {
		return nil, err
	}
	defer assigneeRows.Close()

	byAssignee := map[string]int{}
	for assigneeRows.Next() {
		var name string
		var count int
		assigneeRows.Scan(&name, &count)
		byAssignee[name] = count
	}

	stats["by_status"] = byStatus
	stats["by_assignee"] = byAssignee
	return stats, nil
}
