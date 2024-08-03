package models

import (
	"errors"
	"time"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

var db *bolt.DB
var errUninitializedDb = errors.New("db not initialized")

func InitDb(dbFileName string) (func(), error) {
	dbInstance, err := bolt.Open(dbFileName, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, err
	}
	db = dbInstance
	return func() {
		err := db.Close()
		if err != nil {
			log.Warnf("error closing db: %v", err)
		}
	}, nil
}
