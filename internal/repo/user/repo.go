package pguser

import (
	"context"
	"errors"

	"github.com/defany/db/v2/postgres"
	slerr "github.com/defany/slogger/pkg/err"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
)

type Repo struct {
	db postgres.Postgres
	qb sqlbuilder.Flavor
}

func New(db postgres.Postgres) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) CreateIfNotExists(ctx context.Context, userID int64) error {
	q := `insert into users (id) values ($1) on conflict do nothing`

	_, err := r.db.Exec(ctx, q, userID)
	if err != nil {
		return slerr.WithSource(err)
	}

	return nil
}

func (r *Repo) IsNotificationsEnabled(ctx context.Context, userID int64) (bool, error) {
	q := `select is_notifications_enabled from users where id = $1`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return false, slerr.WithSource(err)
	}

	isNotificationsEnabled, err := pgx.CollectOneRow(rows, pgx.RowTo[bool])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, slerr.WithSource(err)
	}

	return isNotificationsEnabled, nil
}

func (r *Repo) ToggleNotifications(ctx context.Context, userID int64, isEnabled bool) (err error) {
	q := `update users set is_notifications_enabled = $1 where id = $2`

	_, err = r.db.Exec(ctx, q, isEnabled, userID)
	if err != nil {
		return slerr.WithSource(err)
	}

	return nil
}

func (r *Repo) FetchNotificationReceivers(ctx context.Context) ([]int64, error) {
	q := `select id from users where is_notifications_enabled = true`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, slerr.WithSource(err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return nil, slerr.WithSource(err)
	}

	return ids, nil
}
