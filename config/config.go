package config

import (
	"io/ioutil"
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

func (ks *KafkaSink) GetAddrList() []string {
	return strings.Split(ks.Addr, ",")
}

type CDCConfig struct {
	Mysql          MySQLSourceConfig `toml:"mysql"`
	SourceRules    SourceRule        `toml:"source_rules"`
	TransformRules []TransformRule   `toml:"transform_rule"`
	KafkaSink      KafkaSink         `toml:"kafka_sink"`
}

func NewCDCConfig(path string) (CDCConfig, error) {
	var conf CDCConfig

	data, err := ioutil.ReadFile(path)
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
