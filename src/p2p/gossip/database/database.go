package database

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	//	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/protolambda/zrnt/eth2/beacon"
)

// Circular databases for the Received messages
type MessageDatabase struct {
	MessageBuffer *MessageBuffer
	// This seen message list might be a bit too much to keep it in memory at one point,
	// TODO: -Find a cleaner way to have it
	MessageSeenList   map[string]bool                    // messageID - true
	ValidatorMessages map[beacon.ValidatorIndex][]string // ValidatorIndex - messageID
	SlotMessages      map[beacon.Slot][]string           // slot - messageID

	Spec *beacon.Spec
	// Currently only the Beacon Blocks can be supported
	BlockNotChan chan *beacon.SignedBeaconBlock
	// AttestNotChan chan *beacon.Attestation | beacon.PendingAttestation
	DiskDBPath string
	sync.RWMutex
}

func NewMessageDatabase(spec *beacon.Spec, msgLimit int, diskDBPath string) *MessageDatabase {
	mdb := &MessageDatabase{
		MessageBuffer:     NewMessageBuffer(msgLimit, diskDBPath),
		MessageSeenList:   make(map[string]bool, 0),
		ValidatorMessages: make(map[beacon.ValidatorIndex][]string, 0),
		SlotMessages:      make(map[beacon.Slot][]string, 0),
		Spec:              spec,
		BlockNotChan:      make(chan *beacon.SignedBeaconBlock, 5), // Buffer of 5 messages just in case
		DiskDBPath:        diskDBPath,
	}
	// add the messages that where on the Disk Database to the seen msgs
	files, err := ioutil.ReadDir(diskDBPath)
	if err != nil {
		fmt.Println("DEBUG (MessageIDs) error reading the content of the Disk Database content, path:", diskDBPath)
	}
	mdb.RWMutex.Lock()
	for _, f := range files {
		msgID := strings.Trim(f.Name(), ".json")
		mdb.MessageSeenList[msgID] = true
	}
	mdb.RWMutex.Unlock()
	return mdb
}

// TODO: This spec will be hard-coded to the Mainnet Specifications (All the topics will use the same Spec)
// Configure the Specifications for the Received Gossip Messages (Needed to Serialize and Deserialize the Received messages)
func (c *MessageDatabase) SetSpec(spec *beacon.Spec) error {
	if spec != nil {
		c.RWMutex.Lock()
		c.Spec = spec
		c.RWMutex.Unlock()
		return nil
	} else {
		return fmt.Errorf("Specifications were unable to set, Empty pointer")
	}
}

// Returns if the message was already on the DB (to see if we need to notify of a new message)
func (msgdb *MessageDatabase) AddMessage(msg *ReceivedMessage) (bool, error) {
	if msgdb.CheckIfSeen(msg.GetMessageID()) {
		// If the message was already seen
		// Update the message info
		err := msgdb.MessageBuffer.UpdateMessageInfo(msg)
		if err != nil {
			return true, err
		}
		// add the msg to the Validator list
		msgdb.RWMutex.Lock()
		msgList := msgdb.ValidatorMessages[msg.GetValidatorIndex()]
		msgList = append(msgList, msg.GetMessageID())
		msgdb.ValidatorMessages[msg.GetValidatorIndex()] = msgList
		// add the msg to the Slot list
		msgList = msgdb.SlotMessages[msg.GetSlot()]
		msgList = append(msgList, msg.GetMessageID())
		msgdb.SlotMessages[msg.GetSlot()] = msgList
		msgdb.RWMutex.Unlock()
		return true, nil
	} else {
		// If the message has not been seen before
		// add the message to the msgDB
		msgdb.MessageBuffer.AddNewMessage(msg)
		msgdb.RWMutex.Lock()
		// add the message to the Seen list
		msgdb.MessageSeenList[msg.GetMessageID()] = true
		// add the msg to the Validator list
		msgList := msgdb.ValidatorMessages[msg.GetValidatorIndex()]
		msgList = append(msgList, msg.GetMessageID())
		msgdb.ValidatorMessages[msg.GetValidatorIndex()] = msgList
		// add the msg to the Slot list
		msgList = msgdb.SlotMessages[msg.GetSlot()]
		msgList = append(msgList, msg.GetMessageID())
		msgdb.SlotMessages[msg.GetSlot()] = msgList
		msgdb.RWMutex.Unlock()
		return false, nil
	}
}

