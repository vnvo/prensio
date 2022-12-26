
# Prensio
Is a simple yet flexible [Change Data Capture](https://en.wikipedia.org/wiki/Change_data_capture) (CDC) tool. This tool collects mysql changes (insert, update, delete), runs a user supplied transformation logic on them and writes the result into kafka. This tool is inspired by go-mysql-elasticsearch and debezium.

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
![Components](/docs/prensio-components.png?raw=true "Components")

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
	cdc_event.after.forEach((e) => {
		if (e.text_col.length > 3) {
		e.text_col = tObfString(e.text_col)
		}
	});
return ACTION_CONT
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
#### After Transform Action
Transform functions can decide the fate of the individual events. At the moment these actions are supported:
- **ACTION_CONT**: to continue the processing after the transformation.
- **ACTION_DROP**: to drop the current event and pass the next event to the transform function.

NB: Prensio checks the return value of the transform function aginst those constants. Your transform logic should return one of those values.
You can make the decision based on the content of the event or other factors, for example:
```javascript
function transform(cdc_event) {
    if (cdc_event.after[0].text_col == "should-continue")
        return ACTION_CONT

    return ACTION_DROP
}
```

#### User Defined Transformers
Beside having helper functions in the script, If your use case requires extra processing, collecting data from external sources and anything that is not readily available, you can add extra helper/utility functions. Such functions are implemented in Go and are made available to the transform runtime.
For examples see: [TBD]

#### Prensio Event Schema
Pernsio events has this schema and this how structure for the **cdc_event passed to transform functions as well as the final kafka message**:
|  Field | Type | Description  |
| --- | --- | --- |
| schema  | string | name of the mysql schema the event coming from |
| table | string | table name for the mysql table the event coming from  |
| action  | string | one of `insert`, `update`, `delete`  (always lowecase) |
| before  | list of objects  | a list of old values for each affected row. every item is a object/map containing all columns and their value before the change |
| after | list of objects | a list of new values for each affected row. every item is a object/map containing all columns and their value after the change |
| meta | object | predefined map to carry extra data for each event. current items are `pipeline` (name of the prensio pipline producing the event) and `timstamp` (timestamp at which prensio started processing this event, in microseconds) |


### Kafka Sink
Kafka client to produce the events. This can produce to multiple topics.
#### Kafka Topic Creation
Current behavior is to try and create the missing kafka topics. Support for making this configurable will be added.

### State Manager
The package to store and load the state of the pipline between restarts and as the stream of events flows. kafka is used as the backend.



## Alternatives
### Debezium or Maxwell
If you are interested in a more mature solution with a sizeable community you should checkout debezium and maxwell at least.


### Why use Prensio
[TBD]
  

## How to Use
[TBD]

## How to Develop
[TBD]

