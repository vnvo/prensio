CREATE DATABASE to_be_ignored_db;
USE to_be_ignored_db;

CREATE TABLE should_ignore_table (
    id bigint
);

CREATE DATABASE test_mysql_ref_db_01;
USE test_mysql_ref_db_01;

CREATE TABLE test_ref_table_01 (
    # numeric data types
    `bit_col` BIT(3),
    `tinyint_col` TINYINT,
    `bool_col` BOOL,
    `smallint_col` SMALLINT,
    `mediumint_col` MEDIUMINT,
    `int_col` INT,
    `bigint_col` BIGINT,
    `decimal_col` DECIMAL( 10, 2 ),
    `float_col` FLOAT( 10, 2 ),
    `double_col` DOUBLE,

    #data and time data types
    `date_col` DATE,
    `datetime_col` DATETIME,
    `timestamp_col` TIMESTAMP,
    `time_col` TIME,
    `year_col` YEAR,

    #string data types
    `char_col` CHAR( 10 ),
    `varchar_col` VARCHAR( 20 ),
    `binary_col` BINARY( 64 ),
    `varbinary_col` VARBINARY( 20 ),
    `blob_col` BLOB,
    `tinyblob_col` TINYBLOB,
    `mediumblob_col` MEDIUMBLOB,
    `longblob_col` LONGBLOB,
    `text_col` TEXT,
    `tinytext_col` TINYTEXT,
    `mediumtext_col` MEDIUMTEXT,
    `longtext_col` LONGTEXT,    
    `enum_col` ENUM( '1', '2', '3' ),
    `set_col` SET( '1', '2', '3' ),

    #special data types
    `geometry_col` GEOMETRY,
    `point_col` POINT,
    

    `json_col` JSON
);