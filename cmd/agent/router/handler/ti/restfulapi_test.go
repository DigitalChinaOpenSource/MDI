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
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var dbContent *dbbase.DbContent = new(dbbase.DbContent)
var tableName string = "events"
var queryFields []string = []string{"node_id", "event_timestamp"}
var filterFields []dbbase.FilterField
var col = dbbase.Column{FieldName: "node_id", DataType: "integer", DataSize: 32}
var col1 = dbbase.Column{FieldName: "event_timestamp", DataType: "timestamptz", DataSize: -1}

func TestSelectSQL(t *testing.T) {

	filterFields = make([]dbbase.FilterField, 0, 8)
	filterFields = append(filterFields, dbbase.FilterField{
		Field:         dbbase.FieldCell{OriginData: 1001, BelongColumn: &col},
		CompareOption: "=",
	})
	filterFields = append(filterFields, dbbase.FilterField{
		//查询语句中传入 添加sql注入的字符串
		//Field: dbbase.FieldCell{OriginData: "' or 1=1  or event_timestamp > '2020-09-24 02:16:16", BelongColumn: &col1},
		Field:         dbbase.FieldCell{OriginData: "2020-09-24 02:16:16", BelongColumn: &col1},
		CompareOption: ">",
	})
	condition := map[string]string{
		"limit":   "2",
		"offset":  "1",
		"orderBy": "name desc",
	}
	//测试查询方法
	dbSQL, err := dbContent.GenerateQuerySQL(tableName, queryFields, filterFields, condition)
	//fmt.Println(dbSQL.SqlStr)
	//断言

	if err != nil {
		t.Fatal(err)
	} else {

		sqlFilter := "`node_id` = 1001 and event_timestamp > '2020-09-24 02:16:16'"

		sqlField := strings.Join(queryFields, ",")

		conDate := " ORDER BY " + condition["orderBy"] + " LIMIT " + condition["limit"] + " OFFSET " + condition["offset"]

		//应该返回的正确的sql语句   期望值
		sqlSelect := fmt.Sprintf("SELECT %s FROM %s WHERE %s%s", sqlField, tableName, sqlFilter, conDate)

		//fmt.Println(sqlSelect)
		t.Log(dbSQL.SqlStr)

		//把期望值和实际值都变成数组
		mockArrays := strings.Fields(sqlSelect) //默认以空格切割返回[]数组
		sqlArrays := strings.Fields(dbSQL.SqlStr)

		//对比字段名
		sqlQueryField := strings.Join(queryFields, ",")

		fmt.Println(assert.Equal(t, sqlQueryField, sqlArrays[1]))
		//对比表名

		fmt.Println(assert.Equal(t, tableName, sqlArrays[3]))
		//判断是否有关键字
		conOrder := "ORDER"
		conOrder2 := "BY"
		conLimit := "LIMIT"
		conOffSet := "OFFSET"
		fmt.Println(assert.Contains(t, sqlArrays, conOrder, conOrder2, conLimit, conOffSet))

		//全部变成切片排序后对比
		sort.Strings(mockArrays)
		sort.Strings(sqlArrays)
		assert.Equal(t, mockArrays, sqlArrays)

	}

}

func TestInsertSQL(t *testing.T) {

	insertFieldCells := &[]dbbase.FieldCell{}

	*insertFieldCells = append(*insertFieldCells, dbbase.FieldCell{
		//OriginData: "11','1222",
		OriginData:   1222,
		BelongColumn: &col,
	})
	*insertFieldCells = append(*insertFieldCells, dbbase.FieldCell{
		OriginData:   "2020-09-24 02:16:16",
		BelongColumn: &col1,
	})
	insertFields := []dbbase.Row{
		{
			FieldCells: *insertFieldCells,
		},
	}
	//测试生成插入语句
	dbSQL, err := dbContent.GenerateInsertSQL(tableName, insertFields)
	fmt.Println(dbSQL.SqlStr)
	//断言
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	} else {
		t.Log(dbSQL.SqlStr)

		mockSqlFilter := "'1222', '2020-09-24 02:16:16'"
		//mockSqlFilter := "'11'',''1222', '2020-09-24 02:16:16'"

		sqlField := strings.Join(queryFields, "`, `")

		//应该返回的正确的sql语句  期望值
		sqlStr := fmt.Sprintf("INSERT INTO %s (`%s`) VALUES (%s)", tableName, sqlField, mockSqlFilter)
		fmt.Println(sqlStr)

		//把期望值和实际值都变成数组
		mockArrays := strings.Fields(sqlStr)
		sqlArrays := strings.Fields(dbSQL.SqlStr)

		//对比表名
		fmt.Println(assert.Equal(t, tableName, sqlArrays[2]))

		//全部变成切片排序后对比
		sort.Strings(mockArrays)
		sort.Strings(sqlArrays)
		fmt.Println(assert.Equal(t, mockArrays, sqlArrays))

	}

}

