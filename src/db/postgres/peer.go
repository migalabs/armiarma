package postgres

import (
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/src/db/models"
	"github.com/migalabs/armiarma/src/utils"
	"github.com/pkg/errors"
)

// Postgres intregration variables
var (
	createPeerTable = `
	CREATE TABLE IF NOT EXISTS t_peers_summary(
		f_peer_id TEXT,
		f_pubkey TEXT,

		f_node_id TEXT,
		f_user_agent TEXT,
		f_client_name TEXT,
		f_client_os TEXT,
		f_client_version TEXT,
		f_enr TEXT,

		f_ip TEXT,
		f_country TEXT,
		f_country_code TEXT,
		f_city TEXT,
		f_latency FLOAT8,
		f_multi_addrs TEXT[],
		f_protocols TEXT[],
		f_protocol_version TEXT,

		f_connected_direction TEXT[],
		f_is_connected BOOL,
		f_attempted BOOL,
		f_succeed BOOL,
		f_attempts BIGINT,
		f_error TEXT[],
		f_last_error_timestamp TIMESTAMP,
		f_deprecated BOOL,
		f_last_identify_timestamp TIMESTAMP,

		f_negative_conn_attempts TIMESTAMP[],
		f_connection_times TIMESTAMP[],
		f_disconnection_times TIMESTAMP[],
		f_metadata_request BOOL,
		f_metadata_succeed BOOL,

		f_last_export BIGINT,

		PRIMARY KEY (f_peer_id)
	);
	`
	insertPeer = `
	INSERT INTO t_peers_summary(
		f_peer_id,
		f_pubkey,
		f_node_id,
		f_user_agent,
		f_client_name,
		f_client_os,
		f_client_version,
		f_enr,
		f_ip,
		f_country,
		f_country_code,
		f_city,
		f_latency,
		f_multi_addrs,
		f_protocols,
		f_protocol_version,
		f_connected_direction,
		f_is_connected,
		f_attempted,
		f_succeed,
		f_attempts,
		f_error,
		f_last_error_timestamp,
		f_deprecated,
		f_last_identify_timestamp,
		f_negative_conn_attempts,
		f_connection_times,
		f_disconnection_times,
		f_metadata_request,
		f_metadata_succeed,
		f_last_export)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31)
	ON CONFLICT (f_peer_id)
	DO UPDATE SET
		f_pubkey=EXCLUDED.f_pubkey,
		f_node_id=EXCLUDED.f_node_id,
		f_user_agent=EXCLUDED.f_user_agent,
		f_client_name=EXCLUDED.f_client_name,
		f_client_os=EXCLUDED.f_client_os,
		f_client_version=EXCLUDED.f_client_version,
		f_enr=EXCLUDED.f_enr,
		f_ip=EXCLUDED.f_ip,
		f_country=EXCLUDED.f_country,
		f_country_code=EXCLUDED.f_country_code,
		f_city=EXCLUDED.f_city,
		f_latency=EXCLUDED.f_latency,
		f_multi_addrs=EXCLUDED.f_multi_addrs,
		f_protocols=EXCLUDED.f_protocols,
		f_protocol_version=EXCLUDED.f_protocol_version,
		f_connected_direction=EXCLUDED.f_connected_direction,
		f_is_connected=EXCLUDED.f_is_connected,
		f_attempted=EXCLUDED.f_attempted,
		f_succeed=EXCLUDED.f_succeed,
		f_attempts=EXCLUDED.f_attempts,
		f_error=EXCLUDED.f_error,
		f_last_error_timestamp=EXCLUDED.f_last_error_timestamp,
		f_deprecated=EXCLUDED.f_deprecated,
		f_last_identify_timestamp=EXCLUDED.f_last_identify_timestamp,
		f_negative_conn_attempts=EXCLUDED.f_negative_conn_attempts,
		f_connection_times=EXCLUDED.f_connection_times,
		f_disconnection_times=EXCLUDED.f_disconnection_times,
		f_metadata_request=EXCLUDED.f_metadata_request,
		f_metadata_succeed=EXCLUDED.f_metadata_succeed,
		f_last_export=EXCLUDED.f_last_export
		`
)

// creates Peer table in the Postgres DB
func (p *PostgresDBService) createPeerTable() error {
	_, err := p.psqlConn.Exec(p.ctx, createPeerTable)
	if err != nil {
		return errors.Wrap(err, "error creating peers summary table")
	}
	return nil
}

