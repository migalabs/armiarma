package postgresql

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/src/db/models"
	"github.com/migalabs/armiarma/src/utils"
	"github.com/pkg/errors"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

// Postgres intregration variables
var (
	createEth2PeerTable = `
	CREATE TABLE IF NOT EXISTS t_eth2_summary(
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
	insertEth2Peer = `
	INSERT INTO t_eth2_summary(
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

	createEth2PeerMessageMetrics = `
	CREATE TABLE IF NOT EXISTS t_eth2_msg_metrics(
		f_peer_id TEXT,
		f_topic TEXT,
		
		f_count BIGINT,
		f_first_message TIMESTAMP,
		f_last_message TIMESTAMP,

		PRIMARY KEY (f_peer_id, f_topic)
	);
	`

	insertEth2PeerMessageMetrics = `
	INSERT INTO t_eth2_msg_metrics(
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

//
type Eth2Model struct {
	Network string
}

//
func NewEth2Model(network string) Eth2Model {
	return Eth2Model{
		Network: network,
	}
}

func (p *Eth2Model) init(ctx context.Context, pool *pgxpool.Pool) error {
	// create the tables
	err := p.createPeerTable(ctx, pool)
	if err != nil {
		return err
	}
	ok := p.CheckPeersSummaryTableStatus(ctx, pool)
	if !ok {
		return errors.New("unable to check existing connected peers in the postgres db")
	}

	// ---- Message Metrics Table ----
	err = p.createPeerMessageMetricsTable(ctx, pool)
	if err != nil {
		return err
	}

	// ---- Client Diversity Table ----
	err = p.createEth2ClientDiversityTable(ctx, pool)
	if err != nil {
		return err
	}
	return nil
}

// creates Peer table in the Postgres DB
func (p *Eth2Model) createPeerTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, createEth2PeerTable)
	if err != nil {
		return errors.Wrap(err, "error creating peers summary table")
	}
	return nil
}

func (p *Eth2Model) createPeerMessageMetricsTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, createEth2PeerMessageMetrics)
	if err != nil {
		return errors.Wrap(err, "error creating peer-message-metrics table")
	}
	return nil
}

func (p *Eth2Model) StorePeer(ctx context.Context, pool *pgxpool.Pool, peerID string, peer models.Peer) {
	// load Attributes
	var nodeid string
	nid, ok := peer.GetAtt("nodeid")
	if ok {
		nodeid = nid.(string)
	}
	var pubkey string
	pk, ok := peer.GetAtt("pubkey")
	if ok {
		pubkey = pk.(string)
	}
	var enr string
	e, ok := peer.GetAtt("enr")
	if ok {
		enr = e.(string)
	}
	// check
	var beaconmetadata models.BeaconMetadataStamped
	bm, ok := peer.GetAtt("beaconmetadata")
	if ok {
		beaconmetadata = bm.(models.BeaconMetadataStamped)
	}

	var beaconstatus models.BeaconStatusStamped
	bs, ok := peer.GetAtt("beaconstatus")
	if ok {
		beaconstatus = bs.(models.BeaconStatusStamped)
	}

	// store the peer
	_, err := pool.Exec(
		ctx,
		insertEth2Peer,
		peerID,
		pubkey,
		nodeid,
		peer.UserAgent,
		peer.ClientName,
		peer.ClientOS,
		peer.ClientVersion,
		enr,
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
		beaconmetadata.Timestamp,
		beaconmetadata.Metadata.SeqNumber,
		beaconmetadata.Metadata.Attnets,
		beaconstatus.Timestamp,
		beaconstatus.Status.ForkDigest.String(),
		beaconstatus.Status.FinalizedRoot.String(),
		beaconstatus.Status.FinalizedEpoch,
		beaconstatus.Status.HeadRoot.String(),
		beaconstatus.Status.HeadSlot,
	)

	if err != nil {
		// TODO: Add error return value? will need memmory and boltdb update
		log.Errorf("error inserting peer in the psqldb %s", err.Error())
	}
	// Store the Peer-Messge-Metrics
	for topic, metrics := range peer.MessageMetrics {
		_, err = pool.Exec(
			ctx,
			insertEth2PeerMessageMetrics,
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

func (p *Eth2Model) LoadPeer(ctx context.Context, pool *pgxpool.Pool, peerID string) (models.Peer, bool) {
	log.Debugf("loading peer %s", peerID)
	row := pool.QueryRow(
		ctx,
		"SELECT *FROM t_eth2_summary WHERE f_peer_id=$1",
		peerID,
	)
	peer := models.NewPeer("")

	var multiAddrs []string
	// App layer values
	var nodeid string
	var pubkey string
	var enr string

	// BeconMetadata
	var beaconmetadata models.BeaconMetadataStamped
	var seqNumber common.SeqNr
	var attnet string

	// BeaconStatus
	var beaconstatus models.BeaconStatusStamped
	var forkDigest string
	var finalizedRoot string
	var finalizedEpoch int64
	var headRoot string
	var headSlot int64

	err := row.Scan(
		&peer.PeerId,
		&pubkey,
		&nodeid,
		&peer.UserAgent,
		&peer.ClientName,
		&peer.ClientOS,
		&peer.ClientVersion,
		&enr,
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
		&beaconmetadata.Timestamp,
		&seqNumber,
		&attnet,
		&beaconstatus.Timestamp,
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
	//  Set the app attributes
	peer.SetAtt("pubkey", pubkey)
	peer.SetAtt("nodeid", nodeid)
	peer.SetAtt("enr", enr)

	// Compose BeaconMetadata
	if !beaconmetadata.Timestamp.IsZero() {
		beaconmetadata.Metadata.SeqNumber = seqNumber
		err = beaconmetadata.Metadata.Attnets.UnmarshalText(utils.BytesFromString(attnet))
		if err != nil {
			log.Debugf("unable to compose Metadata from given parameters. %s", err.Error())
			return peer, false
		}
	}
	peer.SetAtt("beaconmetadata", beaconmetadata)

	// Compose Beacon Status
	if !beaconstatus.Timestamp.IsZero() {
		bStatus, err := models.ParseBeaconStatusFromBasicTypes(
			beaconstatus.Timestamp,
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
		beaconstatus = bStatus
	}
	peer.SetAtt("beaconstatus", beaconstatus)

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
	rows, err := pool.Query(
		ctx,
		"SELECT *FROM t_eth2_msg_metrics WHERE f_peer_id=$1",
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

func (p *Eth2Model) DeletePeer(ctx context.Context, pool *pgxpool.Pool, peerID string) {
	log.Debugf("deleting item")
	// Delete peer item from the table
	_, _ = pool.Exec(
		ctx,
		"DELETE FROM t_eth2_summary WHERE f_peer_id=$1",
		peerID,
	)
	// Delete any related msg-metrics
	_, _ = pool.Exec(
		ctx,
		"DELETE FROM t_eth2_msg_metrics WEHRE f_peer_id=$1",
		peerID,
	)
}

func (p *Eth2Model) GetPeers(ctx context.Context, pool *pgxpool.Pool) []peer.ID {
	rows, err := pool.Query(ctx,
		"SELECT f_peer_id FROM t_eth2_summary")
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

func (p *Eth2Model) CheckPeersSummaryTableStatus(ctx context.Context, pool *pgxpool.Pool) bool {
	// check the last activity of the tool
	lastActivity, err := p.GetLastActivityTime(ctx, pool)
	if err != nil {
		log.Error(err, "unable to get last activity of the tool")
		return false
	}
	log.Info("Last Activity of the tool has been recorded to be", lastActivity)
	connectedPeers, err := p.GetConnectedPeers(ctx, pool)
	if err != nil {
		log.Errorf("unable to get connected peers. %s", err.Error())
		return false
	}

	var cnt int = 0
	for _, peerID := range connectedPeers {
		peer, ok := p.LoadPeer(ctx, pool, peerID)
		if !ok {
			log.Errorf("unable to load info from peer %s", peerID)
			return false
		}
		peer.DisconnectionEvent(lastActivity)
		p.StorePeer(ctx, pool, peerID, peer)
		if !ok {
			log.Errorf("unable to store info from peer %s", peerID)
			return false
		}
		cnt++
	}
	nPeers, err := p.GetNumberOfPeers(ctx, pool)
	if err != nil {
		log.Errorf("unable to get number of peers in db. %s", err.Error())
		return false
	}
	log.Infof("loaded psql db with %d peers on it (%d connected)", nPeers, cnt)
	return true
}

func (p *Eth2Model) GetConnectedPeers(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	var connPeers []string
	rows, err := pool.Query(ctx,
		"SELECT f_peer_id FROM t_eth2_summary WHERE f_is_connected='true'",
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

func (p *Eth2Model) GetNumberOfPeers(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	rows, err := pool.Query(ctx,
		"SELECT f_peer_id FROM t_eth2_summary")
	if err != nil {
		return -1, errors.Wrap(err, "unable to query the total amount of peers in the table")
	}
	var peerCounter int = 0
	for rows.Next() {
		peerCounter++
	}
	return peerCounter, nil
}

func (p *Eth2Model) GetLastActivityTime(ctx context.Context, pool *pgxpool.Pool) (time.Time, error) {
	var lastActivity time.Time
	query := `
		WITH foo as (
			SELECT UNNEST(f_negative_conn_attempts) AS a_last_activity
			FROM t_eth2_summary
		)
		SELECT MAX(a_last_activity) as a_last_activity
		FROM foo;
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return lastActivity, errors.Wrap(err, "unable to get Last Activity of the tool from postgresql db")
	}
	for rows.Next() {
		var t time.Time
		err := rows.Scan(&t)
		if err != nil {
			// check if DB is empty or if there are actual values
			peers := p.GetPeers(ctx, pool)
			if len(peers) == 0 {
				log.Info("Empty loaded DB")
				return time.Time{}, nil
			} else {
				continue
				//return time.Time{}, errors.Wrap(err, "unable to parse Last Activity of the tool from postgresql db")
			}
		}
		if t.After(lastActivity) {
			lastActivity = t
		}
	}
	return lastActivity, nil
}
