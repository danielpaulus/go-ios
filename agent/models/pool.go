package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var ip = ""

func LoadOrInitPool() (DevicePool, error) {
	result := DevicePool{}
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Pool"))
		if b == nil {
			return errUninitializedDb
		}
		poolConfigBytes := b.Get([]byte("config"))
		e := json.Unmarshal(poolConfigBytes, &result)
		return e
	})
	if os.Getenv("POOL_ID") != "" {
		log.Warnf("Overriding pool id with %s", os.Getenv("POOL_ID"))
		result.ID = os.Getenv("POOL_ID")
	}
	if err == errUninitializedDb {
		return initPoolConfig()
	}
	result.Port = os.Getenv("HTTP_PORT")
	if ip == "" {
		ip = GetOutboundIP()
	}
	result.Ip = ip
	return result, nil
}
func GetOutboundIP() string {
	if os.Getenv("IP_OVERRIDE") != "" {
		log.Infof("Overriding IP with %s", os.Getenv("IP_OVERRIDE"))
		return os.Getenv("IP_OVERRIDE")
	}
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func initPoolConfig() (DevicePool, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return DevicePool{}, err
	}
	id := uuid.New()
	poolConfig := DevicePool{
		Hostname: hostname,
		ID:       id.String(),
		Devices:  []DeviceInfo{},
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b, errb := tx.CreateBucket([]byte("Pool"))
		if errb != nil {
			return fmt.Errorf("create bucket: %s", errb)
		}
		listj, errj := json.Marshal(poolConfig)
		err := b.Put([]byte("config"), listj)
		return errors.Join(errj, err)
	})
	return poolConfig, err
}

func UpdatePool(pool DevicePool) error {
	log.Info("update pool")
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Pool"))
		if b == nil {
			return errUninitializedDb
		}
		listj, errj := json.Marshal(pool)
		err := b.Put([]byte("config"), listj)
		return errors.Join(errj, err)
	})
}
