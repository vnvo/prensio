package main

import (
	"fmt"

	"github.com/vnvo/go-mysql-kafka/config"
)

func main() {
	conf := config.NewCDCConfig("./go_mysql_kafka.toml")
	fmt.Println(conf)

}
