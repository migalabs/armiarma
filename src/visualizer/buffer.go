package visualizer

import (
    "fmt"
    "sync"
)

var MinLen int = 5

type Buffer struct {
    Items   []interface{}
    WPointer int
    Len int
    sync.RWMutex
}

func NewBuffer(length int) *Buffer{
    if length == 0 {
        fmt.Println("Length for the new Buffer wasn't specified, setting default (5 items)")
        length = MinLen
    }
    b := &Buffer {
        Items: make([]interface{}, length),
        WPointer: 0,
        Len: length,
    }
    return b
}

func (c *Buffer) AddItem(item interface{}) error{
    c.RWMutex.Lock()
    defer c.RWMutex.Unlock()
    fmt.Println("Adding New item on item:", c.WPointer)
    if item == nil {
        return fmt.Errorf("Given Item was empty")
    }
    c.Items[c.WPointer] = item
    c.WPointer = c.WPointer + 1
    if c.WPointer >= c.Len {
        c.WPointer = 0
    }
    return nil
}

// Returns the iteration throught the items of the Buffer starting from the latest one
// might make more sense to start from the newest one
// Usage, the Range functions should be able to work with:
// for item := Buffer.Range() {} 
func (c Buffer) Range() chan interface{} {
    auxp := c.WPointer
    chn := make(chan interface{})
    go func() {
        c.RWMutex.Lock()
        for i := 1; i<=c.Len; i++{
            auxp = auxp + i
            if auxp >= c.Len {
                auxp = auxp - c.Len
            }
            chn <- c.Items[auxp]
        }
        c.RWMutex.Unlock()
        close(chn)
    }()
    return chn
}

func (c Buffer) GetItem(itemIdx int, item interface{}) (exists bool, ok bool){
    _ = itemIdx
    _ = item
    return false, true
}
