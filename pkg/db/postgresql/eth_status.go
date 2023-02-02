package postgresql

import (
	log "github.com/sirupsen/logrus"

	eth "github.com/migalabs/armiarma/pkg/networks/ethereum"
)

func (d *DBClient) DropEthereumNodeStatus() error {
	log.Debug("dropping eth_status table from psql-db")
	_, err := d.psqlPool.Exec(
		d.ctx, `
			DROP TABEL eth_status;
		);
	`)
	return err
}

func (d *DBClient) InitEthereumNodeStatus() error {
	log.Debug("init eth_status table in psql-db")
	_, err := d.psqlPool.Exec(
		d.ctx, `
		CREATE TABLE IF NOT EXISTS eth_status(
			id SERIAL,
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

func (d *DBClient) UpsertEthereumNodeStatus(bstatus eth.BeaconStatusStamped) (query string, args []interface{}) {
	log.Trace("upserting beacon status to eth_status in psql-db")
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
			peer_id = excluded.peer_id,
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

func (d *DBClient) UpsertEthereumNodeMetadata(bmetadata eth.BeaconMetadataStamped) (query string, args []interface{}) {
	log.Trace("upserting beacon metadata to eth_status in psql-db")
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
			peer_id = excluded.peer_id,
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