func (p *PostgresDBService) StorePeer(peerID string, peer models.Peer) {
	log.Infof("storing peer %s", peerID)
	_, err := p.psqlConn.Exec(
		p.ctx,
		insertPeer,
		peerID,
		peer.Pubkey,
		peer.NodeId,
		peer.UserAgent,
		peer.ClientName,
		peer.ClientOS,
		peer.ClientVersion,
		peer.BlockchainNodeENR,
		peer.Ip,
		peer.Country,
		peer.CountryCode,
		peer.City,
		peer.Latency,
		peer.MAddrs,
		peer.Protocols,
		peer.ProtocolVersion,
		peer.ConnectedDirection,
		peer.IsConnected,
		peer.Attempted,
		peer.Succeed,
		peer.Attempts,
		peer.Error,
		peer.LastErrorTimestamp,
		peer.Deprecated,
		peer.LastIdentifyTimestamp,
		peer.NegativeConnAttempts,
		peer.ConnectionTimes,
		peer.DisconnectionTimes,
		peer.MetadataRequest,
		peer.MetadataSucceed,
		peer.LastExport,
	)
	if err != nil {
		// TODO: Add error return value? will need memmory and boltdb update
		log.Errorf("error inserting peer in the psqldb %s", err.Error())
	}
	log.Infof("storage success of peer %s", peerID)
}

func (p *PostgresDBService) LoadPeer(peerID string) (peer models.Peer, ok bool) {
	log.Infof("loading peer %s", peerID)
	row := p.psqlConn.QueryRow(
		p.ctx,
		"SELECT *FROM t_peers_summary WHERE f_peer_id=$1",
		peerID,
	)
	var multiAddrs []string
	err := row.Scan(
		&peer.PeerId,
		&peer.Pubkey,
		&peer.NodeId,
		&peer.UserAgent,
		&peer.ClientName,
		&peer.ClientOS,
		&peer.ClientVersion,
		&peer.BlockchainNodeENR,
		&peer.Ip,
		&peer.Country,
		&peer.CountryCode,
		&peer.City,
		&peer.Latency,
		&multiAddrs,
		&peer.Protocols,
		&peer.ProtocolVersion,
		&peer.ConnectedDirection,
		&peer.IsConnected,
		&peer.Attempted,
		&peer.Succeed,
		&peer.Attempts,
		&peer.Error,
		&peer.LastErrorTimestamp,
		&peer.Deprecated,
		&peer.LastIdentifyTimestamp,
		&peer.NegativeConnAttempts,
		&peer.ConnectionTimes,
		&peer.DisconnectionTimes,
		&peer.MetadataRequest,
		&peer.MetadataSucceed,
		&peer.LastExport,
	)

	if err != nil {
		log.Errorf("error loading info from peer %s. %s", peerID, err.Error())
		return peer, ok
	}
	// recompose the multiaddresses from []string
	for _, ma := range multiAddrs {
		maddres, err := utils.UnmarshalMaddr(ma)
		if err != nil {
			log.Warnf("unable to generate multiaddres from %s. %s", ma, err.Error())
			continue
		}
		peer.MAddrs = append(peer.MAddrs, maddres)
	}
	ok = true
	log.Infof("load success of peer %s", peerID)
	return peer, ok
}

func (p *PostgresDBService) DeletePeer(peer string) {
	log.Debug("deleting item")
}

func (p *PostgresDBService) GetPeers() []peer.ID {
	rows, err := p.psqlConn.Query(p.ctx,
		"SELECT f_peer_id FROM t_peers_summary")
	var peerList []peer.ID
	if err != nil {
		return peerList
	}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return peerList
		}
		for _, pID := range values {
			peerID_obj, err := peer.Decode(pID.(string))
			if err != nil {
				continue
			}
			peerList = append(peerList, peerID_obj)
		}
	}
	return peerList
}

func (p *PostgresDBService) GetPeerENR(peerID string) (enr *enode.Node, err error) {
	err = p.psqlConn.QueryRow(p.ctx,
		"SELECT f_enr FROM t_peers_summary WHERE f_peer_id=$1",
		peerID).Scan(&enr)
	return enr, err
}

func (p *PostgresDBService) ExportToCSV(filePath string) error {
	log.Info("Exporting to CSV")
	return nil
}
