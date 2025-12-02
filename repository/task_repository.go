package repository

import (
	"context"
	"database/sql"
	"fmt"
	"go-realtime-workspace/models"
	"time"
)

// TaskRepository handles task database operations.
type TaskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new task repository.
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create creates a new task.
func (r *TaskRepository) Create(ctx context.Context, userID string, req models.CreateTaskRequest) (*models.Task, error) {
	query := `
		INSERT INTO tasks (user_id, title, description, priority, due_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, title, description, status, priority, due_date, created_at, updated_at, completed_at
	`

	task := &models.Task{}
	err := r.db.QueryRowContext(
		ctx, query,
		userID, req.Title, req.Description, req.Priority, req.DueDate,
	).Scan(
		&task.ID, &task.UserID, &task.Title, &task.Description,
		&task.Status, &task.Priority, &task.DueDate,
		&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creating task: %w", err)
	}

	return task, nil
}

// GetByID retrieves a task by ID.
func (r *TaskRepository) GetByID(ctx context.Context, id string) (*models.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, priority, due_date, created_at, updated_at, completed_at
		FROM tasks WHERE id = $1
	`

	task := &models.Task{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.UserID, &task.Title, &task.Description,
		&task.Status, &task.Priority, &task.DueDate,
		&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting task: %w", err)
	}

	return task, nil
}

// GetByUserID retrieves all tasks for a user.
func (r *TaskRepository) GetByUserID(ctx context.Context, userID string, status string) ([]models.Task, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `
			SELECT id, user_id, title, description, status, priority, due_date, created_at, updated_at, completed_at
			FROM tasks WHERE user_id = $1 AND status = $2
			ORDER BY created_at DESC
		`
		args = []interface{}{userID, status}
	} else {
		query = `
			SELECT id, user_id, title, description, status, priority, due_date, created_at, updated_at, completed_at
			FROM tasks WHERE user_id = $1
			ORDER BY created_at DESC
		`
		args = []interface{}{userID}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error getting tasks: %w", err)
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID, &task.UserID, &task.Title, &task.Description,
			&task.Status, &task.Priority, &task.DueDate,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Update updates a task.
func (r *TaskRepository) Update(ctx context.Context, id string, req models.UpdateTaskRequest) (*models.Task, error) {
	query := `
		UPDATE tasks
		SET title = COALESCE(NULLIF($1, ''), title),
		    description = COALESCE(NULLIF($2, ''), description),
		    status = COALESCE(NULLIF($3, ''), status),
		    priority = COALESCE(NULLIF($4, ''), priority),
		    due_date = COALESCE($5, due_date),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
		RETURNING id, user_id, title, description, status, priority, due_date, created_at, updated_at, completed_at
	`

	task := &models.Task{}
	err := r.db.QueryRowContext(
		ctx, query,
		req.Title, req.Description, req.Status, req.Priority, req.DueDate, id,
	).Scan(
		&task.ID, &task.UserID, &task.Title, &task.Description,
		&task.Status, &task.Priority, &task.DueDate,
		&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error updating task: %w", err)
	}

	return task, nil
}

// Delete deletes a task.
func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// GetDueSoon retrieves tasks that are due within the specified duration.
func (r *TaskRepository) GetDueSoon(ctx context.Context, userID string, within time.Duration) ([]models.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, priority, due_date, created_at, updated_at, completed_at
		FROM tasks 
		WHERE user_id = $1 
		  AND status != 'completed'
		  AND due_date IS NOT NULL
		  AND due_date <= $2
		ORDER BY due_date ASC
	`

	dueBy := time.Now().Add(within)
	rows, err := r.db.QueryContext(ctx, query, userID, dueBy)
	if err != nil {
		return nil, fmt.Errorf("error getting due tasks: %w", err)
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID, &task.UserID, &task.Title, &task.Description,
			&task.Status, &task.Priority, &task.DueDate,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
