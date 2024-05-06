package yaml

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	TelegramBot *TelegramBotConfig `yaml:"telegram_bot"`
	Database    *DatabaseConfig    `yaml:"database"`
	DataSeed    string             `yaml:"data_seed"`
	LogLevel    string             `yaml:"log_level"`
	Tokens      []*TokenConfig     `yaml:"tokens"`
}

type TelegramBotConfig struct {
	Token           string `yaml:"token"`
	AnnounceChannel int64  `yaml:"announce_channel"`
}

type DatabaseConfig struct {
	CurrentMode string                `yaml:"current_mode"`
	DBs         map[string]DatabaseDB `yaml:"dbs"`
}

type DatabaseDB struct {
	Type string `yaml:"type"`
	URI  string `yaml:"uri"`
}

type TokenConfig struct {
	Address string   `yaml:"address"`
	Price   float64  `yaml:"price"`
	Chain   string   `yaml:"chain"`
	Wallets []string `yaml:"wallets"`
	Ratio   float64  `yaml:"ratio"`
	Active  bool     `yaml:"active"`
	Sleep   string   `yaml:"sleep"`
	Counter int
}

func (t *TokenConfig) GetSleep() (time.Duration, error) {
	bounds := strings.Split(t.Sleep, "-")
	if len(bounds) != 2 {
		return 0, fmt.Errorf("invalid sleep range format")
	}
	min, err := strconv.Atoi(bounds[0])
	if err != nil {
		return 0, fmt.Errorf("invalid minimum sleep value")
	}
	max, err := strconv.Atoi(bounds[1])
	if err != nil {
		return 0, fmt.Errorf("invalid maximum sleep value")
	}
	if min >= max {
		return 0, fmt.Errorf("minimum sleep value must be less than maximum")
	}
	sleepDuration := rand.Intn(max-min+1) + min
	return time.Duration(sleepDuration), nil
}

func NewConfig(filePath string) (*Config, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
