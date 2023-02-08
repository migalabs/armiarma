package ethereum

import (
	"regexp"
	"strconv"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
)

var (
	ErrorNoSubnet          = errors.New("no subnet found in topic")
	ErrorNotParsableSubnet = errors.New("not parseable subnet int")
)

type TrackedAttestation struct {
	MsgID  string
	Sender peer.ID
	Subnet int

	ArrivalTime time.Time     // time of arrival
	TimeInSlot  time.Duration // exact time inside the slot (range between 0secs and 12s*32slots)

	ValPubkey string
	Slot      int64
}

func (a *TrackedAttestation) IsZero() bool {
	return a.Slot == 0
}

type TrackedBeaconBlock struct {
	MsgID  string
	Sender peer.ID

	ArrivalTime time.Time     // time of arrival
	TimeInSlot  time.Duration // exact time inside the slot (range between 0secs and 12s*32slots)

	ValIndex int64
	Slot     int64
}

func (a *TrackedBeaconBlock) IsZero() bool {
	return a.Slot == 0
}

func GetSubnetFromTopic(topic string) (int, error) {
	re := regexp.MustCompile(`attestation_([0-9]+)`)
	match := re.FindAllString(topic, -1)
	if len(match) < 1 {
		return -1, ErrorNoSubnet
	}

	re2 := regexp.MustCompile("([0-9]+)")
	match = re2.FindAllString(match[0], -1)
	if len(match) < 1 {
		return -1, ErrorNotParsableSubnet
	}
	subnet, err := strconv.Atoi(match[0])
	if err != nil {
		return -1, errors.Wrap(err, "unable to conver subnet to int")
	}
	return subnet, nil
}
