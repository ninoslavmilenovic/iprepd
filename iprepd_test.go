package iprepd

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func baseTest() error {
	_, err := sruntime.redis.flushAll().Result()
	if err != nil {
		return err
	}
	sruntime.cfg.Decay.Points = 0
	sruntime.cfg.Decay.Interval = time.Minute
	r := Reputation{
		Object:     "192.168.0.1",
		Type:       "ip",
		Reputation: 50,
	}
	err = r.set()
	if err != nil {
		return err
	}
	r = Reputation{
		Object:     "10.0.0.1",
		Type:       "ip",
		Reputation: 25,
	}
	err = r.set()
	if err != nil {
		return err
	}
	// Add a legacy format reputation entry for testing, needs to be added
	// manually to bypass the insertion validator
	r = Reputation{
		IP:          "254.254.254.254",
		Reputation:  40,
		LastUpdated: time.Now().UTC(),
	}
	buf, err := json.Marshal(r)
	if err != nil {
		return err
	}
	err = sruntime.redis.set(r.IP, buf, time.Hour*336).Err()
	if err != nil {
		return err
	}
	return nil
}

func TestLoadSampleConfig(t *testing.T) {
	_, err := loadCfg("./iprepd.yaml.sample")
	assert.Nil(t, err)
}

func TestMain(m *testing.M) {
	var (
		err  error
		tcfg serverCfg
	)
	tcfg.Redis.Addr = "127.0.0.1:6379"
	err = tcfg.validate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	renv := os.Getenv("IPREPD_TEST_REDISADDR")
	if renv != "" {
		tcfg.Redis.Addr = renv
	}
	sruntime.statsd, err = newStatsdClient(tcfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	sruntime.redis, err = newRedisLink(tcfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	sruntime.cfg.Auth.Hawk = map[string]string{"root": "toor", "user": "secret"}
	sruntime.cfg.Auth.APIKey = map[string]string{"u1": "key1", "u2": "key2"}
	sruntime.cfg.Exceptions.File = []string{"./testdata/exceptions.txt"}
	sruntime.cfg.Exceptions.AWS = true
	sruntime.cfg.Decay.Points = 0
	sruntime.cfg.Decay.Interval = time.Minute
	sruntime.cfg.Violations = []Violation{
		{"violation1", 5, 25},
		{"violation2", 50, 50},
		{"violation3", 0, 0},
	}
	loadExceptions()
	os.Exit(m.Run())
}
