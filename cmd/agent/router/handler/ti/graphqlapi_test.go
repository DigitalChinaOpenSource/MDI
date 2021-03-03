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
	"testing"
)

// func TestGenerateCreateSql(t *testing.T) {
// 	tableName := "test"
// 	args := make(map[string]interface{})
// 	args["id"] = 1
// 	args["name"] = "ll"
// 	sql := GenerateCreateSql(tableName,args)
// 	expectSql := "INSERT INTO  test(id,name) VALUES(1,'ll')"
// 	if strings.Trim(expectSql," ") != strings.Trim(sql," ")  {
// 		t.Fatal("未预期的sql:"+sql)
// 	} else {
// 		t.Log("生成的sql语句："+sql)
// 	}
// }

func TestGenerateUpdateSql(t *testing.T) {
	//tableName := "test"
	//args := make(map[string]interface{})
	//args["id"] = 1
	//args["name"] = "ll"
	//args["age"] = 98
	//sql := GenerateUpdateSql(tableName,args)
	//expectSql := "UPDATE "+tableName+" SET name = '"+args["name"].(string)+"' ,age="+strconv.Itoa(args["age"].(int))+" WHERE id = "+strconv.Itoa(args["id"].(int))
	//if strings.Trim(expectSql," ") != strings.Trim(sql," ") {
	//	t.Fatal("未预期的sql:"+sql)
	//} else {
	//	t.Log("生成的sql语句："+sql)
	//}
}

func TestConcatConditionToSql(t *testing.T) {
	sql := "SELECT * FROM test"
	order := " ORDER BY id desc"
	limit := " LIMIT 5 "
	offset := " OFFSET 2 "
	sqlStr := ConcatConditionToSql(order, limit, offset, sql)
	if sql+order+limit+offset != sqlStr {
		t.Fatal("未预期的sql:" + sql)
	} else {
		t.Log("生成的sql语句：" + sql)
	}
}
func TestGetWhereInputObjectByType(t *testing.T) {
	intObj := GetWhereInputObjectByType("integer")
	strObj := GetWhereInputObjectByType("string")
	datetimeObj := GetWhereInputObjectByType("datetime")
	if intObj != nil {
		t.Log("创建int对象成功")
	} else {
		t.Fatal("创建int对象失败")
	}
	if strObj != nil {
		t.Log("创建string对象成功")
	} else {
		t.Fatal("创建string对象失败")
	}
	if datetimeObj != nil {
		t.Log("创建datetime对象成功")
	} else {
		t.Fatal("创建datetime对象失败")
	}
}

func TestAddConditionToSql(t *testing.T) {
	//fromFieldAttr := "name"
	//foreignFieldAttr := "name"
	//sql := ""
	//p := graphql.ResolveParams{}
	//p.Args["name"] = "ll"
	//p.Args["id"] = 99
	////p.Source = map[string]interface{}{"name":"llllllll"}
	//AddConditionToSql(p,fromFieldAttr,foreignFieldAttr,sql)
}
