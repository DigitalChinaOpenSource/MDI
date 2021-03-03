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

package ti

import (
	"dataapi/cmd/agent/router/handler/dbbase"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type DbContent struct {
	dbbase.DbContent
}

var UserDbContent DbContent

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

// GenerateQuerySQL ...
// 生成查询语句 单表查询
// 记得对入参进行合法性检查
// 目前只支持全部为and连接的查询过滤语句
func (db *DbContent) GenerateQuerySQL(tableName string, queryFields []string, filterFields []dbbase.FilterField, condition map[string]string) (*dbbase.DbSQL, error) {
	// 核心问题：查询哪个表的哪些字段，过滤条件添加

	// 检查放在上层
	// 检查：表是否存在，字段是否都属于这个表
	/*	isValid, err := db.Valid(TableName, queryFields)
		if !isValid {
			return nil, err
		}*/

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

		// 有过滤条件
		filterFieldsStr, err := dbbase.ToFiltersStr(filterFields)
		if err != nil {
			return nil, err
		}

		sqlFilter := strings.Join(filterFieldsStr, " and ") // 目前只支持全部为and连接的查询过滤语句

		sqlSelect = fmt.Sprintf("SELECT %s FROM %s WHERE %s", sqlQueryField, tableName, sqlFilter)
	} else {
		sqlSelect = fmt.Sprintf("SELECT %s FROM %s", sqlQueryField, tableName)
	}

	// 检查是否有条件限制，如果有则拼接进字符串
	if condition["orderBy"] != "" {
		sqlSelect = sqlSelect + " order by " + condition["orderBy"]
	}

	if condition["limit"] != "" {
		sqlSelect = sqlSelect + " limit " + condition["limit"]
	}

	if condition["offset"] != "" {
		sqlSelect = sqlSelect + " offset " + condition["offset"]
	}

	// 构造一个SQL执行体
	dbSQL := new(dbbase.DbSQL)
	dbSQL.SqlStr = sqlSelect
	dbSQL.TableName = tableName
	return dbSQL, nil
}

// GenerateInsertSQL 生成插入语句
func (db *DbContent) GenerateInsertSQL(tableName string, insertField []dbbase.FieldCell) (*dbbase.DbSQL, error) {
	if insertField == nil || len(insertField) < 1 {
		return nil, fmt.Errorf("the len of insertField must more than 0")
	}

	// 获取插入字段的字段名
	fieldName := make([]string, 0)
	fieldValue := make([]string, 0)

	for _, value := range insertField {
		fieldName = append(fieldName, value.BelongColumn.FieldName)
		fieldValue = append(fieldValue, value.ToFieldStr())
	}

	// 生成Sql语句
	sqlInsertField := strings.Join(fieldName, ", ")
	sqlInsertValue := strings.Join(fieldValue, ", ")

	sqlInsertStr := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, tableName, sqlInsertField, sqlInsertValue)

	// 构造一个SQL执行体
	dbSQL := new(dbbase.DbSQL)
	dbSQL.SqlStr = sqlInsertStr
	dbSQL.TableName = tableName
	return dbSQL, nil
}

// GenerateUpdateSQL 生成更新语句
func (db *DbContent) GenerateUpdateSQL(tableName string, updateField []dbbase.FieldCell, filterFields []dbbase.FilterField) (*dbbase.DbSQL, error) {
	if updateField == nil || len(updateField) < 1 {
		return nil, fmt.Errorf("the len of updateField must more than 0")
	}

	fieldName := make([]string, 0)

	// 将更新字段变为Filter类型, 通过toFiltersStr方法变成 FieldName = fieldValue 字符串
	updateFields := &[]dbbase.FilterField{}

	for _, value := range updateField {
		*updateFields = append(*updateFields, dbbase.FilterField{
			CompareOption: "=",
			Field:         value,
		})
		fieldName = append(fieldName, value.BelongColumn.FieldName)
	}

	FieldStr, err := dbbase.ToFiltersStr(*updateFields)
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

		FilterStr, err := dbbase.ToFiltersStr(filterFields)
		if err != nil {
			return nil, err
		}

		sqlFilter := strings.Join(FilterStr, " and ")

		sqlUpdateStr = fmt.Sprintf("UPDATE %s SET %s WHERE %s", tableName, sqlField, sqlFilter)
	} else {
		sqlUpdateStr = fmt.Sprintf("UPDATE %s SET %s", tableName, sqlField)
	}

	// 构造一个SQL执行体
	dbSQL := new(dbbase.DbSQL)
	dbSQL.SqlStr = sqlUpdateStr
	dbSQL.TableName = tableName
	return dbSQL, nil
}

// GenerateDeleteSQL 生成删除语句
func (db *DbContent) GenerateDeleteSQL(tableName string, filterFields []dbbase.FilterField) (*dbbase.DbSQL, error) {

	var sqlDeleteStr string

	if filterFields != nil && len(filterFields) > 0 {
		// 查看字段名称是否存在于表内
		filterFieldNames := make([]string, 0, len(filterFields))
		for _, value := range filterFields {
			filterFieldNames = append(filterFieldNames, value.Field.BelongColumn.FieldName)
		}

		FilterStr, err := dbbase.ToFiltersStr(filterFields)
		if err != nil {
			return nil, err
		}

		sqlFilter := strings.Join(FilterStr, " and ")

		sqlDeleteStr = fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, sqlFilter)
	} else {
		sqlDeleteStr = fmt.Sprintf("DELETE FROM %s", tableName)
	}

	// 构造一个SQL执行体
	dbSQL := new(dbbase.DbSQL)
	dbSQL.SqlStr = sqlDeleteStr
	dbSQL.TableName = tableName
	return dbSQL, nil
}

// RunQuery 执行sql 查询 相关语句 仅仅只支持单表查询
func (db *DbContent) RunQuery(dbSql *dbbase.DbSQL) error {
	table := dbbase.DbTable{
		Rows: []dbbase.Row{},
		Cols: []dbbase.Column{},
	}

	respond := dbbase.DbRespond{
		RespondStatus: false,
		RespondData:   table,
	}

	var err error

	rows, err := dbbase.UserDbContext.DbConn.Query(dbSql.SqlStr)
	if err != nil {
		return err
	}

	defer rows.Close()

	// 获取查询到的字段名
	fields, err := rows.Columns()
	if err != nil {
		return err
	}

	// 在表中查询相应的字段，获取需要的表头
	column := &[]dbbase.Column{}
	for _, value := range db.AllFieldsOfTable[dbSql.TableName] {
		for _, field := range fields {
			if value.FieldName == field {
				*column = append(*column, value)
			}
		}
	}

	// 根据表头中的数据获取数据
	for rows.Next() {
		// 记录一行的数据
		row := dbbase.Row{
			FieldCells: []dbbase.FieldCell{},
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
				fieldValue, _ = strconv.ParseBool(string(rawResult[index]))
			default:
				fieldValue = string(rawResult[index])
			}
			row.FieldCells = append(row.FieldCells, dbbase.FieldCell{
				//OriginData:   rawResult[index],
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
func (db *DbContent) RunExec(sql *dbbase.DbSQL) error {
	respond := dbbase.DbRespond{
		RespondStatus: false,
	}

	var err error

	// 结果不保留，后期如果需求需要可以将结果返回
	_, err = db.DbConn.Exec(sql.SqlStr)
	if err != nil {
		return err
	}

	respond.RespondStatus = true
	respond.Err = errors.New("OK")
	sql.Respond = &respond
	return nil
}