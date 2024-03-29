package postgresql

import (
	"github.com/pkg/errors"

	eth "github.com/migalabs/armiarma/pkg/networks/ethereum"
	log "github.com/sirupsen/logrus"
)

func (d *DBClient) dropEthNodesTable() error {
	log.Debugf("droping eth_nodes table in psql-db")

	_, err := d.psqlPool.Exec(d.ctx, `
		DROP TABLE eth_nodes;
	`)
	return err

}

func (d *DBClient) InitEthNodesTable() error {
	log.Debugf("init eth_nodes table in psql-db")

	// try create the table in the DB
	_, err := d.psqlPool.Exec(
		d.ctx, `
		CREATE TABLE IF NOT EXISTS eth_nodes(
			id SERIAL,
			timestamp BIGINT NOT NULL,
			peer_id TEXT,
			node_id TEXT NOT NULL,
			seq BIGINT NOT NULL,
			ip TEXT NOT NULL,
			tcp INT,
			udp INT,
			pubkey TEXT NOT NULL,
			fork_digest TEXT,
			next_fork_version TEXT,
			attnets TEXT, 
			attnets_number INT,

			PRIMARY KEY(node_id),	
			UNIQUE(peer_id, pubkey)
		);
		`,
	)
	if err != nil {
		return errors.Wrap(err, "unable to create table eth_nodes in the db")
	}

	return nil
}

// Insert ENR in the DB
// insert into the db if new one, update the data if the ENR has a higher Seq number
func (d *DBClient) UpsertEnrInfo(enr *eth.EnrNode) (query string, args []interface{}) {
	log.Trace("upserting new enr to eth_nodes in psql-db")

	query = `
		INSERT INTO eth_nodes(
			timestamp,
			peer_id,
			node_id,
			seq,
			ip,
			tcp,
			udp,
			pubkey,
			fork_digest,
			next_fork_version,
			attnets,
			attnets_number)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)	
		ON CONFLICT (node_id)
		DO UPDATE SET
			timestamp = excluded.timestamp,
			seq = excluded.seq,
			ip = excluded.ip,
			tcp = excluded.tcp,
			udp = excluded.udp,
			fork_digest = excluded.fork_digest,
			next_fork_version = excluded.next_fork_version,
			attnets = excluded.attnets,
			attnets_number = excluded.attnets_number;
		`

	// if peer_id goes empty, not my fault here we should have checked it before
	var peerIDStr string
	peerId, err := enr.GetPeerID()
	if err == nil {
		peerIDStr = peerId.String()
	}

	args = append(args, enr.Timestamp.Unix())
	args = append(args, peerIDStr)
	args = append(args, enr.ID.String())
	args = append(args, enr.Seq)
	args = append(args, enr.IP)
	args = append(args, enr.TCP)
	args = append(args, enr.UDP)
	args = append(args, enr.GetPubkeyString())
	args = append(args, enr.Eth2Data.ForkDigest.String())
	args = append(args, enr.Eth2Data.NextForkVersion.String())
	args = append(args, enr.GetAttnetsString())
	args = append(args, enr.Attnets.NetNumber)

	return query, args
}
