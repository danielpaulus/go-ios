package service

import "io/ioutil"

//GetVersion reads the contents of the file version.txt and returns it.
//If the file cannot be read, it returns "could not read version"
func GetVersion() string {
	version, err := ioutil.ReadFile("version.txt")
	if err != nil {
		return "could not read version"
	}
	return string(version)
}
