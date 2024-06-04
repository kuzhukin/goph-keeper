package controller

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kuzhukin/goph-keeper/internal/server/handler"
	"github.com/kuzhukin/goph-keeper/internal/zlog"
)

var _ handler.Storage = &Controller{}
var _ handler.Authenticator = &Controller{}
var _ handler.Registrator = &Controller{}

type Controller struct {
	db *sql.DB
}

func StartNewController(dataSourceName string) (*Controller, error) {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("sql open dataSourceName=%s, err=%w", dataSourceName, err)
	}

	ctrl := &Controller{db: db}
	if err := ctrl.init(); err != nil {
		return nil, fmt.Errorf("init, err=%w", err)
	}

	return ctrl, nil
}

func (c *Controller) Stop() error {
	if err := c.db.Close(); err != nil {
		return fmt.Errorf("db close err=%w", err)
	}

	return nil
}

func (c *Controller) init() error {
	createTableQueries := []string{
		createBinaryDataTableQuery,
		createUsersTableQuery,
	}

	for _, q := range createTableQueries {
		if err := c.exec(q); err != nil {
			return fmt.Errorf("exec query=%s, err=%w", q, err)
		}
	}

	return nil
}

func (c *Controller) exec(query string) error {
	const createTablesTimeout = time.Second * 10

	ctx, cancel := context.WithTimeout(context.Background(), createTablesTimeout)
	defer cancel()

	_, err := c.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("exec create table query, err=%w", err)
	}

	return nil
}

func (c *Controller) Save(ctx context.Context, u *handler.User, d *handler.Data) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := c.db.BeginTx(ctxWithTimeout, nil)
	if err != nil {
		return err
	}
	defer recoverAndRollBack(tx)

	if err = checkUserInTransaction(ctx, tx, u); err != nil {
		return err
	}

	query := prepareNewDataQuery(u.Login, d.Key, d.Data)

	err = doTransactionExec(ctx, tx, query)
	if err != nil {
		return fmt.Errorf("add user=%s data=%s, err=%w", u.Login, d.Data, err)
	}

	return tx.Commit()
}

func (c *Controller) Update(ctx context.Context, u *handler.User, d *handler.Data) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := c.db.BeginTx(ctxWithTimeout, nil)
	if err != nil {
		return err
	}
	defer recoverAndRollBack(tx)

	if err = checkUserInTransaction(ctx, tx, u); err != nil {
		return err
	}

	storedData, err := getDataInTransaction(ctx, tx, u, d.Key)
	if err != nil {
		return err
	}

	if storedData.Revision != d.Revision {
		return fmt.Errorf("user=%s data=%s err=%w", u.Login, d.Key, handler.ErrBadRevision)
	}

	if storedData.Data == d.Data {
		// doesn't need in changes
		return nil
	}

	err = doTransactionExec(ctx, tx, prepareUpdateDataQuery(u.Login, d.Key, d.Data, d.Revision))
	if err != nil {
		return fmt.Errorf("do update user=%s, data=%s, rev=%d, err=%w", u.Login, d.Key, d.Revision, err)
	}

	return tx.Commit()
}

func (c *Controller) Load(ctx context.Context, u *handler.User, dataKey string) (*handler.Data, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := c.db.BeginTx(ctxWithTimeout, nil)
	if err != nil {
		return nil, err
	}
	defer recoverAndRollBack(tx)

	if err = checkUserInTransaction(ctx, tx, u); err != nil {
		return nil, err
	}

	d, err := getDataInTransaction(ctx, tx, u, dataKey)
	if err != nil {
		return nil, err
	}

	return d, tx.Commit()
}

func (c *Controller) Delete(ctx context.Context, u *handler.User, d *handler.Data) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := c.db.BeginTx(ctxWithTimeout, nil)
	if err != nil {
		return err
	}
	defer recoverAndRollBack(tx)

	if err = checkUserInTransaction(ctx, tx, u); err != nil {
		return err
	}

	err = deleteDataInTransaction(ctx, tx, u, d.Key)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Register new user
