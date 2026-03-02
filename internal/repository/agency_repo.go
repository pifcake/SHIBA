package repository

import (
	"context"
	"errors"
	"fmt"

	"SHIBA/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AgencyRepository struct {
	pool *pgxpool.Pool
}

func NewAgencyRepository(pool *pgxpool.Pool) *AgencyRepository {
	return &AgencyRepository{pool: pool}
}

func (r *AgencyRepository) Create(ctx context.Context, p *domain.AgencyProfile) error {
	q := `INSERT INTO agency_profiles (id, user_id, company_name, description, phone, logo_url, status)
		  VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.pool.Exec(ctx, q, p.ID, p.UserID, p.CompanyName, p.Description, p.Phone, p.LogoURL, p.Status)
	return err
}

func (r *AgencyRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.AgencyProfile, error) {
	q := `SELECT id, user_id, company_name, description, phone, logo_url, status,
		  approved_by, approved_at, created_at, updated_at
		  FROM agency_profiles WHERE user_id=$1`
	row := r.pool.QueryRow(ctx, q, userID)
	return scanAgencyProfile(row)
}

func (r *AgencyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AgencyProfile, error) {
	q := `SELECT id, user_id, company_name, description, phone, logo_url, status,
		  approved_by, approved_at, created_at, updated_at
		  FROM agency_profiles WHERE id=$1`
	row := r.pool.QueryRow(ctx, q, id)
	return scanAgencyProfile(row)
}

func (r *AgencyRepository) Update(ctx context.Context, p *domain.AgencyProfile) error {
	q := `UPDATE agency_profiles SET company_name=$1, description=$2, phone=$3, logo_url=$4, updated_at=NOW()
		  WHERE id=$5`
	_, err := r.pool.Exec(ctx, q, p.CompanyName, p.Description, p.Phone, p.LogoURL, p.ID)
	return err
}

func (r *AgencyRepository) Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) error {
	q := `UPDATE agency_profiles SET status='active', approved_by=$1, approved_at=NOW(), updated_at=NOW()
		  WHERE id=$2`
	ct, err := r.pool.Exec(ctx, q, approvedBy, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AgencyRepository) Reject(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE agency_profiles SET status='blocked', updated_at=NOW() WHERE id=$1`
	ct, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AgencyRepository) GetPending(ctx context.Context, limit, offset int) ([]domain.AgencyProfile, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM agency_profiles WHERE status='pending_approval'`).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := `SELECT id, user_id, company_name, description, phone, logo_url, status,
		  approved_by, approved_at, created_at, updated_at
		  FROM agency_profiles WHERE status='pending_approval'
		  ORDER BY created_at ASC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var profiles []domain.AgencyProfile
	for rows.Next() {
		p, err := scanAgencyProfile(rows)
		if err != nil {
			return nil, 0, err
		}
		profiles = append(profiles, *p)
	}
	return profiles, total, rows.Err()
}

func scanAgencyProfile(row scanner) (*domain.AgencyProfile, error) {
	var p domain.AgencyProfile
	err := row.Scan(
		&p.ID, &p.UserID, &p.CompanyName, &p.Description, &p.Phone, &p.LogoURL, &p.Status,
		&p.ApprovedBy, &p.ApprovedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan agency profile: %w", err)
	}
	return &p, nil
}
