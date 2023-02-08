package ethereum

import (
	"time"
)

// translates the arrival time into time since slot started
func GetTimeInSlot(genesis time.Time, arrivalTime time.Time, slot int64) time.Duration {
	// get slot time since genesis
	slotTime := genesis.Add((time.Duration(slot) * SecondsPerSlot))

	// compare the arrival time to the base-slot time
	inSlotTime := arrivalTime.Sub(slotTime)
	return inSlotTime
}
