package data

import (
	"context"
	"database/sql"
	"time"

	"jobbe.service/internal/validator"
)

type Subscriber struct {
	ID        int64
	UserID    int64
	Email     string
	Tag       string
	CreatedAt time.Time
}
type SubscriberModel struct {
	DB *sql.DB
}

func (m SubscriberModel) Insert(subscriber *Subscriber) error {
	query := `
		INSERT INTO subscribers (userId, tag)
		VALUES ($1, $2)
		RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	args := []any{subscriber.UserID, subscriber.Tag}

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&subscriber.ID, &subscriber.CreatedAt)
}

func (m SubscriberModel) GetAllById(id int64) ([]*Subscriber, error) {
	query := `
		SELECT id, userid, tag, created_at 
		FROM subscribers 
		WHERE userid = $1;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	subscribers := []*Subscriber{}

	for rows.Next() {
		var subscriber Subscriber

		err := rows.Scan(
			&subscriber.ID,
			&subscriber.UserID,
			&subscriber.Tag,
			&subscriber.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		subscribers = append(subscribers, &subscriber)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subscribers, nil
}

func (m SubscriberModel) GetAllByTag(tag string) ([]*Subscriber, error) {
	query := `
		SELECT subscribers.id, userid, email, tag, subscribers.created_at 
		FROM subscribers 
		JOIN users on subscribers.userid = users.id
		WHERE (to_tsvector('simple', tag) @@ plainto_tsquery('simple', $1));`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, tag)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	subscribers := []*Subscriber{}

	for rows.Next() {
		var subscriber Subscriber

		err := rows.Scan(
			&subscriber.ID,
			&subscriber.UserID,
			&subscriber.Email,
			&subscriber.Tag,
			&subscriber.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		subscribers = append(subscribers, &subscriber)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subscribers, nil
}

func (m SubscriberModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM subscribers
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

func ValidateSubscriber(v *validator.Validator, subscriber *Subscriber) {
	v.Check(subscriber.Tag != "", "tag", "must be provided")
	v.Check(len(subscriber.Tag) <= 500, "tag", "must not be more than 500 bytes long")
}
