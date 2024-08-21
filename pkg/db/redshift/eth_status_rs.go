package redshift

import (
	//"context"
	//"database/sql"

	eth "github.com/migalabs/armiarma/pkg/networks/ethereum"
	log "github.com/sirupsen/logrus"
	_ "github.com/lib/pq"
)


// DropEthereumNodeStatus drops the eth_status table from Redshift
func (d *DBClient) DropEthereumNodeStatus() error {
	log.Debug("dropping eth_status table from Redshift")
	_, err := d.redshiftDB.ExecContext(
		d.ctx, `
			DROP TABLE IF EXISTS eth_status;
	`)
	return err
}

// InitEthereumNodeStatus initializes the eth_status table in Redshift
func (d *DBClient) InitEthereumNodeStatus() error {
	log.Debug("init eth_status table in Redshift")
	_, err := d.redshiftDB.ExecContext(
		d.ctx, `
		CREATE TABLE IF NOT EXISTS eth_status(
			id INTEGER IDENTITY(1,1),
			peer_id TEXT,
			timestamp BIGINT,
			fork_digest TEXT,
			finalized_root TEXT,
			finalized_epoch BIGINT,
			head_root TEXT,
			head_slot BIGINT,
			seq_number BIGINT,
			attnets TEXT,
			syncnets TEXT,

			PRIMARY KEY (peer_id)
		);
	`)
	return err
}

// UpsertEthereumNodeStatus inserts or updates beacon status in the eth_status table
func (d *DBClient) UpsertEthereumNodeStatus(bstatus eth.BeaconStatusStamped) (query string, args []interface{}) {
	log.Trace("upserting beacon status to eth_status in Redshift")
	query = `
		INSERT INTO eth_status(
			peer_id,
			timestamp,
			fork_digest,
			finalized_root,
			finalized_epoch,
			head_root,
			head_slot)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (peer_id)
		DO UPDATE SET
			timestamp = excluded.timestamp,
			fork_digest = excluded.fork_digest,
			finalized_root = excluded.finalized_root,
			finalized_epoch = excluded.finalized_epoch,
			head_root = excluded.head_root,
			head_slot = excluded.head_slot;	
	`

	args = append(args, bstatus.PeerID.String())
	args = append(args, bstatus.Timestamp.Unix())
	args = append(args, bstatus.Status.ForkDigest.String())
	args = append(args, bstatus.Status.FinalizedRoot.String())
	args = append(args, bstatus.Status.FinalizedEpoch)
	args = append(args, bstatus.Status.HeadRoot.String())
	args = append(args, bstatus.Status.HeadSlot)

	return query, args
}

// UpsertEthereumNodeMetadata inserts or updates beacon metadata in the eth_status table
func (d *DBClient) UpsertEthereumNodeMetadata(bmetadata eth.BeaconMetadataStamped) (query string, args []interface{}) {
	log.Trace("upserting beacon metadata to eth_status in Redshift")
	query = `
		INSERT INTO eth_status(
			peer_id,
			timestamp,
			seq_number,
			attnets,
			syncnets)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (peer_id)
		DO UPDATE SET
			timestamp = excluded.timestamp,
			seq_number = excluded.seq_number,
			attnets = excluded.attnets,
			syncnets = excluded.syncnets;
		`

	args = append(args, bmetadata.PeerID.String())
	args = append(args, bmetadata.Timestamp.Unix())
	args = append(args, bmetadata.Metadata.SeqNumber)
	args = append(args, bmetadata.Metadata.Attnets.String())
	args = append(args, bmetadata.Metadata.Syncnets.String())

	return query, args
}
