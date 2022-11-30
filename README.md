
# go-mysql-kafka
to provide a simple yet flexible [Change Data Capture](https://en.wikipedia.org/wiki/Change_data_capture) (CDC) tool. This tool collects mysql changes, runs a user supplied transformation function on them and pushes them into kafka. This tool is inspired by go-mysql-elasticsearch and debezium.

## Status: WIP

## Features
 - [ ] Supports percona 8 and mysql 8
 - [ ] Flexible transformation using a embedded javascript engine
 - [ ] Run multiple pipelines per instance
 - [ ] Track the last state and ability to seek using GTID
 - [ ] Simple to config and use
 - [ ] REST APIs to manage the runtime
 - [ ] Resilient to failures from both ends
 - [ ] Cloud ready
  Jump to **How to Use**

## Components
The components of this system are:

### Pipeline
The high level construct to manage the flow of events from a source (mysql) to a sink (kafka). Pipeline(s) are created at startup using the supplied configuration files but you can use the REST APIs to build automation and dynamically manage the pipelines at runtime.
[more info TBD]

### MySQL Binlog Source
The connector for mysql which reads the mysql binlog events. like go-mysql-elasticsearch the canal package from go-mysql is used.
[more info TBD]

### Transformer
The scriptable transformer using goja package. You will define a transform function which will receive the event as the input and can manipulate it. Example:
```javascript
function transform(cdc_event) {
	cdc_event.ext = 1
	cdc_event.after.forEach((e) => {
		e.changed = 1
		if (e.text_col.length > 3) {
		e.text_col = tObfString(e.text_col)
		}
	});
return cdc_event
}
```
the provided script should always have a function with this signature.
```javascript
function transform(cdc_event)
```
**cdc_event** is the actual event to be processed which is passed by reference meaning any change will be reflected in the actual event.
[more details about the return value to indicate "drop" or "continue"]
your transform script can be as extensive as you want it to be, to the full extent of ECMA 5.1.
**tObfString** is a user defined transformer, defined as a Go function which is made available to the transform function. Read the next section.

#### User Defined Transformers
Beside having helper functions in the script, If your use case requires extra processing, collecting data from external sources and anything that is not readily available, you can add extra helper/utility functions. Such functions are implemented in Go and are made available to the transform runtime.
For examples see: [TBD]

### Kafka Sink
Kafka client to produce the events. This can produce to multiple topics.
#### Kafka Topic Creation
Current behavior is to try and create the missing kafka topics. Support for making this configurable will be added.

### State Manager
The package to store and load the state of the pipline between restarts and as the stream of events flows. kafka is used as the backend.



## Alternatives
### Debezium or Maxwell
If you are interested in a more mature solution with a sizeable community you should checkout debezium and maxwell at least.

  
### Why use go-mysql-kafka then
[TBD]
  

## How to Use
[TBD]

## How to Develop
[TBD]

