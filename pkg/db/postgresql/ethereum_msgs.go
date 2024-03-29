package postgresql

import (
	"time"

	eth "github.com/migalabs/armiarma/pkg/networks/ethereum"
	log "github.com/sirupsen/logrus"
)

func (c *DBClient) dropEtherumAttestationsTable() error {
	log.Info("droping the eth_attestations table")
	_, err := c.psqlPool.Exec(
		c.ctx,
		`
		DROP TABLE eth_attestations;
		`)
	return err
}

func (c *DBClient) initEthereumAttestationsTable() error {
	log.Info("init eth_attestations table in psql-db")
	_, err := c.psqlPool.Exec(
		c.ctx,
		`
		CREATE TABLE IF NOT EXISTS eth_attestations(
			id SERIAL,
			msg_id TEXT NOT NULL,
			sender TEXT NOT NULL,
			subnet INT NOT NULL,
			slot BIGINT NOT NULL,
			arrival_time TIME NOT NULL,
			time_in_slot REAL NOT NULL,
			val_pubkey TEXT,

			PRIMARY KEY(msg_id)
		)
		`)

	return err
}

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
	ON CONFLICT (msg_id) DO NOTHING
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

// Beacon Blocks
func (c *DBClient) dropEtherumBeaconBlocksTable() error {
	log.Info("droping the eth_blocks table")
	_, err := c.psqlPool.Exec(
		c.ctx,
		`
		DROP TABLE eth_blocks;
		`)
	return err
}

func (c *DBClient) initEthereumBeaconBlocksTable() error {
	log.Info("init eth_blocks table in psql-db")
	_, err := c.psqlPool.Exec(
		c.ctx,
		`
		CREATE TABLE IF NOT EXISTS eth_blocks(
			id SERIAL,
			msg_id TEXT NOT NULL,
			sender TEXT NOT NULL,
			slot BIGINT NOT NULL,
			arrival_time TIME NOT NULL,
			time_in_slot REAL NOT NULL,
			val_idx BIGINT,

			PRIMARY KEY(msg_id)
		)
		`)

	return err
}

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
	ON CONFLICT (msg_id) DO NOTHING
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
