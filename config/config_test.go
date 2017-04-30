package config

import (
	"testing"
	"time"
	. "github.com/aandryashin/matchers"
	"github.com/docker/docker/api/types/container"
)

const (
	dataDir = "test-dir"
	timeout = 2 * time.Hour
	shutdownTimeout = 5 * time.Minute
)

var (
	config = NewConfig(dataDir, timeout, shutdownTimeout)
)

func TestLoadCorrectConfig(t *testing.T) {
	err := config.Load("test-config.json", "test-log-config.json")
	AssertThat(t, err, Is{nil})
	AssertThat(t, config.DataDir, EqualTo{dataDir})
	AssertThat(t, config.Timeout, EqualTo{timeout})
	AssertThat(t, config.ShutdownTimeout, EqualTo{shutdownTimeout})
	AssertThat(t, *config.LogConfig, EqualTo{container.LogConfig{Type: "json-file"}})
	
	ct, exists := config.GetContainer("maven")
	AssertThat(t, exists, Is{true})
	AssertThat(t, ct, Is{Not{nil}})
	
	_, exists = config.GetContainer("missing")
	AssertThat(t, exists, Is{false})
}

func TestLoadMissingConfig(t *testing.T) {
	AssertThat(t, config.Load("missing.json", "anything.json"), Is{Not{nil}})
}

func TestLoadBrokenConfig(t *testing.T) {
	AssertThat(t, config.Load("broken-config.json", "anything.json"), Is{Not{nil}})
}

func TestLoadMissingLogConfig(t *testing.T) {
	AssertThat(t, config.Load("test-config.json", "missing.json"), Is{nil})
	AssertThat(t, *config.LogConfig, EqualTo{container.LogConfig{}})
}