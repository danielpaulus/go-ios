package api

import (
	"encoding/json"
	"io/ioutil"
)

//GetVersion reads the contents of the file version.txt and returns it.
//If the file cannot be read, it returns "could not read version"
func GetVersion() string {
	version, err := ioutil.ReadFile("version.txt")
	if err != nil {
		return "could not read version"
	}
	return string(version)
}

func MustMarshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
