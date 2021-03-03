/* Copyright (c) 2021 Digital China Group Co.,Ltd
 * Licensed under the GNU General Public License, Version 3.0 (the "License").
 * You may not use this file except in compliance with the license.
 * You may obtain a copy of the license at
 *     https://www.gnu.org/licenses/
 *
 * This program is free; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; version 3.0 of the License.
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
**/

package dbbase

import (
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const (
	EMPTY   = ""
	ORDERBY = " ORDER BY "
	LIMIT   = " LIMIT "
	OFFSET  = " OFFSET "
)

// 存在两个数据库： 系统数据库 和 用户数据库
// 系统数据库为环境变量获取
// 用户数据库连接信息通过查询系统数据库获取

// 用户数据库连接
var UserDbContext DbContent

// 获取系统数据库连接
func newSystemDbConnection() (*sql.DB, error) {
	// 测试使用连接字符串
	connStr := os.Getenv("TiDB_Conn_Str")

	if connStr == "" {
		return nil, errors.New("tidb_conn_str not found")
	}

	dbConner, err := sql.Open("mysql", connStr)

	if err != nil {
		return nil, err
	}

	err = dbConner.Ping()
	if err != nil {
		return nil, err
	}

	return dbConner, nil
}

// getUserInfo 获取用户数据库信息
func getUserInfo(db *DbContent) error {
	conn, err := newSystemDbConnection()
	if err != nil {
		return err
	}

	defer conn.Close()

	environmentId := os.Getenv("Environment_Id")

	if environmentId == "" {
		return errors.New("environment id not found")
	}

	//测试使用环境Id
	sqlStr := "SELECT sql_host, sql_port, sql_user, sql_password, sql_dbname, sql_schema, metadata_published FROM environment WHERE environment_id='" + environmentId + "'"

	rows, err := conn.Query(sqlStr)
	if err != nil {
		return err
	}
	for rows.Next() {
		if err := rows.Scan(&db.Host, &db.Port, &db.UserName, &db.Password, &db.DbName, &db.SchemaName, &db.DbModeling); err != nil {
			return err
		}
	}

	return nil
}

// getUserDbSchema 通过Modeling 解析数据库架构，获取表名和字段名
func getDbSchema(db *DbContent) error {
	err := xml.Unmarshal([]byte(db.DbModeling), &db.DbEntity)
	if err != nil {
		return err
	}

	if len(db.DbEntity.Entities) <= 0 || db.DbEntity.Entities == nil {
		return errors.New("table does not exist in the database")
	}

	// 所有的表
	allTablesOfSchema := make([]string, 0)

	// 每个表中所有的字段
	allFieldsOfTable := make(map[string][]Column, 0)

	for _, entityValue := range db.DbEntity.Entities {
		allTablesOfSchema = append(allTablesOfSchema, entityValue.SchemaNameAttr)
		columns := make([]Column, 0)
		if entityValue.Fields == nil || len(entityValue.Fields) <= 0 {
			allFieldsOfTable[entityValue.SchemaNameAttr] = columns
			continue
		}

		for _, fieldValue := range entityValue.Fields {
			columns = append(columns, Column{
				FieldName: fieldValue.SchemaNameAttr,
				DataType:  fieldValue.DataTypeAttr,
				DataSize:  fieldValue.TypeOption.LengthAttr,
			})
		}

		allFieldsOfTable[entityValue.SchemaNameAttr] = columns
	}

	db.AllTablesOfSchema = allTablesOfSchema
	db.AllFieldsOfTable = allFieldsOfTable

	return nil
}

// NewDbConnection ...
// 获取一个新的数据库连接 Open状
// 获取的为用户数据库连接
func (db *DbContent) NewDbConnection() error {
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8",
		db.UserName, db.Password, db.Host, strconv.Itoa(db.Port), db.DbName)

	var err error
	db.DbConn, err = sql.Open("mysql", connStr)

	if err != nil {
		return err
	}

	err = db.DbConn.Ping()
	if err != nil {
		return err
	}

	return nil
}

