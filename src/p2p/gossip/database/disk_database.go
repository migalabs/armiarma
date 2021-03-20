package database

import (
	"fmt"
	"os"
	"encoding/json"
	"io/ioutil"
	"strings"
)

// currently the msgInfo will be exported as json
type DiskDatabase struct {
	Path string
}

func NewDiskDatabase(path string) *DiskDatabase {
	ddb := &DiskDatabase {
		Path: path,
	}
	return ddb
}

// ---  DiskDatabase related Methods  -----

func (ddb *DiskDatabase) MessageExist(msgID string) bool {
	fileName := ddb.Path + "/" + msgID + ".json"
	if FileExists(fileName) {
		return true
	}
	return false
}

func (ddb *DiskDatabase) Read(msgID string, msgInfo *MessageInfo) error {
	fileName := ddb.Path + "/" + msgID + ".json"
	// Check if file exist
	if FileExists(fileName) { // if exists, read it
		// get the json of the file
		jsonFile, err := os.Open(fileName)
		defer jsonFile.Close()
		if err != nil {
			return err
		}
		byteValue, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			return err
		}
		json.Unmarshal(byteValue, &msgInfo)
		return nil
	}
	return fmt.Errorf("File: %S doesn't exist", fileName)	
}

func (ddb *DiskDatabase) Write(msgInfo *MessageInfo) error {
	fileName := ddb.Path + "/" + msgInfo.GetMessageID() + ".json"
	mm, err :=json.Marshal(*msgInfo)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fileName, mm, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Currently no needed
func (ddb *DiskDatabase) Delete(msgID string) {
	fmt.Println("Delete msg ", msgID)
}

// right now only supports to add a new sender and the arrival time
func (ddb *DiskDatabase) UpdateValue(msg *ReceivedMessage) error {
	var msgInfo MessageInfo
	err := ddb.Read(msg.GetMessageID(), &msgInfo)
	if err != nil {
		return err
	}
	msgInfo.AddNewMsgSender(msg)
	err = ddb.Write(&msgInfo)
	if err != nil {
		return err
	}
	return nil
}

func (ddb *DiskDatabase) MessageIDs() []string {
	var msgIDs []string 
	files, err := ioutil.ReadDir(ddb.Path)
    if err != nil {
        fmt.Println("DEBUG (MessageIDs) error reading the content of the Disk Database content, path:", ddb.Path)
    }
    for _, f := range files {
		msgID := strings.Trim(f.Name(), ".json")
		msgIDs = append(msgIDs, msgID)
    }
    return msgIDs
}

func (ddb *DiskDatabase) TotalMessages() int{
	files, err := ioutil.ReadDir(ddb.Path)
    if err != nil {
        fmt.Println("DEBUG (TotalMessages) error reading the content of the Disk Database content, path:", ddb.Path)
    }
    return len(files)
}

// ----  Auxiliar functions  ----
 
// Exists reports whether the named file or directory exists.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// 
