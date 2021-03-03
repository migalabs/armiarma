package gossipimport

import (
    "github.com/protolambda/rumor/chain"
    "github.com/protolambda/zrnt/eth2/beacon"
    "github.com/protolambda/ztyp/tree"
)

// Chain utils 

// Create a new HotEntry for the Chain 
func GenerateNewHotEntry(stateView *beacon.BeaconStateView, spec *beacon.Spec) (*chain.HotEntry, error) {
    var hotEntry *chain.HotEntry
    // once we get the latest BeaconStateView get the LatestBlockHeader
    slot, err := stateView.Slot()
    if err != nil {
        return hotEntry, err
    }
    // get the latestBlockHeader
    latestHeader, err := stateView.LatestBlockHeader()
	if err != nil {
		return hotEntry, err
	}
	latestHeader, err = beacon.AsBeaconBlockHeader(latestHeader.Copy())
	if err != nil {
		return hotEntry, err
	}
	headerStateRoot, err := latestHeader.StateRoot()
	if err != nil {
		return hotEntry, err
	}
	if headerStateRoot == (beacon.Root{}) {
		if err := latestHeader.SetStateRoot(stateView.HashTreeRoot(tree.GetHashFn())); err != nil {
			return hotEntry, err
		}
	}
	blockRoot := latestHeader.HashTreeRoot(tree.GetHashFn())
	parentRoot, err := latestHeader.ParentRoot()
	if err != nil {
		return hotEntry, err
	}
	epc, err := spec.NewEpochsContext(stateView)
	if err != nil {
		return hotEntry, err
	}
	hotEntry= chain.NewHotEntry(slot, blockRoot, parentRoot, stateView, epc)
    return hotEntry, err
}

// Function that will generate a beacon.Status{} from the given BeaconBlock and BeaconStatus
func NodeStatusFromBeaconState(bstate *beacon.BeaconState, forkDigest beacon.ForkDigest) (*beacon.Status, error) {
	blockRoot := bstate.LatestBlockHeader.HashTreeRoot(tree.GetHashFn())

    st := &beacon.Status{
        ForkDigest: forkDigest,
        FinalizedRoot: bstate.CurrentJustifiedCheckpoint.Root,
        FinalizedEpoch: bstate.CurrentJustifiedCheckpoint.Epoch,
        HeadRoot: blockRoot,
        HeadSlot: bstate.Slot,
    }
    return st, nil
}


