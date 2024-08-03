package main

import (
	"log"
	"os"

	"github.com/danielpaulus/go-ios/agent"
	"github.com/danielpaulus/go-ios/agent/models"
)

func main() {

	os.MkdirAll("pool-data", os.ModePerm)
	closeFunc, err := models.InitDb("pool-data/my.db")
	if err != nil {
		log.Fatalf("error creating local db: %v", err)
	}
	defer closeFunc()

	agent.Main()
}