func TestUpdateSQL(t *testing.T) {

	filterFields := make([]dbbase.FilterField, 0, 8)
	filterFields = append(filterFields, dbbase.FilterField{
		Field:         dbbase.FieldCell{OriginData: 1001, BelongColumn: &col},
		CompareOption: "=",
	})
	filterFields = append(filterFields, dbbase.FilterField{
		Field:         dbbase.FieldCell{OriginData: "2020-09-23 02:16:16", BelongColumn: &col1},
		CompareOption: ">",
	})
	insertFieldCells := &[]dbbase.FieldCell{}

	*insertFieldCells = append(*insertFieldCells, dbbase.FieldCell{
		OriginData:   1222,
		BelongColumn: &col,
	})

	*insertFieldCells = append(*insertFieldCells, dbbase.FieldCell{
		OriginData:   "2021-03-2 03:26:32",
		BelongColumn: &col1,
	})
	//测试方法
	dbSQL, err := dbContent.GenerateUpdateSQL(tableName, *insertFieldCells, filterFields)
	//fmt.Println(dbSQL.SqlStr)
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	} else {
		t.Log(dbSQL.SqlStr)

		updateSqlField := "`node_id` = 1222, event_timestamp = '2021-03-2 03:26:32'"

		mockSqlFilter := "`node_id` = 1001 and event_timestamp > '2020-09-23 02:16:16'"
		//应该返回的正确的sql语句  期望值
		sqlStr := fmt.Sprintf("UPDATE %s SET %s WHERE %s", tableName, updateSqlField, mockSqlFilter)

		//把期望值和实际值都变成数组
		mockArrays := strings.Fields(sqlStr)
		sqlArrays := strings.Fields(dbSQL.SqlStr)

		//对比表名
		fmt.Println(assert.Equal(t, tableName, sqlArrays[1]))

		//全部变成切片排序后对比
		sort.Strings(mockArrays)
		sort.Strings(sqlArrays)
		assert.Equal(t, mockArrays, sqlArrays)
	}
}

func TestDeleteSQL(t *testing.T) {

	filterFields := make([]dbbase.FilterField, 0, 8) //结构体 filterfield 空切片  最大容量是8  过滤条件

	filterFields = append(filterFields, dbbase.FilterField{ //给这个空切片赋值
		Field:         dbbase.FieldCell{OriginData: 1001, BelongColumn: &col}, //结构体FilterField 赋值
		CompareOption: "=",
	})
	filterFields = append(filterFields, dbbase.FilterField{
		Field:         dbbase.FieldCell{OriginData: "2020-09-24 02:16:16", BelongColumn: &col1},
		CompareOption: ">",
	})

	// 测试删除语句的生成
	dbSQL, err := dbContent.GenerateDeleteSQL(tableName, filterFields)
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	} else {
		t.Log(dbSQL.SqlStr)
		fmt.Println(dbSQL.SqlStr)
		mockSqlFilter := "`node_id` = 1001 and event_timestamp > '2020-09-24 02:16:16'"
		//应该返回的正确的sql语句  期望值
		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, mockSqlFilter)
		//把期望值和实际值都变成数组
		mockArrays := strings.Fields(sqlStr)
		sqlArrays := strings.Fields(dbSQL.SqlStr)
		//对比表名
		fmt.Println(assert.Equal(t, tableName, sqlArrays[2]))

		//全部变成切片排序后对比
		sort.Strings(mockArrays)
		sort.Strings(sqlArrays)
		assert.Equal(t, mockArrays, sqlArrays)

	}

}

