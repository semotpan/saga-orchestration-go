package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // load PG driver
)

// StoreProps contain the postgres settings
type StoreProps struct {
	Host     string
	Port     string
	User     string
	Password string
	Dbname   string
}

type Store struct {
	conn *sql.DB
}

// NewStore constructor
func NewStore(sp StoreProps) (*Store, error) {
	psqlURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		sp.Host, sp.Port, sp.User, sp.Password, sp.Dbname)

	db, err := sql.Open("postgres", psqlURL)
	if err != nil || db.Ping() != nil {
		log.Fatalf("failed to connect to database %v", err)
		return nil, err
	}
	return &Store{conn: db}, nil
}

func (s Store) Transact(ctx context.Context, f func(tx *sql.Tx) (interface{}, error)) (interface{}, error) {
	tx, e := s.conn.BeginTx(ctx, nil)
	// Any error here is non-retryable
	if e != nil {
		return nil, e
	}

	// Defer a rollback in case anything fails.
	defer func() {
		// NOTE: should not have effect if the transaction has been committed
		_ = tx.Rollback()
	}()

	val, err := f(tx)

	if err != nil {
		return nil, err
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return val, nil //TODO use retry and Optimistic Locking
}

/*
func retryable(wrapped func() (interface{}, error), onError func(Error) FutureNil) (ret interface{}, e error) {
	for {
		ret, e = wrapped()

		// No error means success!
		if e == nil {
			return
		}

		// Check if the error chain contains an
		// fdb.Error
		var ep Error
		if errors.As(e, &ep) {
			e = onError(ep).Get()
		}

		// If OnError returns an error, then it's not
		// retryable; otherwise take another pass at things
		if e != nil {
			return
		}
	}
}
*/
