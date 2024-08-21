package redshift

import (
	"time"

	eth "github.com/migalabs/armiarma/pkg/networks/ethereum"
	log "github.com/sirupsen/logrus"
	_ "github.com/lib/pq"
)


// DropEthereumAttestationsTable drops the eth_attestations table
func (c *DBClient) DropEthereumAttestationsTable() error {
	log.Info("dropping the eth_attestations table")
	_, err := c.redshiftDB.ExecContext(
		c.ctx,
		`
		DROP TABLE IF EXISTS eth_attestations;
		`)
	return err
}

// InitEthereumAttestationsTable initializes the eth_attestations table in Redshift
func (c *DBClient) InitEthereumAttestationsTable() error {
	log.Info("init eth_attestations table in Redshift")
	_, err := c.redshiftDB.ExecContext(
		c.ctx,
		`
		CREATE TABLE IF NOT EXISTS eth_attestations(
			id INTEGER IDENTITY(1,1),
			msg_id TEXT NOT NULL,
			sender TEXT NOT NULL,
			subnet INT NOT NULL,
			slot BIGINT NOT NULL,
			arrival_time TIMESTAMP NOT NULL,
			time_in_slot REAL NOT NULL,
			val_pubkey TEXT,

			PRIMARY KEY(msg_id)
		);
		`)

	return err
}

// InsertNewEthereumAttestation inserts a new Ethereum attestation into the eth_attestations table
func (c *DBClient) InsertNewEthereumAttestation(attMsg *eth.TrackedAttestation) (query string, args []interface{}) {

	query = `
	INSERT INTO eth_attestations(
		msg_id,
		sender,
		subnet,
		slot,
		arrival_time,
		time_in_slot,
		val_pubkey)
	VALUES($1,$2,$3,$4,$5,$6,$7)
	ON CONFLICT (msg_id) DO NOTHING;
	`

	// args
	args = append(args, attMsg.MsgID)
	args = append(args, attMsg.Sender.String())
	args = append(args, attMsg.Subnet)
	args = append(args, attMsg.Slot)
	args = append(args, attMsg.ArrivalTime)
	args = append(args, float64(attMsg.TimeInSlot)/float64(time.Second))
	args = append(args, attMsg.ValPubkey)

	return query, args
}

// DropEthereumBeaconBlocksTable drops the eth_blocks table
func (c *DBClient) DropEthereumBeaconBlocksTable() error {
	log.Info("dropping the eth_blocks table")
	_, err := c.redshiftDB.ExecContext(
		c.ctx,
		`
		DROP TABLE IF EXISTS eth_blocks;
		`)
	return err
}

// InitEthereumBeaconBlocksTable initializes the eth_blocks table in Redshift
func (c *DBClient) InitEthereumBeaconBlocksTable() error {
	log.Info("init eth_blocks table in Redshift")
	_, err := c.redshiftDB.ExecContext(
		c.ctx,
		`
		CREATE TABLE IF NOT EXISTS eth_blocks(
			id INTEGER IDENTITY(1,1),
			msg_id TEXT NOT NULL,
			sender TEXT NOT NULL,
			slot BIGINT NOT NULL,
			arrival_time TIMESTAMP NOT NULL,
			time_in_slot REAL NOT NULL,
			val_idx BIGINT,

			PRIMARY KEY(msg_id)
		);
		`)

	return err
}

// InsertNewEthereumBeaconBlock inserts a new Ethereum beacon block into the eth_blocks table
func (c *DBClient) InsertNewEthereumBeaconBlock(bblock *eth.TrackedBeaconBlock) (query string, args []interface{}) {

	query = `
	INSERT INTO eth_blocks(
		msg_id,
		sender,
		slot,
		arrival_time,
		time_in_slot,
		val_idx)
	VALUES($1,$2,$3,$4,$5,$6)
	ON CONFLICT (msg_id) DO NOTHING;
	`

	// args
	args = append(args, bblock.MsgID)
	args = append(args, bblock.Sender.String())
	args = append(args, bblock.Slot)
	args = append(args, bblock.ArrivalTime)
	args = append(args, float64(bblock.TimeInSlot)/float64(time.Second))
	args = append(args, bblock.ValIndex)

	return query, args
}
