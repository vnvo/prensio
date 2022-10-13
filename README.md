# Goal
to provide a simple Change Data Capture (CDC) tool to collect mysql changes and push them into kafka

# High Level Design
this tool is inspired by go-mysql-elasticsearch. The components of this sustem are: 
## Pipeline
Is the high level construct to represt the actual flow of events
## MySQL Binlog Source
The connector for mysql which reads the mysql binlog events. like go-mysql-elasticsearch here we are using canal package from go-mysql.
## Transformer
The scriptable transformer using goja package and predefined helper functions. The helper function is are drived by EMCA 5.1 compatible scripts. Basically it's an embeded javascript execution engine

## Kafka Sink
Kafka client to produce the events.

## State Manager
the package to store and load the state of the pipline between restarts and as the stream of events flows. kafka is used as the backend


# Alternatives
## Debezium or Maxwell
If you are interested in a more mature solution with.

# Why using this
However, this service can provide a more lightweight solution with a high degree of flexibility.

# How to Use it
[TBD]