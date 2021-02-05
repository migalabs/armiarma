package database

import (
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/protolambda/zrnt/eth2/beacon"
)

type ReceivedMessage interface {
	// From the stored Messages get the Time of receival
	Time() time.Time
	// From the stored Message get the ID of the peer sending the message
	From() peer.ID
	// From the stored Message get the ID of the message
	ID() interface{}
	// Get the content of the received Message
	MessageContent() interface{}
}

// Circular Databases for the received Messages on different topics
type TopicDatabase struct {
	TopicDB sync.Map
	Spec    *beacon.Spec
    NotChan map[string](chan bool)
}

func NewTopicDatabase(config *beacon.Spec) TopicDatabase {
    tdb := TopicDatabase {
        Spec:       config,
        NotChan:    make(map[string](chan bool)),
    }
    return tdb
}

// TODO: This spec will be hard-coded to the Mainnet Specifications (All the topics will use the same Spec)
// Configure the Specifications for the Received Gossip Messages (Needed to Serialize and Deserialize the Received messages)
func (c *TopicDatabase) SetSpec(spec *beacon.Spec) error {
	if spec != nil {
		c.Spec = spec
		return nil
	} else {
		return fmt.Errorf("Specifications were unable to set, Empty pointer")
	}
}

// Interface of TopicDatabase that Adds a new topic database to the General Database
func (c *TopicDatabase) NewTopic(topic string, msgLimit int) error {
	// check if a topic with the same names is already on the database map
	_, ok := c.TopicDB.Load(topic)
	if !ok { // If it doesn't exist
		db := &MessagesDB{
			MessageList: make(map[beacon.Root]ReceivedMessage),
			Buffer:      NewCircularBuffer(msgLimit),
		}
		// Include the topic to the sync.Map
		c.TopicDB.Store(topic, db)
		// generate the Notification Channel if it doesn't exist
        if _, ok := c.NotChan[topic]; !ok {
            c.NotChan[topic] = make(chan bool, 5) // 5 is the buffer size that for the channel
        }
        return nil
	} else {
		return fmt.Errorf("cannot create a new database for topic: %s, it already exists", topic)
	}
}

// Remove topic from the TopicDB
func (c *TopicDatabase) RemoveTopic(topic string) error {
	_, ok := c.TopicDB.Load(topic)
	if ok { // If it doesn't exist
		c.TopicDB.Delete(topic)
		return nil
	} else {
		return fmt.Errorf("cannot remove the database of the topic: %s, it doesn't exists", topic)
	}
}

// will receive a Message if there was anything to read, or an error if there wasn't anything to read
func (c *TopicDatabase) ReadMessage(topic string) (interface{}, error) {
	value, ok := c.TopicDB.Load(topic)
	if ok { // If message exists
		msgdb := value.(*MessagesDB)
		message, err := msgdb.Read()
		if err != nil {
			return nil, fmt.Errorf("cannot read from the temporary database")
		}
		return message, nil
	} else { // If not
		return nil, fmt.Errorf("cannot remove the database of the topic: %s, it doesn't exists", topic)
	}
}

// Will add a Message to the database of the given topic
func (c *TopicDatabase) WriteMessage(msg ReceivedMessage, topic string) error {
	value, ok := c.TopicDB.Load(topic)
	if !ok { // If it doesn't exist
		return fmt.Errorf("cannot read from the database because there is no database for the topic: %s", topic)
	} else {
		msgbd := value.(*MessagesDB)
		msgbd.Write(msg)
        // Send "true" notification through the NotChan so that the gossip-import can know that there is a new message
        c.NotChan[topic] <- true // if it crashes means that it wasn't initialized
	}
	return nil
}

// Beacon block temporary database where X items will be stored
// Intended to be like a cache of received messages
// one of each topic of messages
// TODO: Change the Hard-coded Mainnet config from the TopicDB to here (so that every topic can have its own Spec)
type MessagesDB struct {
	MessageList map[beacon.Root]ReceivedMessage
	Buffer      CircularBuffer
	// Spec *beacon.Spec
}

// TODO: Function that returns back the spec of the MessagesDB
// func (c *MessagesDB) Spec() *beacon.Spec {}

func (c *MessagesDB) Read() (ReceivedMessage, error) {
	root, err := c.Buffer.Read()
	if err != nil {
		return nil, fmt.Errorf("cannot read any message from the circular buffer")
	}
	msg, ok := c.MessageList[root]
	if !ok {
		return nil, fmt.Errorf("problem reading message %s on the database, Message with that root was not found", root)
	}
	// once the item has been loaded we delete it from the map (both the DB and the Buffer)
	delete(c.MessageList, root)
	return msg, nil
}

func (c *MessagesDB) Write(msg ReceivedMessage) error {
	// Obtain the message root
	rt := msg.ID()
	root := rt.(beacon.Root)
	// check if the root is already at the database (check the Circular Buffer is the root is recorded)
	already := c.Buffer.Already(root)
	if already {
		return fmt.Errorf("the message couldn't be added to the temporary database because it was already there")
	}
	// If the message was not there, add it to the Circular Buffer and to the List of messages
	full, oldestRoot := c.Buffer.Write(root)
	if full != false { // The buffers is full and the oldest message needs to be deleted
		delete(c.MessageList, oldestRoot.(beacon.Root))
		fmt.Println("oldest message with Root:", oldestRoot, "was deleted from Temp Database")
	}
	c.MessageList[root] = msg
	fmt.Println("Buffer Len:", len(c.MessageList))
	fmt.Println("Message", root, "was included to the database")
	return nil
}

func (c *MessagesDB) Delete(root beacon.Root) error {
	_, ok := c.MessageList[root]
	if !ok {
		return fmt.Errorf("problem deleting message %s on the database, Message with that root was not found", root)
	}
	delete(c.MessageList, root)
	return nil
}
