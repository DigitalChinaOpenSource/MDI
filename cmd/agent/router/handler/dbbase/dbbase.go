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
	"dataapi/internal/kernel/metadata/modeling"
	"database/sql"
	"time"
)

/*
 * ============= 基础结构信息 ==========
**/

// ColumnType ...
type ColumnType string

// Column ...
// 一个列头信息
type Column struct {
	FieldName string
	DataType  string
	DataSize  int
}

// FieldCell ...
// 每个字段的值
type FieldCell struct {
	OriginData   interface{}
	BelongColumn *Column
}

// Row ...
// 一行数据
type Row struct {
	FieldCells []FieldCell
}

// DbTable ...
// 多行数据
type DbTable struct {
	Rows []Row
	Cols []Column
}

// FilterField ...
// 一个过滤条件
type FilterField struct {
	Field         FieldCell
	CompareOption string
}

/*
 * ================ 定义数据库基本信息 ===================
**/

// DbContent ...
// 数据库基本信息
type DbContent struct {
	UserName          string
	Host              string
	Port              int
	DbName            string
	Password          string
	SchemaName        string
	DbConn            *sql.DB
	DbModeling        string
	DbEntity          modeling.Model
	AllTablesOfSchema []string            // tables name
	AllFieldsOfTable  map[string][]Column // key：Table name, value: fields of table
	LastRefreshTime   time.Time
}

// DbConnectioner ...
// 获取一个新的数据库连接 Open状态
type DbConnectioner interface {
	NewDbConnection() error
}

// DbContentRefresher ...
// 刷新DbContent的缓存内容
type DbContentRefresher interface {
	Refresh() error
}

// DbValider ...
// 判断数据库是否含有这些内容
type DbValider interface {
	Valid(tableName string, field []string) (bool, error)
}

// DbSQLGenerate ...
type DbSQLGenerater interface {
	GenerateQuerySQL(tableName string, queryFields []string, filterFields []FilterField, condition map[string]string) (*DbSQL, error)
	GenerateInsertSQL(tableName string, insertField []FieldCell) (*DbSQL, error)
	GenerateUpdateSQL(tableName string, updateField []FieldCell, filterFields []FilterField) (*DbSQL, error)
	GenerateDeleteSQL(tableName string, filterFields []FilterField) (*DbSQL, error)
}

/*
 *  =============== 定义每个SQL请求信息 ===================
**/

// DbSQL ...
// 一个要执行的SQL 包含所有必要的信息 方便并发执行
type DbSQL struct {
	SqlStr    string
	TableName string
	SqlParams map[string]interface{}
	Respond   *DbRespond
}

// SQLRunner ...
// 直接执行SQL（根据SQL文本，判断应该执行CURD哪个）。将执行结果写入 DbSQL.ChsRespon 中
type SQLRunner interface {
	RunQuery(sql *DbSQL) error
	RunExec(sql *DbSQL) error
}

/*
 * =============== 定义SQL请求响应信息 ====================
 *
**/

// DbRespond ...
type DbRespond struct {
	RespondStatus bool
	RespondData   DbTable
	Err           error
}

// DbResponder ...
type DbResponder interface {
	GetCellDataByIndex(rowIndex uint32, colIndex uint32) (FieldCell, error)
	GetColumnDataByIndex(colIndex uint32) ([]FieldCell, error)
	GetColumnDataByFieldName(fieldName string) ([]FieldCell, error)
	DbRespondToMap(res DbRespond) (map[string]interface{}, error)
}
