package config

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/siddontang/go-log/log"
)

type MySQLSourceConfig struct {
	Name     string `toml:"name"`
	Addr     string `toml:"addr"`
	Port     int16  `toml:"port"`
	User     string `toml:"user"`
	Pass     string `toml:"pass"`
	ServerId uint32 `toml:"server_id"`
}

type SourceRule struct {
	IncludeSchemas []string `toml:"include_schemas"`
	ExcludeSchemas []string `toml:"exclude_schemas"`
	IncludeTables  []string `toml:"include_tables"`
	ExcludeTables  []string `toml:"exclude_tables"`
}

type TransformRule struct {
	Name          string
	MatchSchema   string `toml:"match_schema"`
	MatchTable    string `toml:"match_table"`
	TransformFunc string `toml:"transform_func,multiline"`
}

type KafkaSink struct {
	Name           string `toml:"name"`
	Addr           string `toml:"addr"`
	Port           int16  `toml:"port"`
	BatchSizeBytes int16  `toml:"batch_size_bytes"`
}

type StateTracking struct {
	KafkaAddr string `toml:"kafka_addr"`
	Topic     string `toml:"state_topic"`
}

func (ks *KafkaSink) GetAddrList() []string {
	return strings.Split(ks.Addr, ",")
}

type CDCConfig struct {
	Mysql          MySQLSourceConfig `toml:"mysql"`
	SourceRules    SourceRule        `toml:"source_rules"`
	TransformRules []TransformRule   `toml:"transform_rule"`
	KafkaSink      KafkaSink         `toml:"kafka_sink"`
	State          StateTracking     `toml:"state_tracking"`
}

func NewCDCConfig(path string) (CDCConfig, error) {
	var conf CDCConfig

	data, err := os.ReadFile(path)
	if err != nil {
		return conf, err
	}

	log.Debugf("config. data: \n%s - err:%v\n", data, err)
	err = toml.Unmarshal(data, &conf)

	if err != nil {
		panic(err)
	}

	return conf, nil
}

func NewCDCConfigList(path string) ([]CDCConfig, error) {
	cdcConfs := make([]CDCConfig, 0)

	cfgFiles, err := getCfgFiles(path)
	if err != nil {
		return nil, err
	}
	log.Infof("config files loaded: %d", len(cfgFiles))

	for _, cfg := range cfgFiles {
		cdcConf, err := NewCDCConfig(cfg)
		if err != nil {
			return nil, err
		}

		cdcConfs = append(cdcConfs, cdcConf)
	}

	return cdcConfs, nil
}

func getCfgFiles(confPath string) ([]string, error) {
	cfgFiles := make([]string, 0)
	var cfg string

	err := filepath.WalkDir(
		confPath,
		func(p string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}

			if filepath.Ext(d.Name()) != ".toml" {
				log.Infof("ignoring: %s", d.Name())
				return nil
			}

			cfg = filepath.Join(confPath, d.Name())
			log.Infof("discovered cdc config: %s", cfg)
			cfgFiles = append(cfgFiles, cfg)

			return nil
		})

	if err != nil {
		return nil, err
	}

	return cfgFiles, nil
}
