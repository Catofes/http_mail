package main

import (
	"context"
	"database/sql"
)

type row map[string]any

type statement struct {
	sql  string
	args []any
}

type store interface {
	query(context.Context, string, ...any) ([]row, error)
	execute(context.Context, string, ...any) error
	transaction(context.Context, ...statement) error
}

type postgresStore struct{ db *sql.DB }

func (s *postgresStore) query(ctx context.Context, query string, args ...any) ([]row, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apiError(1)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, apiError(1)
	}
	// The Python database wrapper returned None when fetchall() was empty.
	var result []row
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, apiError(1)
		}
		r := row{}
		for i, name := range columns {
			if data, ok := values[i].([]byte); ok {
				values[i] = string(data)
			}
			r[name] = values[i]
		}
		result = append(result, r)
	}
	if rows.Err() != nil {
		return nil, apiError(1)
	}
	return result, nil
}

func (s *postgresStore) execute(ctx context.Context, query string, args ...any) error {
	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return apiError(1)
	}
	return nil
}

func (s *postgresStore) transaction(ctx context.Context, statements ...statement) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return apiError(1)
	}
	defer tx.Rollback()
	for _, st := range statements {
		if _, err := tx.ExecContext(ctx, st.sql, st.args...); err != nil {
			return apiError(1)
		}
	}
	if err := tx.Commit(); err != nil {
		return apiError(1)
	}
	return nil
}