func TestGenerateSQL(t *testing.T) {
	dbContent := new(dbbase.DbContent)
	tableName := "events"
	queryFields := []string{"node_id", "event"}
	filterFields := make([]dbbase.FilterField, 0, 8)
	col := dbbase.Column{FieldName: "node_id", DataType: "int4", DataSize: 32}
	col1 := dbbase.Column{FieldName: "event_timestamp", DataType: "timestamptz", DataSize: -1}
	filterFields = append(filterFields, dbbase.FilterField{
		Field:         dbbase.FieldCell{OriginData: 1001, BelongColumn: &col},
		CompareOption: "=",
	})
	filterFields = append(filterFields, dbbase.FilterField{
		Field:         dbbase.FieldCell{OriginData: "2020-09-24 02:16:16", BelongColumn: &col1},
		CompareOption: ">",
	})

	condition := map[string]string{
		"limit":   "2",
		"offset":  "1",
		"orderBy": "name desc",
	}

	// 测试查询语句的生成
	dbSQL, err := dbContent.GenerateQuerySQL(tableName, queryFields, filterFields, condition)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(dbSQL.SqlStr)
	}

	insertFieldCells := &[]dbbase.FieldCell{}

	*insertFieldCells = append(*insertFieldCells, dbbase.FieldCell{
		OriginData:   1222,
		BelongColumn: &col,
	})

	*insertFieldCells = append(*insertFieldCells, dbbase.FieldCell{
		OriginData:   "2020-09-24 02:16:16",
		BelongColumn: &col1,
	})

	insertFields := []dbbase.Row{
		{
			FieldCells: *insertFieldCells,
		},
	}

	// 测试插入语句的生成
	dbSQL, err = dbContent.GenerateInsertSQL(tableName, insertFields)
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	} else {
		t.Log(dbSQL.SqlStr)
	}

	dbSQL, err = dbContent.GenerateUpdateSQL(tableName, *insertFieldCells, filterFields)
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	} else {
		t.Log(dbSQL.SqlStr)
	}

	dbSQL, err = dbContent.GenerateDeleteSQL(tableName, filterFields)
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	} else {
		t.Log(dbSQL.SqlStr)
	}
}

func TestDbRespondToJSONStr(t *testing.T) {

	er := errors.New("aaaaaaa")

	col := &[]dbbase.Column{}

	col1 := dbbase.Column{
		FieldName: "Id",
		DataType:  "integer",
		DataSize:  64,
	}

	col2 := dbbase.Column{
		FieldName: "Name",
		DataType:  "string",
		DataSize:  255,
	}

	*col = append(*col, col1)
	*col = append(*col, col2)

	rows := &[]dbbase.Row{}

	row1 := dbbase.Row{
		FieldCells: []dbbase.FieldCell{},
	}

	row2 := dbbase.Row{
		FieldCells: []dbbase.FieldCell{},
	}

	row1.FieldCells = append(row1.FieldCells, dbbase.FieldCell{
		OriginData:   1,
		BelongColumn: &col1,
	})

	row1.FieldCells = append(row1.FieldCells, dbbase.FieldCell{
		OriginData:   "jk",
		BelongColumn: &col2,
	})

	row2.FieldCells = append(row2.FieldCells, dbbase.FieldCell{
		OriginData:   2,
		BelongColumn: &col1,
	})

	row2.FieldCells = append(row2.FieldCells, dbbase.FieldCell{
		OriginData:   "zs",
		BelongColumn: &col2,
	})

	*rows = append(*rows, row1)
	*rows = append(*rows, row2)

	respond := dbbase.DbRespond{
		Err:           er,
		RespondStatus: true,
		RespondData: dbbase.DbTable{
			Cols: *col,
			Rows: *rows,
		},
	}

	jsonStr, err := dbbase.DbRespondToMap(respond)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(jsonStr)
}
