
# Prensio
Is a simple and flexible [Change Data Capture](https://en.wikipedia.org/wiki/Change_data_capture) (CDC) tool. It collects mysql changes (insert, update, delete), runs a user supplied transformation logic on them and writes the result into kafka. This tool is inspired by go-mysql-elasticsearch and debezium.

## Status: WIP

## Features and Capabilities
 - [x] Supports Percona 8 and Mysql 8
 - [ ] Supports writing to multiple kafka clusters per pipeline
 - [x] Flexible transformation using a embedded javascript engine
 - [x] Low overhead and sub-millisecond processing per event on default settings
 - [x] Dynamic topic and key selection
 - [x] Selective drop/skip based on the event content
 - [ ] Run multiple pipelines per instance
 - [ ] Supports outbox pattern
 - [ ] Track the last state and ability to seek using GTID
 - [x] Simple to config and use
 - [ ] REST APIs to manage the runtime
 - [ ] Resilient to failures from both ends
 - [ ] Cloud ready


  Jump to **How to Use**

## Default Behaviour
*NB: this can change before the first stable release.*
Default behaviour of Prensio is summarizezd below. Please note that a big part these points are available to transform functions and can be overwritten on the fly (see Transformer):
- Sequentional processing per pipeline. Change events are processes in the order they arrive from the source database.
- Pipelines are independent processes.
- On irrecoverable errors, prensio will log the error and will exit
- Kafka topic: default value is `"{schema-name}[.]{table-name}"` for each event.
- Kafka topic will be created if it doesn't exist
- Kafka message key: 
-- If table has primary key, it will be a dot-separate list of all the columns that are part of the PK - `"{col1}[.]{col2}..."`.
-- If not, it will be same as default kafka topic value.
- Events are json encoded when produced to kafka
- Transformer scripts are checked on startup and before any event processing. Any failure will be logged and Prensio will exit abnormally. 

## Components
Components of this system are:
![Components](/docs/prensio-components.png?raw=true "Components")

### Pipeline
The high level construct to manage the flow of events from a source (mysql) to a sink (kafka). Pipeline(s) are created at startup using the supplied configuration files but you can use the REST APIs to build automation and dynamically manage the pipelines at runtime.
[more info TBD]

### MySQL Binlog Source
The connector for mysql which reads the mysql binlog events. like go-mysql-elasticsearch the canal package from go-mysql is used.
[more info TBD]

### Transformer
The scriptable transformer using goja package. Using this javascript sandbox, you will define a transform function which will receive the event and can manipulate it. A simple example:

*Caution: with flexibility comes ... hidden snakes. Refer to the "Usage Recommendations" section for some guidelines!*

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
the provided script should always have a function with this signature. `transform` function is the entry point. Of course you can have multiple other helper function beside that, as well.
```javascript
function transform(cdc_event)
```
**cdc_event** is the actual event to be processed which is passed by reference meaning any change will be reflected in the actual event. (See Event Schema below)

Your transform script can be as extensive as you want it to be, to the full extent of ECMA 5.1.

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
Prensio events have this schema and this is the structure for the **cdc_event passed to transform functions as well as the final kafka message**:
|  Field | Type | Description  |
| --- | --- | --- |
| `schema`  | string | name of the mysql schema the event coming from |
| `table` | string | table name for the mysql table the event coming from  |
| `action`  | string | one of `insert`, `update`, `delete`  (always lowecase) |
| `before`  | list of objects  | a list of old values for each affected row. every item is a object/map containing all columns and their value before the change |
| `after` | list of objects | a list of new values for each affected row. every item is a object/map containing all columns and their value after the change |
| `meta` | object | predefined map to carry extra data for each event. current items are `pipeline` (name of the prensio pipline producing the event) and `timstamp` (timestamp at which prensio started processing this event, in microseconds) |
| `kafka` | object | predefined map to control the destination topic and the key for the kafka message (see default behavior for more details). Exposed properties are `topic` and `key` |


### Kafka Sink
Kafka client to produce the events. This can produce to multiple topics.

### State Manager
The package to store and load the state of the pipline between restarts and as the stream of events flows. kafka is used as the backend.

## Usage Recommendations
Given the flexibility provided with Prensio, it is important to be careful with that degree of freedom.
- Consider keeping your transformation logic as simple as possible. It is very easy to get confused and shoot yourself in the foot with a big js code
- It is tempting to enrich the event on the spot using the transformer, be mindful about the trade offs between performance, reliability and maintainability
- Depending on your tolerance level, conside running one pipeline in each Prensio instance
- [more to come]

## Alternatives
### Debezium or Maxwell
If you are interested in a more mature solution with a sizeable community you should checkout debezium and maxwell at least.


### Why use Prensio
The biggest selling point of Prensio is its flexibility. It is also expected that it has a smaller resource footprint compared to solutions like debezium.
If js sandbox is not enough for your need, you can add custom helpers to fetch data from external sources to manipulate an event. In this context, the limit of what is achiveable with Prensio in regards to these requirements is the sky:
- on the spot event enrichment
- highly dynamic event distribution to different kafka clusters and topics

On some not-so-reliable tests, the total overhead of processing single change event was between 20 to 25 ms.

## How to Use
[TBD]

## How to Develop
Dev env is defined in the `docker-compose-devenv.yaml`. To use this and run the latest local code run this command:
```console
make devenv 
```
... and wait for the containers to be created. This will include a database and a kafka cluster of 3 brokers .

Then you can run your local code against that env using the command:
```console
make rundev
```
which will use the config file at path `sample_configs/prensio_dev_default.toml` by default.

To run tests, run this command:
```console
make test
```
this will spin up a test env similar to dev env but coming from `docker-compose-testenv.yaml`. Container ports are allocated dynamically and you need to have **"ginkgo"** testing tool installed.

You can then manipulate the tables and use any kafka client to monitor the results. Using kafkacat as an example:
```console
kafkacat -C -b 127.0.0.1:29092  -t test_mysql_ref_db_01.test_ref_table_01 -f "Topic=%t\nPart=%p\nOffset=%o\nheaders=%h\nKey=%k\nMsg=%s\n\n"
```