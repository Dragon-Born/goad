package watcher

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"goad/database"
	"goad/pkg/yaml"
)

func init() {
	config, err := yaml.NewConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error loading initial config: %v", err)
	}
	database.Config = config
	level, err := log.ParseLevel(database.Config.LogLevel)
	if err == nil {
		log.SetLevel(level)
	}
	log.Info("Initial configuration loaded")
}

func WatchConfig() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func(watcher *fsnotify.Watcher) {
		err = watcher.Close()
		if err != nil {
			log.Error(err)
		}
	}(watcher)
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Info("Configuration file modified. Reloading...")
					newConfig, err := yaml.NewConfig("config.yaml")
					if err != nil {
						log.Errorf("Error reloading config: %v", err)
						continue
					}
					database.Config = newConfig
					log.Info("Configuration reloaded successfully")
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error("Error:", err)
			}
		}
	}()

	err = watcher.Add("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
