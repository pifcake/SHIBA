package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"SHIBA/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ComplaintRepository struct {
	pool *pgxpool.Pool
}

func NewComplaintRepository(pool *pgxpool.Pool) *ComplaintRepository {
	return &ComplaintRepository{pool: pool}
}

func (r *ComplaintRepository) Create(ctx context.Context, c *domain.Complaint) error {
	q := `INSERT INTO complaints (id, reporter_id, target_type, target_id, reason_category, description, status)
		  VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.pool.Exec(ctx, q, c.ID, c.ReporterID, c.TargetType, c.TargetID, c.ReasonCategory, c.Description, c.Status)
	return err
}

func (r *ComplaintRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Complaint, error) {
	q := `SELECT id, reporter_id, target_type, target_id, reason_category, description,
		  status, reviewed_by, resolution_note, resolved_at, created_at
		  FROM complaints WHERE id=$1`
	row := r.pool.QueryRow(ctx, q, id)
	return scanComplaint(row)
}

type ComplaintFilter struct {
	Status string
	Limit  int
	Offset int
}

func (r *ComplaintRepository) List(ctx context.Context, f ComplaintFilter) ([]domain.Complaint, int, error) {
	conditions := []string{"1=1"}
	args := []any{}
	argIdx := 1

	if f.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status=$%d", argIdx))
		args = append(args, f.Status)
		argIdx++
	}
	where := "WHERE " + strings.Join(conditions, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM complaints %s", where), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if f.Limit == 0 {
		f.Limit = 20
	}
	args = append(args, f.Limit, f.Offset)
	q := fmt.Sprintf(`SELECT id, reporter_id, target_type, target_id, reason_category, description,
		  status, reviewed_by, resolution_note, resolved_at, created_at
		  FROM complaints %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var complaints []domain.Complaint
	for rows.Next() {
		c, err := scanComplaint(rows)
		if err != nil {
			return nil, 0, err
		}
		complaints = append(complaints, *c)
	}
	return complaints, total, rows.Err()
}

func (r *ComplaintRepository) Update(ctx context.Context, id uuid.UUID, status domain.ComplaintStatus, reviewedBy *uuid.UUID, note string) error {
	q := `UPDATE complaints SET status=$1, reviewed_by=$2, resolution_note=$3,
		  resolved_at=CASE WHEN $1 IN ('resolved','dismissed') THEN NOW() ELSE NULL END
		  WHERE id=$4`
	ct, err := r.pool.Exec(ctx, q, status, reviewedBy, note, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanComplaint(row scanner) (*domain.Complaint, error) {
	var c domain.Complaint
	err := row.Scan(
		&c.ID, &c.ReporterID, &c.TargetType, &c.TargetID, &c.ReasonCategory,
		&c.Description, &c.Status, &c.ReviewedBy, &c.ResolutionNote, &c.ResolvedAt, &c.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan complaint: %w", err)
	}
	return &c, nil
}