// Refresh ...
// 刷新数据库信息
func (db *DbContent) Refresh() error {

	var err error
	err = getUserInfo(db)
	if err != nil {
		return err
	}

	if err = db.NewDbConnection(); err != nil {
		return err
	}

	if err = getDbSchema(db); err != nil {
		return err
	}

	/*	if err = getAllTablesOfSchema(db); err != nil {
			return err
		}
		// 此处可能为性能瓶颈 待优化
		if err = getAllFieldsOfTable(db); err != nil {
			return err
		}*/

	db.LastRefreshTime = time.Now()
	return nil
}

// GenerateQuerySQL ...
// 生成查询语句 单表查询
// 记得对入参进行合法性检查
// 目前只支持全部为and连接的查询过滤语句
func (db *DbContent) GenerateQuerySQL(tableName string, queryFields []string, filterFields []FilterField, condition map[string]string) (*DbSQL, error) {
	// 核心问题：查询哪个表的哪些字段，过滤条件添加
	// 检查放在上层
	// 检查：表是否存在，字段是否都属于这个表
	if queryFields == nil || len(queryFields) < 1 {
		return nil, fmt.Errorf("the len of queryFields must more than 0")
	}

	var sqlSelect string

	// 查询语句生成 + 过滤条件 + 参数列表
	// 查询主体语句
	sqlQueryField := strings.Join(queryFields, ",")

	if filterFields != nil && len(filterFields) > 0 {
		// 查看字段名称是否存在于表内
		filterFieldNames := make([]string, 0, len(filterFields))
		for _, value := range filterFields {
			filterFieldNames = append(filterFieldNames, value.Field.BelongColumn.FieldName)
		}

		// 检查放在上层
		/*isFilterFieldsValid, err := db.Valid(TableName, filterFieldNames)
		if !isFilterFieldsValid {
			return nil, err
		}*/

		// 有过滤条件,这里需要重构。odata需要支持and , or ,分组等连接符。
		//SQL注入防范未成功
		filterFieldsStr, err := ToFiltersStr(filterFields)
		if err != nil {
			return nil, err
		}

		sqlFilter := strings.Join(filterFieldsStr, " and ") // 目前只支持全部为and连接的查询过滤语句

		sqlSelect = fmt.Sprintf("SELECT %s FROM %s WHERE %s", sqlQueryField, tableName, sqlFilter)
	} else {
		sqlSelect = fmt.Sprintf("SELECT %s FROM %s", sqlQueryField, tableName)
	}
	//todo 补上判断字段是否在表内
	//query := url.Values{}
	//query["$filter"] = []string{condition["filter"]}
	//filter, _ := ODataSQLFilter(query)
	//if filter != "" {
	//	sqlSelect = fmt.Sprintf("SELECT %s FROM %s WHERE %s", sqlQueryField, tableName, filter)
	//} else {
	//	sqlSelect = fmt.Sprintf("SELECT %s FROM %s", sqlQueryField, tableName)
	//}

	// 检查是否有条件限制，如果有则拼接进字符串
	if condition["orderBy"] != EMPTY {
		sqlSelect = sqlSelect + ORDERBY + condition["orderBy"]
	}

	if condition["limit"] != EMPTY {
		sqlSelect += LIMIT + condition["limit"]
	}

	if condition["offset"] != EMPTY {
		sqlSelect += OFFSET + condition["offset"]
	}
	fmt.Println(sqlSelect)
	// 构造一个SQL执行体
	dbSQL := new(DbSQL)
	dbSQL.SqlStr = sqlSelect
	dbSQL.TableName = tableName
	return dbSQL, nil
}