func (msgdb *MessageDatabase) GetMessageInfo(msgID string, msgInfo *MessageInfo) error {
	// Check if the message has been alredy seen
	if msgdb.MessageSeenList[msgID] {
		err := msgdb.MessageBuffer.GetMessageInfo(msgID, msgInfo)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("The message %s wasn't seen and therefore couldn't be found", msgID)
}

func (msgdb *MessageDatabase) GetValidators() []beacon.ValidatorIndex {
	var valList []beacon.ValidatorIndex
	for validator, _ := range msgdb.ValidatorMessages {
		valList = append(valList, validator)
	}
	return valList

}

func (msgdb *MessageDatabase) GetMessageIDsFromValidator(validatorIndex beacon.ValidatorIndex) []string {
	var msgList []string
	msgList = msgdb.ValidatorMessages[validatorIndex]
	return msgList
}

func (msgdb *MessageDatabase) GetSlots() []beacon.Slot {
	var slotList []beacon.Slot
	for slot, _ := range msgdb.SlotMessages {
		slotList = append(slotList, slot)
	}
	return slotList
}

// if empty, returns an empty array
func (msgdb *MessageDatabase) GetMessageIDsFromSlot(slot beacon.Slot) []string {
	var msgList []string
	msgList = msgdb.SlotMessages[slot]
	return msgList
}

func (msgdb *MessageDatabase) DBExportPath() string {
	return msgdb.DiskDBPath
}

// Return True if the message was seen before
func (msgdb *MessageDatabase) CheckIfSeen(msgID string) bool {
	return msgdb.MessageSeenList[msgID]
}

// ---- Message Buffer ----

// Intended to be like a cache of received messages
// TODO: Change the Hard-coded Mainnet config from the TopicDB to here (so that every topic can have its own Spec)
type MessageBuffer struct {
	MessageList  sync.Map // map[messageID(string)]ReceivedMessage
	Buffer       *CircularBuffer
	DiskDatabase *DiskDatabase // Path for the phisical DB of messages
	// Spec *beacon.Spec
}

func NewMessageBuffer(msgLimit int, diskDBPath string) *MessageBuffer {
	mdb := &MessageBuffer{
		Buffer:       NewCircularBuffer(msgLimit),
		DiskDatabase: NewDiskDatabase(diskDBPath),
	}
	return mdb
}

func (mbf *MessageBuffer) GetMessageInfo(msgID string, mi *MessageInfo) error {
	// check if the msg is on the cache
	if mbf.Buffer.MsgOnBuffer(msgID) {
		//fmt.Println("DEBUG - Message", msgID, "was on cache (buffered)")
		bufferedMsg, ok := mbf.MessageList.Load(msgID)
		if !ok {
			return fmt.Errorf("Error Loading msg %s from the Buffer", msgID)
		}
		mi = bufferedMsg.(*MessageInfo)
	} else {
		//fmt.Println("DEBUG - Message", msgID, "was NOT cache (buffered), reading DB")
		err := mbf.DiskDatabase.Read(msgID, mi)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mbf *MessageBuffer) AddNewMessage(msg *ReceivedMessage) error {
	/* --- Deprecated because when the MessageDatabase is generated all the messages on the
	// We should have checked before if the msg was on the seen list
	// might be that the msg is on the DiskDatabase because there was an error and we restarted the Rumor Session
	if mbf.DiskDatabase.MessageExist(msg.ID()){
		fmt.Println("DEBUG - The message was already existing on the Disk DB, reading the file")
		var msgInfo MessageInfo
		err := mbf.DiskDatabase.Read(msg.ID(), &MessageInfo)
		if err != nil {
			return err
		}
		MessageInfo.AddNewSender(msg)
		// add the msg to the Buffer and to the Message Caché
		mbf.MessageList.Store(msg.ID(), &MessageInfo)
		full, oldestId := mbf.Buffer.Write(msg.ID())
		if full {
			// read the oldest message from the cache of msgs
			bufferedMsg, ok := mbf.MessageList.Load(oldestId)
			if !ok {
				return fmt.Errorf("Error Loading msg %s from the Buffer", oldestId)
			}
			mbf.DiskDatabase.Write(bufferedMsg)
			// delete the oldmsg from the msg cache
			delete(mbf.MessageList, oldestId)
			return nil
		}
		// if the buffer isn't full,
		return nil
	}
	*/
	//fmt.Println("DEBUG - The message doesn't exist, saving it into")
	// if the message isn't in the Disk Database,
	// just create a new MessageInfo template and store it in the msg cache
	msgInfo := NewMessageInfo(msg)
	mbf.MessageList.Store(msg.GetMessageID(), msgInfo)
	full, oldestId := mbf.Buffer.Write(msg.GetMessageID())
	if full {
		// read the oldest message from the cache of msgs
		bufferedMsg, ok := mbf.MessageList.Load(oldestId)
		if !ok {
			return fmt.Errorf("Error Loading msg %s from the Buffer", oldestId)
		}
		mbf.DiskDatabase.Write(bufferedMsg.(*MessageInfo))
		// delete the oldmsg from the msg cache
		mbf.MessageList.Delete(oldestId)
		return nil
	}
	// if the buffer isn't full,
	//fmt.Println("DEBUG - The message has been added, still space in the buffer")
	return nil
}

// Update the message info of a message that we already sawfmt.
func (mbf *MessageBuffer) UpdateMessageInfo(msg *ReceivedMessage) error {
	// check if the msg is on the cache
	if mbf.Buffer.MsgOnBuffer(msg.GetMessageID()) {
		//fmt.Println("DEBUG - Message", msg.GetMessageID(), "was on cache (buffered)")
		bufferedMsg, ok := mbf.MessageList.Load(msg.GetMessageID())
		if !ok {
			return fmt.Errorf("Error Loading msg %s from the Buffer", msg.GetMessageID())
		}
		messageInfo := bufferedMsg.(*MessageInfo)
		messageInfo.AddNewMsgSender(msg)
		mbf.MessageList.Store(msg.GetMessageID(), messageInfo)
		//fmt.Println("DEBUG - Updated value of the msg", msg.GetMessageID())
	} else {
		//fmt.Println("DEBUG - Message", msg.GetMessageID(), "was NOT cache (buffered), reading DB")
		err := mbf.DiskDatabase.UpdateValue(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete the MessageInfo of a given Msg, currently not needed
func (mbf *MessageBuffer) Delete() error {
	fmt.Println("Deleting msg")
	return nil
}

// could be used to export the MessageInfo tha was in the Circular Buffer
func (mbf *MessageBuffer) Export(msgID string) error {
	//fmt.Println("Exporting buffered Msg", msgID)
	// Give as granted that will be on the buffer, otherwise stop
	bufferedMsg, ok := mbf.MessageList.Load(msgID)
	if !ok {
		return fmt.Errorf("Error Loading msg %s from the Buffer", msgID)
	}
	err := mbf.DiskDatabase.Write(bufferedMsg.(*MessageInfo))
	if err != nil {
		return err
	}
	//fmt.Println("DEBUG - Updated value of the msg", msgID)
	return nil
}
