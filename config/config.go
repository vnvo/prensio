package config

import (
	"fmt"
	"io/ioutil"

	"github.com/pelletier/go-toml/v2"
)

type MySQLSourceConfig struct {
	Name     string `toml:"name"`
	Addr     string `toml:"addr"`
	Port     int16  `toml:"port"`
	User     string `toml:"user"`
	Pass     string `toml:"pass"`
	ServerId int16  `toml:"server_id"`
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

type CDCConfig struct {
	Mysql          MySQLSourceConfig `toml:"mysql"`
	SourceRules    SourceRule        `toml:"source_rules"`
	TransformRules []TransformRule   `toml:"transform_rule"`
	KafkaSink      KafkaSink         `toml:"kafka_sink"`
}

func NewCDCConfig(path string) CDCConfig {
	var conf CDCConfig

	data, err := ioutil.ReadFile(path)
	fmt.Printf("%s - %v\n", data, err)
	err = toml.Unmarshal(data, &conf)

	if err != nil {
		panic(err)
	}

	fmt.Printf("MySQL = %v\n", conf.Mysql)
	fmt.Printf("Source Rules = %v\n", conf.SourceRules)
	fmt.Printf("Transform Rule = %v\n", conf.TransformRules)
	fmt.Printf("Kafka Sink = %v\n", conf.KafkaSink)

	return conf
}
