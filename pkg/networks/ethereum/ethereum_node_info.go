package ethereum

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/pkg/errors"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

// Basic BeaconMetadata struct that includes the timestamp of the received beacon metadata
type BeaconMetadataStamped struct {
	Timestamp time.Time
	PeerID    peer.ID
	Metadata  common.MetaData
}

// NewbeaconMetadata generates a timestamped beacon.Metadata structure
func NewBeaconMetadata(peerId peer.ID, bMetadata common.MetaData) BeaconMetadataStamped {
	return BeaconMetadataStamped{
		Timestamp: time.Now(),
		PeerID:    peerId,
		Metadata:  bMetadata,
	}
}

// IsEmpty checks whether the Metadata is empty or not
func (b *BeaconMetadataStamped) IsEmpty() bool {
	return b.Timestamp.IsZero()
}

// Basic BeaconMetadata struct that includes The timestamp of the received beacon Status
type BeaconStatusStamped struct {
	Timestamp time.Time
	PeerID    peer.ID
	Status    common.Status
}

// IsEmpty checks whether the
func (b *BeaconStatusStamped) IsEmpty() bool {
	return b.Timestamp.IsZero()
}

// NewBeaconStatus generates a timestamped OBJ that has all the content of the
func NewBeaconStatus(peerId peer.ID, bStatus common.Status) BeaconStatusStamped {
	return BeaconStatusStamped{
		Timestamp: time.Now(),
		PeerID:    peerId,
		Status:    bStatus,
	}
}

// --- Parsers ----

// ParseBeaconStatusFromInterfaced returns the Timestamped beaconStatus structure from a input interface
// JSON oriented -> Seems to be deprecated
func ParseBeaconStatusFromInterface(input interface{}) (BeaconStatusStamped, error) {
	var result BeaconStatusStamped
	var err error

	inputMap := input.(map[string]interface{})

	// timestamp
	result.Timestamp, err = time.Parse(time.RFC3339, inputMap["Timestamp"].(string))
	if err != nil {
		return result, errors.Wrap(err, "unable to compose BeaconStatus.Timestamp from readed interface")
	}

	// peerID
	peerId, err := peer.Decode(inputMap["PeerID"].(string))
	if err != nil {
		return result, errors.Wrap(err, "unable to decode peerID from readed interface")
	}
	result.PeerID = peerId

	// BeaconStatus
	status := inputMap["Status"].(map[string]interface{})
	// if the forkdigest field is empty, return empty BeaconStatus
	fd, _ := status["ForkDigest"].(string)
	if len(fd) == 0 {
		return result, nil
	}
	// otherwise, compose the readed beaconStatus
	err = result.Status.ForkDigest.UnmarshalText([]byte(fd))
	if err != nil {
		return result, errors.Wrap(err, "unable to compose BeaconStatus.ForkDigest from readed interface")
	}
	fr, _ := status["FinalizedRoot"].(string)
	var frByte [32]byte
	copy(frByte[:], fr[:32])
	result.Status.FinalizedRoot = common.Root(frByte)
	e, err := strconv.ParseUint(status["Epoch"].(string), 0, 64)
	if err != nil {
		return result, errors.Wrap(err, "unable to compose BeaconStatus.Epoch from readed interface")
	}
	result.Status.FinalizedEpoch = common.Epoch(uint64(e))
	hr, _ := status["HeadRoot"].(string)
	var hrBytes [32]byte
	copy(hrBytes[:], hr[:32])
	result.Status.HeadRoot = common.Root(hrBytes)
	s, err := strconv.ParseUint(status["HeadSlot"].(string), 0, 64)
	if err != nil {
		return result, errors.Wrap(err, "unable to compose BeaconStatus.HeadSlot from readed interface")
	}
	result.Status.HeadSlot = common.Slot(uint64(s))
	return result, nil
}

func ParseBeaconStatusFromBasicTypes(
	t time.Time,
	peerIdStr string,
	forkdigest string,
	finaRoot string,
	finaEpoch int64,
	headRoot string,
	headSlot int64) (BeaconStatusStamped, error) {

	var result BeaconStatusStamped
	var err error

	// timestamp
	result.Timestamp = t

	// peerID
	peerId, err := peer.Decode(peerIdStr)
	if err != nil {
		return result, errors.Wrap(err, "unable to decode peerID from readed interface")
	}
	result.PeerID = peerId

	err = result.Status.ForkDigest.UnmarshalText(utils.BytesFromString(forkdigest))
	if err != nil {
		return result, fmt.Errorf("unable to compose BeaconStatus.ForkDigest from readed interface")
	}

	// FINALIZED ROOT
	// remove 0x if exists from the root string
	if strings.Contains(finaRoot, "0x") {
		finaRoot = strings.Replace(finaRoot, "0x", "", 1)
	}
	// conver strig to hex bytes
	fr, err := hex.DecodeString(finaRoot)
	if err != nil {
		return result, fmt.Errorf("unable to decode finalizedRoot %s", err.Error())
	}
	// copy the bytes of the root into a [32]byte varible (otherwis, commmon.Root complains)
	var frBytes [32]byte
	copy(frBytes[:], fr[:32])
	result.Status.FinalizedRoot = common.Root(frBytes)
	result.Status.FinalizedEpoch = common.Epoch(uint64(finaEpoch))
	// HEAD ROOT
	// remove 0x if exists from the root string
	if strings.Contains(headRoot, "0x") {
		headRoot = strings.Replace(headRoot, "0x", "", 1)
	}
	// conver strig to hex bytes
	hr, err := hex.DecodeString(headRoot)
	if err != nil {
		return result, fmt.Errorf("unable to decode finalizedRoot %s", err.Error())
	}
	var hrBytes [32]byte
	copy(hrBytes[:], hr[:32])
	// copy the bytes of the root into a [32]byte varible (otherwis, commmon.Root complains)
	result.Status.HeadRoot = common.Root(hrBytes)

	result.Status.HeadSlot = common.Slot(uint64(headSlot))
	return result, nil
}
