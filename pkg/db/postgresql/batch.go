package postgresql

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	QueryTimeout = 1 * time.Minute
	MaxRetries   = 2

	ErrorNoConnFree = "no connection adquirable"
)

type QueryBatch struct {
	ctx     context.Context
	pgxPool *pgxpool.Pool
	batch   *pgx.Batch
	size    int
}

func NewQueryBatch(ctx context.Context, pgxPool *pgxpool.Pool, batchSize int) *QueryBatch {
	return &QueryBatch{
		ctx:     ctx,
		pgxPool: pgxPool,
		batch:   &pgx.Batch{},
		size:    batchSize,
	}
}

func (q *QueryBatch) IsReadyToPersist() bool {
	return q.batch.Len() >= q.size
}

func (q *QueryBatch) AddQuery(query string, args ...interface{}) {
	q.batch.Queue(query, args...)
}

func (q *QueryBatch) Len() int {
	return q.batch.Len()
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

	// begin pgx.Tx
	logEntry.Trace("beginning a new transaction to store the batched queries")
	tx, err := q.pgxPool.Begin(ctx)
	if err != nil {
		return err
	}
	// Add batch to TX
	logEntry.Trace("sending batch over transaction")
	batchResults := tx.SendBatch(ctx, q.batch)

	// Exec the queries
	var qerr error
	var rows pgx.Rows
	var cnt int
	for qerr == nil {
		rows, qerr = batchResults.Query()
		rows.Close()
		cnt++
	}
	logEntry.Trace("readed all the result of the queries inside the batch")
	// check if there was any error
	if qerr.Error() != noQueryResult {
		log.Errorf("unable to persist betch because an error on row %d \n %+v\n", cnt, rows)
		return err
	}
	return tx.Commit(q.ctx)
}

func (q *QueryBatch) cleanBatch() {
	q.batch = &pgx.Batch{}
}
