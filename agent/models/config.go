package models

import (
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"sync"
)

type Config struct {
	ResignerUrl     string
	ProfilePassword string
	Hostname        string
	PoolId          uuid.UUID
}

var configMux = sync.Mutex{}
var currentconfig = Config{
	ResignerUrl:     "",
	ProfilePassword: "",
	Hostname:        "localhost",
	PoolId:          uuid.UUID{},
}

func GetConfig() (Config, error) {
	configMux.Lock()
	defer configMux.Unlock()
	pool, err := LoadOrInitPool()
	if err != nil {
		return Config{}, err
	}
	currentconfig.PoolId = uuid.MustParse(pool.ID)
	return currentconfig, nil
}

func GetPoolId() uuid.UUID {
	pool, err := LoadOrInitPool()
	if err != nil {
		log.Errorf("failed to load pool: %v", err)
	}
	return uuid.MustParse(pool.ID)
}
func UpdateConfigFromCloud(config Config) {
	configMux.Lock()
	defer configMux.Unlock()
	if currentconfig.ResignerUrl != config.ResignerUrl {
		log.WithField("new", config.ResignerUrl).
			WithField("old", currentconfig.ResignerUrl).
			Info("updating resigner URL from cloud")
		currentconfig.ResignerUrl = config.ResignerUrl
	}
	if currentconfig.ProfilePassword != config.ProfilePassword {
		log.Info("updating profile password from cloud")
		currentconfig.ProfilePassword = config.ProfilePassword
	}
}
