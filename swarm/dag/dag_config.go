package dag

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type DagConfig struct {

	Version 			string  			`json:"version"`
	StorageLocation 	string 				`json:"storage_location"`
	DagNodes			DagNodes 			`json:"nodes"`
}

func LoadConfigFromJsonFile(jsonFileName string) *DagConfig {

	jsonFile, err := os.Open(jsonFileName)

	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	config := new(DagConfig)
	json.Unmarshal(byteValue, config)

	return config
}
