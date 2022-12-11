DELIMITER $$
DROP PROCEDURE IF EXISTS create_mysql_ref_table;

CREATE PROCEDURE create_mysql_ref_table(IN num bigint)
BEGIN

SET @createtable = CONCAT ('CREATE TABEL mysql_ref_table_',  num);
SET @tablecols = '
    `varchar_col` VARCHAR( 20 )  ,
    `tinyint_col` TINYINT  ,
    `text_col` TEXT  ,
    `date_col` DATE  ,
    `smallint_col` SMALLINT  ,
    `mediumint_col` MEDIUMINT  ,
    `int_col` INT  ,
    `bigint_col` BIGINT  ,
    `float_col` FLOAT( 10, 2 )  ,
    `double_col` DOUBLE  ,
    `decimal_col` DECIMAL( 10, 2 )  ,
    `datetime_col` DATETIME  ,
    `timestamp_col` TIMESTAMP  ,
    `time_col` TIME  ,
    `year_col` YEAR  ,
    `char_col` CHAR( 10 )  ,
    `tinyblob_col` TINYBLOB  ,
    `tinytext_col` TINYTEXT  ,
    `blob_col` BLOB  ,
    `mediumblob_col` MEDIUMBLOB  ,
    `mediumtext_col` MEDIUMTEXT  ,
    `longblob_col` LONGBLOB  ,
    `longtext_col` LONGTEXT  ,
    `json_col` JSON,
    `enum_col` ENUM( '1', '2', '3' )  ,
    `set_col` SET( '1', '2', '3' )  ,
    `bool_col` BOOL  ,
    `binary_col` BINARY( 64 )  ,
    `varbinary_col` VARBINARY( 20 )
'

