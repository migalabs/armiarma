package redshift

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	_ "github.com/lib/pq"
)

var (
	QueryTimeout = 5 * time.Minute
	MaxRetries   = 2

	ErrorNoConnFree = "no connection adquirable"
	ErrorNoRows     = "sql: no rows in result set"
)

type QueryBatch struct {
	ctx      context.Context
	sqlDB    *sql.DB
	queries  []string
	args     [][]interface{}
	size     int
}

func NewQueryBatch(ctx context.Context, sqlDB *sql.DB, batchSize int) *QueryBatch {
	return &QueryBatch{
		ctx:      ctx,
		sqlDB:    sqlDB,
		queries:  make([]string, 0, batchSize),
		args:     make([][]interface{}, 0, batchSize),
		size:     batchSize,
	}
}

func (q *QueryBatch) IsReadyToPersist() bool {
	return len(q.queries) >= q.size
}

func (q *QueryBatch) AddQuery(query string, args ...interface{}) {
	q.queries = append(q.queries, query)
	q.args = append(q.args, args)
}

func (q *QueryBatch) Len() int {
	return len(q.queries)
}

func (q *QueryBatch) PersistBatch() error {
	logEntry := log.WithFields(log.Fields{
		"mod": "batch-persister",
	})
	logEntry.Debugf("persisting batch of queries with len(%d)", q.Len())
	var err error
persistRetryLoop:
	for i := 0; i <= MaxRetries; i++ {
		t := time.Now()
		err = q.persistBatch()
		duration := time.Since(t)
		switch err {
		case nil:
			logEntry.Debugf("persisted %d queries in %s seconds", q.Len(), duration)
			break persistRetryLoop
		default:
			logEntry.Debugf("attempt numb %d failed %s", i+1, err.Error())
		}
	}
	q.cleanBatch()
	return errors.Wrap(err, "unable to persist batch query")
}

func (q *QueryBatch) persistBatch() error {
	logEntry := log.WithFields(log.Fields{
		"mod": "batch-persister",
	})
	// if batch len == 0, don't even query
	if q.Len() == 0 {
		logEntry.Debug("skipping batch-query, no queries to persist")
		return nil
	}

	// generate a timeout to the batch-persisting
	ctx, cancel := context.WithTimeout(q.ctx, QueryTimeout)
	defer cancel()

	tx, err := q.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for i, query := range q.queries {
		_, err := tx.ExecContext(ctx, query, q.args[i]...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (q *QueryBatch) cleanBatch() {
	q.queries = q.queries[:0]
	q.args = q.args[:0]
}