// return:
// * nil - user registred successfully
// * ErrUserAlreadyRegistred - user with password already registred
// * ErrPasswordConflict - user registred with other password
// * otherErr - internal
func (c *Controller) Register(ctx context.Context, login string, password string) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := c.db.BeginTx(ctxWithTimeout, nil)
	if err != nil {
		return err
	}
	defer recoverAndRollBack(tx)

	query := prepareGetUserQuery(login)

	user, err := doTransactionQuery(ctxWithTimeout, tx, query, func(rows *sql.Rows) (*handler.User, error) {
		if !rows.Next() {
			return nil, handler.ErrUnknownUser
		}

		user := &handler.User{}
		if err := rows.Scan(&user.Login, &user.Password); err != nil {
			return nil, err
		}

		return user, nil
	})
	if err == nil {
		if user.Password != password {
			return handler.ErrBadPassword
		}

		return nil
	}

	if !errors.Is(err, handler.ErrUnknownUser) {
		return err
	}

	query = prepareCreateUserQuery(login, password)

	err = doTransactionExec(ctxWithTimeout, tx, query)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func recoverAndRollBack(tx *sql.Tx) {
	err := recover()
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			zlog.Logger().Errorf("tx rollback failed, err=%w", err)
		}
	}
}

func (c *Controller) Authenticate(ctx context.Context, user string, password string) error {
	// TODO: скорей всего это не нужный метод
	return nil
}

func checkUserInTransaction(
	ctx context.Context,
	tx *sql.Tx,
	u *handler.User,
) error {
	userQuery := prepareGetUserQuery(u.Login)

	storedUser, err := doTransactionQuery(ctx, tx, userQuery, func(rows *sql.Rows) (*handler.User, error) {
		storedUser := &handler.User{}

		if !rows.Next() {
			return nil, fmt.Errorf("user=%s isn't registred err=%w", u.Login, handler.ErrUnknownUser)
		}

		if err := rows.Scan(&storedUser.Login, &storedUser.Password); err != nil {
			return nil, err
		}

		return storedUser, nil
	})
	if err != nil {
		return err
	}

	if storedUser.Password != u.Password {
		return fmt.Errorf("user=%s, err=%w", u.Login, handler.ErrBadPassword)
	}

	return err
}

func getDataInTransaction(
	ctx context.Context,
	tx *sql.Tx,
	u *handler.User,
	dataKey string,
) (*handler.Data, error) {
	return doTransactionQuery(ctx, tx, prepareGetDataQuery(u.Login, dataKey), func(rows *sql.Rows) (*handler.Data, error) {
		if !rows.Next() {
			return nil, fmt.Errorf("key=%s, err=%w", dataKey, handler.ErrDataNotFound)
		}

		storedData := &handler.Data{Key: dataKey}
		if err := rows.Scan(&storedData.Data, &storedData.Revision); err != nil {
			return nil, err
		}

		return storedData, nil
	})
}

func deleteDataInTransaction(
	ctx context.Context,
	tx *sql.Tx,
	u *handler.User,
	dataKey string,
) error {
	return doTransactionExec(ctx, tx, prepareDeleteDataQuery(u.Login, dataKey))
}

// ----------------------------------------------------------------------------------------------
// -------------------------------------- Internal Methods --------------------------------------
// ----------------------------------------------------------------------------------------------

var tryingIntervals = []time.Duration{
	time.Millisecond * 100,
	time.Millisecond * 300,
	time.Millisecond * 500,
}

func doQuery[T any](ctx context.Context, queryFunc func(context.Context) (T, error)) (T, error) {
	var commonErr error
	max := len(tryingIntervals)

	var result T
	var err error

	for trying := 0; trying <= max; trying++ {
		result, err = queryFunc(ctx)
		if err != nil {
			commonErr = errors.Join(commonErr, err)

			if trying < max && isRetriableError(err) {
				time.Sleep(tryingIntervals[trying])
				continue
			}

			return result, commonErr
		}

		return result, nil
	}

	return result, commonErr
}

func isRetriableError(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)
}

// func isNotUniqueError(err error) bool {
// 	var pgErr *pgconn.PgError

// 	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
// }

func doTransactionExec(ctx context.Context, tx *sql.Tx, query *query) error {
	_, err := doQuery(ctx, func(ctx context.Context) (sql.Result, error) {
		return tx.ExecContext(ctx, query.request, query.args...)
	})

	return err
}

func doTransactionQuery[T any](ctx context.Context, tx *sql.Tx, query *query, parseFun func(rows *sql.Rows) (T, error)) (T, error) {
	var result T

	rows, err := doQuery(ctx, func(ctx context.Context) (*sql.Rows, error) {
		return tx.QueryContext(ctx, query.request, query.args...)
	})
	if err != nil {
		return result, nil
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			zlog.Logger().Errorf("rows close err=%s", err)
		}
	}()

	result, err = parseFun(rows)
	if err != nil {
		return result, err
	}

	if err := rows.Err(); err != nil {
		return result, err
	}

	return result, nil
}
