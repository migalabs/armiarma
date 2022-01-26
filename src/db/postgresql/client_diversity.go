package postgresql

import (
	"fmt"
	"time"

	"github.com/migalabs/armiarma/src/db/models"
	"github.com/pkg/errors"
)

var (
	createClientDiversity = `
	CREATE TABLE IF NOT EXISTS t_client_diversity(
		f_snapshot_timestamp TIMESTAMP,
		f_prysm BIGINT,
		f_lighthouse BIGINT,
		f_teku BIGINT,
		f_nimbus BIGINT,
		f_grandine BIGINT,
		f_lodestar BIGINT,
		f_others BIGINT, 
		
		PRIMARY KEY (f_snapshot_timestamp)
	);
	`
	insertClientDiversitySnapshot = `
	INSERT INTO t_client_diversity(
		f_snapshot_timestamp,
		f_prysm,
		f_lighthouse,
		f_teku,
		f_nimbus,
		f_grandine,
		f_lodestar,
		f_others)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (f_snapshot_timestamp)
	DO UPDATE SET
		f_snapshot_timestamp=EXCLUDED.f_snapshot_timestamp,
		f_prysm=EXCLUDED.f_prysm,
		f_lighthouse=EXCLUDED.f_lighthouse,
		f_teku=EXCLUDED.f_teku,
		f_nimbus=EXCLUDED.f_nimbus,
		f_grandine=EXCLUDED.f_grandine,
		f_lodestar=EXCLUDED.f_lodestar,
		f_others=EXCLUDED.f_others
	`
)

func (p *PostgresDBService) createClientDiversityTable() error {
	log.Debugf("creating client diversity table in psql")
	_, err := p.psqlPool.Exec(p.ctx,
		createClientDiversity,
	)
	if err != nil {
		return errors.Wrap(err, "unable to create client diversity table")
	}
	return nil
}

func (p *PostgresDBService) StoreClientDiversitySnapshot(cliDiver models.ClientDiversity) error {
	log.Debugf("adding new client diversity item in psql")
	_, err := p.psqlPool.Exec(
		p.ctx,
		insertClientDiversitySnapshot,
		cliDiver.Timestamp,
		cliDiver.Prysm,
		cliDiver.Lighthouse,
		cliDiver.Teku,
		cliDiver.Nimbus,
		cliDiver.Grandine,
		cliDiver.Lodestar,
		cliDiver.Others,
	)
	if err != nil {
		errors.Wrap(err, "error storing client diversity snapshot in postgresql")
	}
	return nil
}

// So far not used since it's just for exporting
// Doesn't make much sense to add it to the crawler (no idea why would we need to access the snapshot of a given time)
func (p *PostgresDBService) LoadClientDiversitySnapshot(qTime time.Time) (models.ClientDiversity, error) {
	log.Debugf("Loading client diversity of time %s", qTime)
	cliDist := models.NewClientDiversity()
	err := p.psqlPool.QueryRow(
		p.ctx,
		"SELECT * FROM t_client_diversity WHERE f_snapshot_timestamp=$1",
		qTime,
	).Scan(
		&cliDist.Timestamp,
		&cliDist.Prysm,
		&cliDist.Lighthouse,
		&cliDist.Teku,
		&cliDist.Nimbus,
		&cliDist.Grandine,
		&cliDist.Lodestar,
		&cliDist.Others,
	)
	if err != nil {
		return cliDist, errors.Wrap(err, fmt.Sprintf("error loading client distribution of %s", qTime))
	}
	return cliDist, nil
}
