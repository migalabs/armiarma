package peering

import (
	"math"
	"time"

	"github.com/migalabs/armiarma/pkg/hosts"
	log "github.com/sirupsen/logrus"
)

type Delay string

var (
	// define the types of delays in string
	PositiveDelay           Delay = "Positive"
	NegativeWithHopeDelay   Delay = "NegativeWithHope"
	NegativeWithNoHopeDelay Delay = "NegativeWithNoHope"
	ZeroDelay               Delay = "Zero"
	Minus1Delay             Delay = "Minus1"
	TimeoutDelay            Delay = "Timeout"

	MaxDelayTime time.Duration = time.Duration(math.Pow(2, 11) * float64(time.Minute))

	// define the initial delay we apply in each of the types
	initialDelayTime = map[Delay]time.Duration{
		PositiveDelay:           4 * time.Minute,
		NegativeWithHopeDelay:   2 * time.Minute,
		NegativeWithNoHopeDelay: 120 * time.Minute,
		ZeroDelay:               0 * time.Minute,
		Minus1Delay:             -1000 * time.Hour,
		TimeoutDelay:            60 * time.Minute, //experimental
	}
)

// ErrorToDelayType transforms an error into a DelayType.
func ErrorToDelayType(errString string) Delay {
	switch errString {
	case hosts.NoConnError:
		return PositiveDelay

	case hosts.DialErrorConnectionResetByPeer,
		hosts.DialErrorConnectionRefused,
		hosts.DialErrorContextDeadlineExceeded,
		hosts.DialErrorBackOff,
		hosts.ErrorRequestingMetadta,
		"unknown":
		return NegativeWithHopeDelay

	case hosts.DialErrorNoRouteToHost,
		hosts.DialErrorNetworkUnreachable,
		hosts.DialErrorPeerIDMismatch,
		hosts.DialErrorSelfAttempt,
		hosts.DialErrorNoGoodAddresses:
		return NegativeWithNoHopeDelay

	case hosts.DialErrorIoTimeout:
		return TimeoutDelay

	default:
		log.Tracef("Default Delay applied, error: %s\n", errString)
		return NegativeWithHopeDelay
	}
}

type DelayObject struct {
	delayDegree int   // number of times we have delayed
	dtype       Delay // type of delay we apply (positive, negativewithhope...)

}

func NewDelayObject(inputType Delay) DelayObject {
	return DelayObject{
		delayDegree: 0,
		dtype:       inputType,
	}
}

func (d *DelayObject) IncreaseDegree() {
	d.delayDegree++
}

func (d *DelayObject) SetDegree(newDegree int) {
	d.delayDegree = newDegree
}

func (d *DelayObject) CalculateDelay() time.Duration {
	var delay time.Duration
	switch d.dtype {
	case PositiveDelay, ZeroDelay, Minus1Delay:
		// return 2 hours * the degree (6,12,18...)
		//return time.Duration(d.DelayDegree) * InitialDelayTime[d.Type]
		delay = initialDelayTime[d.dtype]

	case NegativeWithHopeDelay, NegativeWithNoHopeDelay, TimeoutDelay:
		// if there are no attempts, there is no delay
		if d.delayDegree == 0 {
			delay = time.Duration(0)
		} else {
			// delay is (2 ** (delaydegree-1)) * 2 minutes (2,4,8,16,32...)
			delay = time.Duration(math.Pow(2, float64(d.delayDegree-1))) * initialDelayTime[d.dtype]
		}
	default:
		delay = initialDelayTime[Minus1Delay]
	}
	return delay
}
