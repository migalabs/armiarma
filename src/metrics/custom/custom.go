package custom

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type CustomMetrics struct {
	StartTime TimeStruct
	StopTime TimeStruct
	PeerStore PeerStore
}

func NewCustomMetrics() CustomMetrics {
	cm := CustomMetrics {
		StartTime: NewTimeStruct(),
		StopTime: NewTimeStruct(),
		PeerStore: NewPeerStore(),
	}
	// Since we initialize the CustomMetrucs at the beginning
	// we can already stamp the initial time
	cm.StartTime.StampCurrentTime()
	return cm
}

func (cm *CustomMetrics) ExportJson(jsonPath string) error {
	// Fullfill the Stop-data of the CustomMetrics struct
	cm.StopTime.StampCurrentTime()
	// Marshal / Serialize the struct 
	bytes, err := json.Marshal(cm)
	if err != nil {
		fmt.Println("Error Marshalling the metrics", err)
	}
	// write the bytes into a given path
	err = ioutil.WriteFile(jsonPath, bytes, 0644)
	if err != nil {
		fmt.Println("Error opening file: ", jsonPath)
		return err
	}
	return nil
}