func (db *DbContent) GenerateOdataQuerySQL(tableName string, queryFields []string, filterFields []FilterField, condition map[string]string) (*DbSQL, error) {
	// 核心问题：查询哪个表的哪些字段，过滤条件添加
	// 检查放在上层
	// 检查：表是否存在，字段是否都属于这个表
	if queryFields == nil || len(queryFields) < 1 {
		return nil, fmt.Errorf("the len of queryFields must more than 0")
	}

	var sqlSelect string

	// 查询语句生成 + 过滤条件 + 参数列表
	// 查询主体语句
	sqlQueryField := strings.Join(queryFields, ",")

	//todo 补上判断字段是否在表内
	query := url.Values{}
	query["$filter"] = []string{condition["filter"]}
	filter, _ := ODataSQLFilter(query)
	if filter != "" {
		sqlSelect = fmt.Sprintf("SELECT %s FROM %s WHERE %s", sqlQueryField, tableName, filter)
	} else {
		sqlSelect = fmt.Sprintf("SELECT %s FROM %s", sqlQueryField, tableName)
	}

	// 检查是否有条件限制，如果有则拼接进字符串
	if condition["orderBy"] != "" {
		sqlSelect = sqlSelect + ORDERBY + condition["orderBy"]
	}

	if condition["limit"] != EMPTY {
		sqlSelect += LIMIT + condition["limit"]
	}

	if condition["offset"] != EMPTY {
		sqlSelect += OFFSET + condition["offset"]
	}

	// 构造一个SQL执行体
	dbSQL := new(DbSQL)
	dbSQL.SqlStr = sqlSelect
	dbSQL.TableName = tableName
	return dbSQL, nil
}

func (db *DbContent) GenerateCountSQL(tableName string) (*DbSQL, error) {
	sqlStr := "SELECT COUNT(*) FROM " + tableName
	dbSQL := new(DbSQL)
	dbSQL.SqlStr = sqlStr
	dbSQL.TableName = tableName
	return dbSQL, nil
}

