package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"jobbe.service/internal/validator"
)

type Vacancy struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Company   string    `json:"company"`
	Active    bool      `json:"active,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	Version   int32     `json:"version"`
}
type VacancyModel struct {
	DB *sql.DB
}

func (m VacancyModel) Insert(vacancy *Vacancy) error {
	query := `
		INSERT INTO vacancies (title, company, tags)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	args := []any{vacancy.Title, vacancy.Company, pq.Array(vacancy.Tags)}

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&vacancy.ID, &vacancy.CreatedAt, &vacancy.Version)
}

func (m VacancyModel) Get(id int64) (*Vacancy, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, title, company, tags, version
		FROM vacancies
		WHERE id = $1`

	var vacancy Vacancy

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&vacancy.ID,
		&vacancy.CreatedAt,
		&vacancy.Title,
		&vacancy.Company,
		pq.Array(&vacancy.Tags),
		&vacancy.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &vacancy, nil
}

func (m VacancyModel) GetAll(title string, tags []string, filters Filters) ([]*Vacancy, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) over(), id, created_at, title, company, tags, version
		FROM vacancies
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, title, pq.Array(tags), filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	vacancies := []*Vacancy{}

	for rows.Next() {
		var vacancy Vacancy

		err := rows.Scan(
			&vacancy.ID,
			&vacancy.CreatedAt,
			&vacancy.Title,
			&vacancy.Company,
			pq.Array(&vacancy.Tags),
			&vacancy.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		vacancies = append(vacancies, &vacancy)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return vacancies, metadata, nil
}

func (m VacancyModel) Update(vacancy *Vacancy) error {
	query := `
		UPDATE vacancies
		SET title = $1, company = $2, tags = $3, active = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	args := []any{
		vacancy.Title,
		vacancy.Company,
		pq.Array(vacancy.Tags),
		vacancy.Active,
		vacancy.ID,
		vacancy.Version,
	}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&vacancy.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m VacancyModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM vacancies
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func ValidateVacancy(v *validator.Validator, vacancy *Vacancy) {
	v.Check(vacancy.Title != "", "title", "must be provided")
	v.Check(len(vacancy.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(vacancy.Company != "", "company", "must be provided")
	v.Check(len(vacancy.Company) <= 500, "company", "must not be more than 500 bytes long")
	v.Check(vacancy.Tags != nil, "tags", "must be provided")
	v.Check(len(vacancy.Tags) >= 1, "tags", "must contain at least 1 tag")
	v.Check(len(vacancy.Tags) <= 5, "tags", "must not contain more than 5 tags")
	v.Check(validator.Unique(vacancy.Tags), "tags", "must not contain duplicate values")
}
