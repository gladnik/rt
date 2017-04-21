package config

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

// Config file struct
type Containers map[string]Container

type Container struct {
	Image   string            `json:"image"`
	DataDir string            `json:"dataDir"`
	Tmpfs   map[string]string `json:"tmpfs"`
}

// Config current configuration
type Config struct {
	lock       sync.RWMutex
	Containers map[string]Container
	LogConfig  *container.LogConfig
	DataDir    string
	Timeout    time.Duration
}

// NewConfig creates new config
func NewConfig(dataDir string, timeout time.Duration) *Config {
	return &Config{Containers: make(map[string]Container), LogConfig: new(container.LogConfig), DataDir: dataDir, Timeout: timeout}
}

func (c *Config) Load(containers, containerLogs string) error {
	log.Println("Loading configuration files...")
	ct := make(Containers)
	err := loadJSON(containers, &ct)
	if err != nil {
		return fmt.Errorf("browsers config: %v", err)
	}
	log.Printf("Loaded configuration from [%s]\n", containers)
	var cl *container.LogConfig
	err = loadJSON(containerLogs, &cl)
	if err != nil {
		log.Printf("Using default containers log configuration because of: %v\n", err)
		cl = &container.LogConfig{}
	} else {
		log.Printf("Loaded log configuration from [%s]\n", containerLogs)
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Containers, c.LogConfig = ct, cl
	return nil
}

func loadJSON(filename string, v interface{}) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Failed to read config file: %v", err)
	}
	if err := json.Unmarshal(buf, v); err != nil {
		return fmt.Errorf("Failed to read config file: %v", err)
	}
	return nil
}

func (c *Config) GetContainer(containerType string) (*Container, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c, ok := c.Containers[containerType]; ok {
		return &c, true
	}
	return nil, false
}