// GenerateInsertSQL 生成插入语句
func (db *DbContent) GenerateInsertSQL(tableName string, insertFields []Row) (*DbSQL, error) {
	if insertFields == nil || len(insertFields) < 1 {
		return nil, fmt.Errorf("the len of insertField must more than 0")
	}

	//多行数据
	insertRow := make([]string, 0)

	// 插入语句的字段名，只有一组，所以所有的值都应该是同一组字段名
	fieldName := make([]string, 0)

	for index, _ := range insertFields[0].FieldCells {
		fieldName = append(fieldName, "`"+insertFields[0].FieldCells[index].BelongColumn.FieldName+"`")
	}

	// 生成Sql语句中的字段名
	sqlInsertField := strings.Join(fieldName, ", ")

	for _, row := range insertFields { // 获取插入字段的字段名和值
		fieldValue := make([]string, 0)
		for index, _ := range row.FieldCells {
			if row.FieldCells[index].BelongColumn.DataType == "integer" {
				fieldValue = append(fieldValue, "'"+row.FieldCells[index].ToFieldStr()+"'")
			} else {
				fieldValue = append(fieldValue, row.FieldCells[index].ToFieldStr())
			}
		}
		sqlInsertValue := strings.Join(fieldValue, ", ")
		sqlInsertValue = "(" + sqlInsertValue + ")"
		insertRow = append(insertRow, sqlInsertValue)
	}

	sqlInsertRows := strings.Join(insertRow, ", ")

	// 检查放在上层中，避免多次检验
	// 检查：表是否存在，字段是否都属于这个表
	/*	isValid, err := db.Valid(TableName, FieldName)
		if !isValid {
			return nil, err
		}*/

	sqlInsertStr := fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s`, tableName, sqlInsertField, sqlInsertRows)

	// 构造一个SQL执行体
	dbSQL := new(DbSQL)
	dbSQL.SqlStr = sqlInsertStr
	dbSQL.TableName = tableName
	return dbSQL, nil
}

// GenerateUpdateSQL 生成更新语句
func (db *DbContent) GenerateUpdateSQL(tableName string, updateField []FieldCell, filterFields []FilterField) (*DbSQL, error) {
	if updateField == nil || len(updateField) < 1 {
		return nil, fmt.Errorf("the len of updateField must more than 0")
	}

	fieldName := make([]string, 0)

	// 将更新字段变为Filter类型, 通过toFiltersStr方法变成 FieldName = fieldValue 字符串
	updateFields := &[]FilterField{}

	for _, value := range updateField {
		*updateFields = append(*updateFields, FilterField{
			CompareOption: "=",
			Field:         value,
		})
		fieldName = append(fieldName, value.BelongColumn.FieldName)
	}

	// 检查：表是否存在，字段是否都属于这个表
	// 放到上层检验
	/*	isValid, err := db.Valid(TableName, FieldName)
		if !isValid {
			return nil, err
		}
	*/
	FieldStr, err := ToFiltersStr(*updateFields)
	if err != nil {
		return nil, err
	}

	sqlField := strings.Join(FieldStr, ", ")

	var sqlUpdateStr string

	if filterFields != nil && len(filterFields) > 0 {
		// 查看字段名称是否存在于表内
		filterFieldNames := make([]string, 0, len(filterFields))
		for _, value := range filterFields {
			filterFieldNames = append(filterFieldNames, value.Field.BelongColumn.FieldName)
		}

		// 字段检查放在上层
		/*		isFilterFieldsValid, err := db.Valid(TableName, filterFieldNames)
				if !isFilterFieldsValid {
					return nil, err
				}*/

		FilterStr, err := ToFiltersStr(filterFields)
		if err != nil {
			return nil, err
		}

		sqlFilter := strings.Join(FilterStr, " and ")

		sqlUpdateStr = fmt.Sprintf("UPDATE %s SET %s WHERE %s", tableName, sqlField, sqlFilter)
	} else {
		sqlUpdateStr = fmt.Sprintf("UPDATE %s SET %s", tableName, sqlField)
	}

	// 构造一个SQL执行体
	dbSQL := new(DbSQL)
	dbSQL.SqlStr = sqlUpdateStr
	dbSQL.TableName = tableName
	return dbSQL, nil
}

// GenerateDeleteSQL 生成删除语句
func (db *DbContent) GenerateDeleteSQL(tableName string, filterFields []FilterField) (*DbSQL, error) {

	// 检测全部放到上层
	// 检查：表是否存在，字段是否都属于这个表
	/*	isValid, err := db.Valid(TableName, []string{})
		if !isValid {
			return nil, err
		}*/
	var sqlDeleteStr string

	if filterFields != nil && len(filterFields) > 0 {
		// 查看字段名称是否存在于表内
		filterFieldNames := make([]string, 0, len(filterFields))
		for _, value := range filterFields {
			filterFieldNames = append(filterFieldNames, value.Field.BelongColumn.FieldName)
		}

		// 检测放到上层
		/*isFilterFieldsValid, err := db.Valid(TableName, filterFieldNames)
		if !isFilterFieldsValid {
			return nil, err
		}*/

		FilterStr, err := ToFiltersStr(filterFields)
		if err != nil {
			return nil, err
		}

		sqlFilter := strings.Join(FilterStr, " and ")

		sqlDeleteStr = fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, sqlFilter)
	} else {
		sqlDeleteStr = fmt.Sprintf("DELETE FROM %s", tableName)
	}

	// 构造一个SQL执行体
	dbSQL := new(DbSQL)
	dbSQL.SqlStr = sqlDeleteStr
	dbSQL.TableName = tableName
	return dbSQL, nil
}

// RunQuery 执行sql 查询 相关语句 仅仅只支持单表查询
func (db *DbContent) RunQuery(dbSql *DbSQL) error {
	table := DbTable{
		Rows: []Row{},
		Cols: []Column{},
	}

	respond := DbRespond{
		RespondStatus: false,
		RespondData:   table,
	}

	var err error

	rows, err := UserDbContext.DbConn.Query(dbSql.SqlStr)
	if err != nil {
		return err
	}

	defer rows.Close()

	// 获取查询到的字段名
	fields, err := rows.Columns()
	if err != nil {
		return err
	}

	// 在表中的表头查询相应的字段，获取需要
	column := &[]Column{}
	for _, value := range db.AllFieldsOfTable[dbSql.TableName] {
		for _, field := range fields {
			if value.FieldName == field {
				*column = append(*column, value)
			}
		}
	}
	//count查询的结果返回字段名不在字段之中，所以需要额外判断一些，有没有count字段
	for _, field := range fields {
		if field == "COUNT(*)" {
			*column = append(*column, Column{
				FieldName: "COUNT(*)",
				DataType:  "integer",
				DataSize:  4,
			})
		}
	}

	// 根据表头中的数据获取数据
	for rows.Next() {
		// 记录一行的数据
		row := Row{
			FieldCells: []FieldCell{},
		}

		// 读取一行中的所有数据
		rawResult := make([][]uint8, len(*column))
		dest := make([]interface{}, len(*column))

		for i, _ := range dest {
			dest[i] = &rawResult[i]

		}

		if err := rows.Scan(dest...); err != nil {
			return err
		}

		// 读取一行中的每一个字段
		for index, _ := range *column {
			var fieldValue interface{}
			switch (*column)[index].DataType {
			case "integer":
				fieldValue, _ = strconv.Atoi(string(rawResult[index]))
			case "decimal", "money":
				fieldValue, _ = strconv.ParseFloat(string(rawResult[index]), 32)
			case "boolean":
				fieldValue, _ = strconv.ParseBool(strconv.Itoa(int(rawResult[index][0])))
			default:
				fieldValue = string(rawResult[index])
			}
			row.FieldCells = append(row.FieldCells, FieldCell{
				OriginData:   fieldValue,
				BelongColumn: &(*column)[index],
			})
		}

		respond.RespondData.Rows = append(respond.RespondData.Rows, row)
	}

	respond.RespondData.Cols = *column
	respond.RespondStatus = true
	respond.Err = errors.New("OK")
	dbSql.Respond = &respond
	return nil
}

// RunExec 执行sql 删除 修改 和 插入语句 仅仅只支持单表查询
func (db *DbContent) RunExec(sql *DbSQL) error {
	respond := DbRespond{
		RespondStatus: false,
	}

	var err error

	// 结果不保留，后期如果需求需要可以将结果返回
	_, err = UserDbContext.DbConn.Exec(sql.SqlStr)
	if err != nil {
		return err
	}

	respond.RespondStatus = true
	respond.Err = errors.New("OK")
	sql.Respond = &respond
	return nil
}

// Valid ...
// 验证此表是否属于此数据库，此字段是否属于此表
func (db *DbContent) Valid(tableName *string, field []string) (bool, error) {
	// 判断这个表是否在db.AllTablesOfSchema 这个集合中
	index := -1
	for i, item := range db.AllTablesOfSchema {
		if strings.EqualFold(item, *tableName) {
			*tableName = item
			index = i
			break
		}
	}

	if index < 0 {
		return false, fmt.Errorf("db.AllTablesOfSchema don't contains '%s'", *tableName)
	}

	// 判断字段是否属于这个表
	isContain := false
	for i := range field {
		for _, item := range db.AllFieldsOfTable[*tableName] {
			if isContain = strings.EqualFold(field[i], item.FieldName); isContain {
				break
			}
		}
		if !isContain {
			return false, fmt.Errorf("db.AllFieldsOfTable[%s] don't contains '%s'", *tableName, field[i])
		}
		isContain = false
	}

	return true, nil
}

// ToFilterStr ...
// 将过滤字段转换未sql中的过滤短语 例如 name = 'Tim'、age = 18
// TODO: 注入式攻击检验
func (ff *FilterField) ToFilterStr() (string, error) {

	switch ff.Field.BelongColumn.DataType {
	case "datetime", "date", "time", "timestamp", "timestamptz":
		switch ff.Field.OriginData.(type) {
		case time.Time:
			t, _ := ff.Field.OriginData.(time.Time)
			return fmt.Sprintf("%s %s '%s'", ff.Field.BelongColumn.FieldName, ff.CompareOption, t.Format("2006-01-02 15:04:05")), nil
		case string:
			return fmt.Sprintf("%s %s '%v'", ff.Field.BelongColumn.FieldName, ff.CompareOption, ff.Field.OriginData), nil
		default:
			return "", fmt.Errorf("ToFilterStr error,Unimplemented type filter field convert datatime FieldName:'%s',CompareOption:'%s',OriginData:'%v'", ff.Field.BelongColumn.FieldName, ff.CompareOption, ff.Field.OriginData)
		}
	case "varchar", "text", "string":
		return fmt.Sprintf("%s %s '%v'", ff.Field.BelongColumn.FieldName, ff.CompareOption, ff.Field.OriginData), nil
	case "integer", "decimal", "int2", "int4", "int8", "float4", "float8", "boolean", "bool", "money":
		return fmt.Sprintf("`%s` %s %v", ff.Field.BelongColumn.FieldName, ff.CompareOption, ff.Field.OriginData), nil
	default:
		return "", fmt.Errorf("Unimplemented filter field convert，FieldName:'%s',CompareOption:'%s',OriginData:'%v'", ff.Field.BelongColumn.FieldName, ff.CompareOption, ff.Field.OriginData)
	}
}

// ToFieldStr ...
// 将不同类型的值转为相应的字符串类型
func (field *FieldCell) ToFieldStr() string {
	switch field.BelongColumn.DataType {
	case "integer", "decimal", "money", "float4", "float8", "boolean", "bool":
		return fmt.Sprintf(`%v`, field.OriginData)
	//case "datetime", "date", "time", "timestamp", "timestamptz":
	//	switch field.OriginData.(type) {
	//	case time.Time:
	//		t, _ := field.OriginData.(time.Time)
	//		return fmt.Sprintf(`'%v'`, t.Format("2006-01-02 15:04:05"))
	//	default:
	//		return fmt.Sprintf(`'%v'`, field.OriginData)
	//	}
	default:
		return fmt.Sprintf(`'%v'`, field.OriginData)
	}
}

// ToFiltersStr ...
// 将一组过滤字段批量进行转换成过滤短语 使用切片进行返回
func ToFiltersStr(ffs []FilterField) ([]string, error) {
	length := len(ffs)
	resFiltersStr := make([]string, 0, length)
	var filterStr string
	var err error
	for i := range ffs {
		if filterStr, err = ffs[i].ToFilterStr(); err != nil {
			return nil, err
		}
		resFiltersStr = append(resFiltersStr, filterStr)
	}
	return resFiltersStr, nil
}

/*Json 格式
{
	"RespondStatus": true,
	"Err": "OK",
	"RespondData": [
	{
		"FieldName": "fieldValue",
		"fieldNameA": "fieldValueA"
	},{
		"FieldName": "fieldValue"
		"fieldNameA": "fieldValueA"
	}]
}
*/

// TableToJSONStr 将自定义的相应结构体转为Map
func DbRespondToMap(res DbRespond) (map[string]interface{}, error) {
	jsonMap := map[string]interface{}{
		"RespondStatus": res.RespondStatus,
	}

	if res.Err != nil {
		jsonMap["Err"] = res.Err.Error()
	} else {
		jsonMap["Err"] = "OK"
	}

	var err error
	jsonMap["RespondData"], err = GetDataArr(res)
	if err != nil {
		return nil, err
	}

	return jsonMap, nil
}

// 将自定义数据存储结构转换为数组数据
func GetDataArr(res DbRespond) ([]map[string]interface{}, error) {
	data := make([]map[string]interface{}, 0)

	for _, rowValue := range res.RespondData.Rows {
		if len(rowValue.FieldCells) > len(res.RespondData.Cols) {
			return nil, errors.New("row length more than column length")
		}
		rowData := map[string]interface{}{}
		for colIndex, colValue := range res.RespondData.Cols {
			rowData[colValue.FieldName] = rowValue.FieldCells[colIndex].OriginData
		}
		data = append(data, rowData)
	}

	return data, nil
}
