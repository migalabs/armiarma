package postgres

import (
	"os"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/src/db/models"
	"github.com/migalabs/armiarma/src/utils"
	"github.com/pkg/errors"
	"github.com/protolambda/zrnt/eth2/beacon/common"
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

		f_metadata_timestamp TIMESTAMP,
		f_seq_number BIGINT,
		f_attnets TEXT,

		f_status_timestamp TIMESTAMP,
		f_fork_digest TEXT,
		f_finalized_root TEXT,
		f_finalized_epoch BIGINT,
		f_head_root TEXT,
		f_head_slot BIGINT,

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
		f_last_export,
		f_metadata_timestamp,
		f_seq_number,
		f_attnets,
		f_status_timestamp,
		f_fork_digest,
		f_finalized_root,
		f_finalized_epoch,
		f_head_root,
		f_head_slot)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, 
		$21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40)
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
		f_last_export=EXCLUDED.f_last_export,
		f_metadata_timestamp=EXCLUDED.f_metadata_timestamp,
		f_seq_number=EXCLUDED.f_seq_number,
		f_attnets=EXCLUDED.f_attnets,
		f_status_timestamp=EXCLUDED.f_status_timestamp,
		f_fork_digest=EXCLUDED.f_fork_digest,
		f_finalized_root=EXCLUDED.f_finalized_root,
		f_finalized_epoch=EXCLUDED.f_finalized_epoch,
		f_head_root=EXCLUDED.f_head_root,
		f_head_slot=EXCLUDED.f_head_slot
		`

	createPeerMessageMetrics = `
	CREATE TABLE IF NOT EXISTS t_peers_msg_metrics(
		f_peer_id TEXT,
		f_topic TEXT,
		
		f_count BIGINT,
		f_first_message TIMESTAMP,
		f_last_message TIMESTAMP,

		PRIMARY KEY (f_peer_id, f_topic)
	);
	`

	insertPeerMessageMetrics = `
	INSERT INTO t_peers_msg_metrics(
		f_peer_id,
		f_topic,
		f_count,
		f_first_message,
		f_last_message)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (f_peer_id, f_topic)
	DO UPDATE SET
		f_peer_id=EXCLUDED.f_peer_id,
		f_topic=EXCLUDED.f_topic,
		f_count=EXCLUDED.f_count,
		f_first_message=EXCLUDED.f_first_message,
		f_last_message=EXCLUDED.f_last_message
	`
)

// creates Peer table in the Postgres DB
func (p *PostgresDBService) createPeerTable() error {
	_, err := p.psqlPool.Exec(p.ctx, createPeerTable)
	if err != nil {
		return errors.Wrap(err, "error creating peers summary table")
	}
	return nil
}

func (p *PostgresDBService) createPeerMessageMetricsTable() error {
	_, err := p.psqlPool.Exec(p.ctx, createPeerMessageMetrics)
	if err != nil {
		return errors.Wrap(err, "error creating peer-message-metrics table")
	}
	return nil
}

func (p *PostgresDBService) StorePeer(peerID string, peer models.Peer) {
	_, err := p.psqlPool.Exec(
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
		peer.BeaconMetadata.Timestamp,
		peer.BeaconMetadata.Metadata.SeqNumber,
		peer.BeaconMetadata.Metadata.Attnets,
		peer.BeaconStatus.Timestamp,
		peer.BeaconStatus.Status.ForkDigest.String(),
		peer.BeaconStatus.Status.FinalizedRoot.String(),
		peer.BeaconStatus.Status.FinalizedEpoch,
		peer.BeaconStatus.Status.HeadRoot.String(),
		peer.BeaconStatus.Status.HeadSlot,
	)

	if err != nil {
		// TODO: Add error return value? will need memmory and boltdb update
		log.Errorf("error inserting peer in the psqldb %s", err.Error())
	}
	// Store the Peer-Messge-Metrics
	for topic, metrics := range peer.MessageMetrics {
		_, err = p.psqlPool.Exec(
			p.ctx,
			insertPeerMessageMetrics,
			peerID,
			topic,
			metrics.Count,
			metrics.FirstMessageTime,
			metrics.LastMessageTime,
		)
		if err != nil {
			// TODO: Add error return value? will need memmory and boltdb update
			log.Errorf("error inserting metrics of topic %s in the psqldb %s", topic, err.Error())
		}
	}

}

func (p *PostgresDBService) LoadPeer(peerID string) (models.Peer, bool) {
	log.Debugf("loading peer %s", peerID)
	row := p.psqlPool.QueryRow(
		p.ctx,
		"SELECT *FROM t_peers_summary WHERE f_peer_id=$1",
		peerID,
	)
	peer := models.NewPeer("")

	var multiAddrs []string
	// BeconMetadata
	var seqNumber common.SeqNr
	var attnet string
	// BeaconStatus
	var bStatusTS time.Time
	var forkDigest string
	var finalizedRoot string
	var finalizedEpoch int64
	var headRoot string
	var headSlot int64
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
		&peer.BeaconMetadata.Timestamp,
		&seqNumber,
		&attnet,
		&bStatusTS,
		&forkDigest,
		&finalizedRoot,
		&finalizedEpoch,
		&headRoot,
		&headSlot,
	)

	if err != nil {
		log.Debugf("error loading info from peer %s. %s", peerID, err.Error())
		return peer, false
	}
	// Compose BeaconMetadata
	if peer.BeaconMetadata.Timestamp.IsZero() {
		peer.BeaconMetadata.Metadata.SeqNumber = seqNumber
		err = peer.BeaconMetadata.Metadata.Attnets.UnmarshalText(utils.BytesFromString(attnet))
		if err != nil {
			log.Debugf("error loading info from peer %s. %s", peerID, err.Error())
			return peer, false
		}
	}
	// Compose Beacon Status
	if peer.BeaconStatus.Timestamp.IsZero() {
		bStatus, err := models.ParseBeaconStatusFromBasicTypes(
			bStatusTS,
			forkDigest,
			finalizedRoot,
			finalizedEpoch,
			headRoot,
			headSlot,
		)
		if err != nil {
			log.Debugf("unable to compose BeaconStatus from given parameters. %s", err.Error())
			return peer, false
		}
		peer.BeaconStatus = bStatus
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

	// Compose MessageMetrics from different table
	rows, err := p.psqlPool.Query(
		p.ctx,
		"SELECT *FROM t_peers_msg_metrics WHERE f_peer_id=$1",
		peerID,
	)
	if err != nil {
		log.Warnf("unable to load msg-metrics from peer %s. %s", peerID, err.Error())
		return peer, false
	}
	for rows.Next() {
		var peerID string
		var topic string
		var msgMetrics models.MessageMetric
		err = rows.Scan(
			&peerID,
			&topic,
			&msgMetrics.Count,
			&msgMetrics.FirstMessageTime,
			&msgMetrics.LastMessageTime,
		)
		peer.MessageMetrics[topic] = msgMetrics
		if err != nil {
			log.Warnf("peer %s - unable to parse msg-metrics from topic %s. %s", peerID, topic, err.Error())
			continue
		}
	}
	log.Debugf("load success of peer %s", peerID)
	return peer, true
}

func (p *PostgresDBService) DeletePeer(peerID string) {
	log.Debugf("deleting item")
	// Delete peer item from the table
	_, _ = p.psqlPool.Exec(
		p.ctx,
		"DELETE FROM t_peers_summary WHERE f_peer_id=$1",
		peerID,
	)
	// Delete any related msg-metrics
	_, _ = p.psqlPool.Exec(
		p.ctx,
		"DELETE FROM t_peers_msg_metrics WEHRE f_peer_id=$1",
		peerID,
	)
}

func (p *PostgresDBService) GetPeers() []peer.ID {
	rows, err := p.psqlPool.Query(p.ctx,
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
	err = p.psqlPool.QueryRow(p.ctx,
		"SELECT f_enr FROM t_peers_summary WHERE f_peer_id=$1",
		peerID).Scan(&enr)
	return enr, err
}

func (p *PostgresDBService) CheckPeersSummaryTableStatus() bool {
	// check the last activity of the tool
	lastActivity, err := p.GetLastActivityTime()
	if err != nil {
		log.Error(err, "unable to get last activity of the tool")
		return false
	}
	log.Info("Last Activity of the tool has been recorded to be", lastActivity)
	connectedPeers, err := p.GetConnectedPeers()
	if err != nil {
		log.Errorf("unable to get connected peers. %s", err.Error())
		return false
	}

	var cnt int = 0
	for _, peerID := range connectedPeers {
		peer, ok := p.LoadPeer(peerID)
		if !ok {
			log.Errorf("unable to load info from peer %s", peerID)
			return false
		}
		peer.DisconnectionEvent(lastActivity)
		p.StorePeer(peerID, peer)
		if !ok {
			log.Errorf("unable to store info from peer %s", peerID)
			return false
		}
		cnt++
	}
	nPeers, err := p.GetNumberOfPeers()
	if err != nil {
		log.Errorf("unable to get number of peers in db. %s", err.Error())
		return false
	}
	log.Infof("loaded psql db with %d peers on it (%d connected)", nPeers, cnt)
	return true
}

func (p *PostgresDBService) GetConnectedPeers() ([]string, error) {
	var connPeers []string
	rows, err := p.psqlPool.Query(p.ctx,
		"SELECT f_peer_id FROM t_peers_summary WHERE f_is_connected='true'",
	)
	if err != nil {
		return connPeers, errors.Wrap(err, "unable to query peers with IsConnected flag = true")
	}
	// Iterate through the peers whos flag IsConnecte=True
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return connPeers, errors.Wrap(err, "unable to parse PeerID of peer with IsConnected=true")
		}
		// aggregate the disconnection for the peer that was connected
		for _, peerID := range values {
			connPeers = append(connPeers, peerID.(string))
		}
	}

	return connPeers, nil
}

func (p *PostgresDBService) GetNumberOfPeers() (int, error) {
	rows, err := p.psqlPool.Query(p.ctx,
		"SELECT f_peer_id FROM t_peers_summary")
	if err != nil {
		return -1, errors.Wrap(err, "unable to query the total amount of peers in the table")
	}
	var peerCounter int = 0
	for rows.Next() {
		peerCounter++
	}
	return peerCounter, nil
}

func (p *PostgresDBService) GetLastActivityTime() (time.Time, error) {
	var lastActivity time.Time
	query := `
		WITH foo as (
			SELECT UNNEST(f_negative_conn_attempts) AS a_last_activity
			FROM t_peers_summary
		)
		SELECT MAX(a_last_activity) as a_last_activity
		FROM foo;
	`

	rows, err := p.psqlPool.Query(p.ctx, query)
	if err != nil {
		return lastActivity, errors.Wrap(err, "unable to get Last Activity of the tool from postgresql db")
	}
	for rows.Next() {
		var t time.Time
		err := rows.Scan(&t)
		if err != nil {
			// check if DB is empty or if there are actual values
			peers := p.GetPeers()
			if len(peers) == 0 {
				log.Info("Empty loaded DB")
				return time.Time{}, nil
			} else {
				return time.Time{}, errors.Wrap(err, "unable to parse Last Activity of the tool from postgresql db")
			}
		}
		if t.After(lastActivity) {
			lastActivity = t
		}
	}
	return lastActivity, nil
}

func (p *PostgresDBService) ExportToCSV(filePath string) error {
	log.Info("Exporting metrics to csv: ", filePath)
	csvFile, err := os.Create(filePath)
	if err != nil {
		return errors.Wrap(err, "error opening the file "+filePath)
	}
	defer csvFile.Close()

	// First raw of the file will be the Titles of the columns
	_, err = csvFile.WriteString("Peer Id,Node Id,Fork Digest,User Agent,Client,Version,Pubkey,Address,Ip,Country,City,ENR,Request Metadata,Success Metadata,Attempted,Succeed,Deprecated,ConnStablished,IsConnected,Attempts,Error,Last Error Timestamp,Last Identify Timestamp,Latency,Connections,Disconnections,Last Connection,Conn Direction,Connected Time,Beacon Blocks,Beacon Aggregations,Voluntary Exits,Proposer Slashings,Attester Slashings,Total Messages\n")
	if err != nil {
		errors.Wrap(err, "error while writing the titles on the csv "+filePath)
	}

	// get the list of all the peers
	peers := p.GetPeers()

	for _, nextPeer := range peers {
		nPeer, ok := p.LoadPeer(nextPeer.String())
		if !ok {
			log.Errorf("error reading peer %s for metrics export. %s", nextPeer.String(), err.Error())
			continue
		}
		_, err := csvFile.WriteString(nPeer.ToCsvLine())
		if err != nil {
			log.Errorf("error reading peer %s for metrics export. %s", nextPeer.String(), err.Error())
			continue
		}
	}
	return nil
}
