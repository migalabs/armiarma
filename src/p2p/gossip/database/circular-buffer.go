package database

import (
    "fmt"
    "github.com/protolambda/zrnt/eth2/beacon"
)

// Definition of the struck that will contain the Circular Buffers for the messages received on the Gossip Topics 
type CircularBuffer struct {
    List        map[int]beacon.Root
    ReadP       int
    WriteP      int
    Limit       int
}

// Gen and Initialize a Circular Buffer
func NewCircularBuffer (messageLimit int) CircularBuffer{
    cb := CircularBuffer{
        List:   make(map[int]beacon.Root),
        ReadP:  0,
        WriteP: 0,
        Limit:  messageLimit,
    }
    return cb
}

// Obtain the root of the next message on the reading queue
func (c *CircularBuffer) Read() (beacon.Root, error){
    // Read the root of the next Message on the list (c.ReadP points to it)
    root := c.List[c.ReadP]
    // Check if the there is anything new to Read
    if c.ReadP == c.WriteP {
        return root, fmt.Errorf("there isn't any new content to read")
    }
    // We update the c.ReadP pointer to read the next on in the list
    c.ReadP += 1
    // if the pointer exceeds the limit, back to 0
    if c.ReadP >= c.Limit {
        c.ReadP = 0
    }
    // Once we have the root, the message can be obtained from the Database and then deleted
    return root, nil
}

// Write new root on the Circular Buffer
// Will return true and the root value of the oldest message if the circular was Full
// Will return False if the buffer is not full 
func (c *CircularBuffer) Write(root beacon.Root) (bool, interface{}) {
    var full bool
    var oldestRoot beacon.Root
    // Check if the following incrementation will cause a buffer rebase
    auxP := c.WriteP + 1
    if auxP >= c.Limit {
        auxP = 0
    }
    if auxP == c.ReadP {
        fmt.Println("WARNING: the buffer will rebase, consider increasing the temporary database size, Last block will be lost", c.List[c.ReadP])
        full = true
        oldestRoot = c.List[c.ReadP]
        // so far we will just keep the latest x messages, so the ReadP pointer will be increased aswell, loosing the latest root on the list
        c.ReadP += 1
        if c.ReadP >= c.Limit {
            c.ReadP = 0
        }
    } else {
        full = false
        //oldestRoot = false
    }
    // NOTE: -Might be interesting to add certaing flags to know when there was anything to read and when the buffer got full
    c.List[c.WriteP] = root
    fmt.Println("Included root on circular buffer")
    c.WriteP += 1
    // if the pointer exceeds the limit, back to 0
    if c.WriteP >= c.Limit {
        c.WriteP = 0
    }
    fmt.Println("Map:", len(c.List))
    fmt.Println("WriteP:", c.WriteP)
    fmt.Println("ReadP:", c.ReadP)
    return full, oldestRoot
}

// Check if the root is already at the circular buffer
// GossipSub already has a cache of previously received messages,
// but since we pretend to change that, we might not keep more than once the same block
// With this we achieve a proper message count, but we just save it once
func (c *CircularBuffer) Already(root beacon.Root) bool{
    for _, value := range c.List {
        if value == root {
            return true
        }
    }
    return false
}
