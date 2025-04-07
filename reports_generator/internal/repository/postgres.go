package repository

import (
    "context"
    "fmt"
    "github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
    pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
    return &PostgresRepo{pool: pool}
}

func (r *PostgresRepo) GetUserIDs(ctx context.Context) ([]int, error) {
    const query = `SELECT id FROM users`
    rows, err := r.pool.Query(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("database query failed: %v", err)
    }
    defer rows.Close()

    var ids []int
    for rows.Next() {
        var id int
        if err := rows.Scan(&id); err != nil {
            return nil, fmt.Errorf("row scan failed: %v", err)
        }
        ids = append(ids, id)
    }

    return ids, nil
}