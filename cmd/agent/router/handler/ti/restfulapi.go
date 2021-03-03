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
	"dataapi/internal/kernel/metadata/swagger"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"strings"
	"time"
)

const TimeTemplate = "2006-01-02 15:04:05"//go中时间格式的模板

// GetData 获取一个Entity中的所有数据
func GetData(tableName string, condition map[string]string) (dbbase.DbRespond, map[string]interface{}) {
	res := dbbase.DbRespond{
		RespondStatus: false,
	}

	// 判断表是否在库中
	isTableValid, err := dbbase.UserDbContext.Valid(&tableName, []string{})
	if !isTableValid {
		res.Err = err
		return res, nil
	}

	// 获取select选定的字段名
	queryFields := make([]string, 0)
	selectArr := strings.Split(condition["select"],",")
	if len(selectArr) > 0 && selectArr[0] != ""{
		for index := range selectArr{
			queryFields = append(queryFields,selectArr[index])
		}
	}else {
		//如果没有select字段，那么就查所有
		for _, value := range dbbase.UserDbContext.AllFieldsOfTable[tableName] {
			queryFields = append(queryFields, "`"+value.FieldName+"`")
		}
	}

	// 根据字段名获取数据
	sql, err := dbbase.UserDbContext.GenerateQuerySQL(tableName, queryFields, []dbbase.FilterField{}, condition)

	if err != nil {
		res.Err = err
		return res, nil
	}

	// 执行Sql
	if err = dbbase.UserDbContext.RunQuery(sql); err != nil {
		res.Err = err
		return res, nil
	}

	res = *sql.Respond

	// 转为Json数据
	jsonResult, err := dbbase.DbRespondToMap(res)
	if err != nil {
		res.RespondStatus = false
		res.Err = err
		return res, nil
	}

	res.Err = errors.New("OK")
	return res, jsonResult
}

func GetOdataData(tableName string, condition map[string]string) (dbbase.DbRespond, map[string]interface{}) {
	res := dbbase.DbRespond{
		RespondStatus: false,
	}

	// 判断表是否在库中
	isTableValid, err := dbbase.UserDbContext.Valid(&tableName, []string{})
	if !isTableValid {
		res.Err = err
		return res, nil
	}

	// 获取select选定的字段名
	queryFields := make([]string, 0)
	selectArr := strings.Split(condition["select"],",")
	if len(selectArr) > 0 && selectArr[0] != ""{
		for index := range selectArr{
			queryFields = append(queryFields,selectArr[index])
		}
	}else {
		//如果没有select字段，那么就查所有
		for _, value := range dbbase.UserDbContext.AllFieldsOfTable[tableName] {
			queryFields = append(queryFields, "`"+value.FieldName+"`")
		}
	}

	// 根据字段名获取数据
	sql, err := dbbase.UserDbContext.GenerateOdataQuerySQL(tableName, queryFields, []dbbase.FilterField{}, condition)

	if err != nil {
		res.Err = err
		return res, nil
	}

	// 执行Sql
	if err = dbbase.UserDbContext.RunQuery(sql); err != nil {
		res.Err = err
		return res, nil
	}

	res = *sql.Respond

	// 转为Json数据
	jsonResult, err := dbbase.DbRespondToMap(res)
	if err != nil {
		res.RespondStatus = false
		res.Err = err
		return res, nil
	}

	res.Err = errors.New("OK")
	return res, jsonResult
}

// CreateData 增加多条新的数据
func CreateData(tableName string, fields []map[string]interface{}) dbbase.DbRespond {
	res := dbbase.DbRespond{
		RespondStatus: false,
	}

	if fields == nil || len(fields) < 0 {
		res.Err = errors.New("there is no data")
		return res
	}

	// 获取插入数据的字段名
	var fieldName []string
	for key, _ := range fields[0] {
		fieldName = append(fieldName, key)
	}

	isValid, err := dbbase.UserDbContext.Valid(&tableName, fieldName)
	if !isValid {
		res.Err = err
		return res
	}

	// 获取插入数据的所有数据
	insertFields := MapToRows(tableName, fields)

	sql, err := dbbase.UserDbContext.GenerateInsertSQL(tableName, insertFields)
	if err != nil {
		res.Err = err
		return res
	}

	if err = dbbase.UserDbContext.RunExec(sql); err != nil {
		res.Err = err
		return res
	}
	return *sql.Respond
}

// GetDataById 通过ID获取一条数据
func GetDataById(tableName string, id interface{}) (dbbase.DbRespond, map[string]interface{}) {
	res := dbbase.DbRespond{
		RespondStatus: false,
	}

	// 判断表是否在库中
	isTableValid, err := dbbase.UserDbContext.Valid(&tableName, []string{})
	if !isTableValid {
		res.Err = err
		return res, nil
	}

	// 获取所有字段名
	queryFields := make([]string, 0)
	for _, value := range dbbase.UserDbContext.AllFieldsOfTable[tableName] {
		queryFields = append(queryFields, "`"+value.FieldName+"`")
	}

	// 生成id约束条件
	filterFields := &[]dbbase.FilterField{}
	GetIdFilter(tableName, id, filterFields, &dbbase.UserDbContext)

	// 生成sql语句
	sql, err := dbbase.UserDbContext.GenerateQuerySQL(tableName, queryFields, *filterFields, nil)

	if err != nil {
		res.Err = err
		return res, nil
	}

	// 执行Sql
	if err = dbbase.UserDbContext.RunQuery(sql); err != nil {
		res.Err = err
		return res, nil
	}

	res = *sql.Respond

	// 转为Json数据
	jsonResult, err := dbbase.DbRespondToMap(res)
	if err != nil {
		res.RespondStatus = false
		res.Err = err
		return res, nil
	}

	res.Err = errors.New("OK")
	return res, jsonResult
}

func GetODataById(tableName string, id interface{}) (dbbase.DbRespond, map[string]interface{}) {
	res := dbbase.DbRespond{
		RespondStatus: false,
	}

	// 判断表是否在库中
	isTableValid, err := dbbase.UserDbContext.Valid(&tableName, []string{})
	if !isTableValid {
		res.Err = err
		return res, nil
	}

	// 获取所有字段名
	queryFields := make([]string, 0)
	for _, value := range dbbase.UserDbContext.AllFieldsOfTable[tableName] {
		queryFields = append(queryFields, value.FieldName)
	}

	// 生成id约束条件
	filterFields := &[]dbbase.FilterField{}
	GetIdFilter(tableName, id, filterFields, &dbbase.UserDbContext)

	// 生成sql语句
	sql, err := dbbase.UserDbContext.GenerateQuerySQL(tableName, queryFields, *filterFields, nil)

	if err != nil {
		res.Err = err
		return res, nil
	}

	// 执行Sql
	if err = dbbase.UserDbContext.RunQuery(sql); err != nil {
		res.Err = err
		return res, nil
	}

	res = *sql.Respond

	// 转为Json数据
	jsonResult, err := dbbase.DbRespondToMap(res)
	if err != nil {
		res.RespondStatus = false
		res.Err = err
		return res, nil
	}

	res.Err = errors.New("OK")
	return res, jsonResult
}

// UpdateDataById 通过ID更新某一条数据
func UpdateDataById(tableName string, id interface{}, fields map[string]interface{}) dbbase.DbRespond {
	res := dbbase.DbRespond{
		RespondStatus: false,
	}

	// 获取所有字段名
	var fieldName []string
	for key, _ := range fields {
		fieldName = append(fieldName, key)
	}

	// 检擦表名和字段是否合规
	isValid, err := dbbase.UserDbContext.Valid(&tableName, fieldName)
	if !isValid {
		res.Err = err
		return res
	}

	// 生成字段数组
	var fieldCells []dbbase.FieldCell
	for key, value := range fields {
		for index, field := range dbbase.UserDbContext.AllFieldsOfTable[tableName] {
			if strings.EqualFold(key , field.FieldName) {
				belongColumn := &dbbase.UserDbContext.AllFieldsOfTable[tableName][index]
				originData := value
				if belongColumn.DataType == "datetime" {
					timeStamp, err := time.Parse(TimeTemplate, fmt.Sprintf("%v",originData))
					if err != nil{
						timeStamp, _ = time.Parse(time.RFC3339, fmt.Sprintf("%v",originData))
					}
					originData = timeStamp.UTC().Format(TimeTemplate)
				}
				fieldCells = append(fieldCells, dbbase.FieldCell{
					OriginData: originData,
					BelongColumn: belongColumn,
				})
			}
		}
	}

	// 生成约束条件
	filterFields := &[]dbbase.FilterField{}
	GetIdFilter(tableName, id, filterFields, &dbbase.UserDbContext)

	// 生成更新语句
	sql, err := dbbase.UserDbContext.GenerateUpdateSQL(tableName, fieldCells, *filterFields)
	if err != nil {
		res.Err = err
		return res
	}

	// 执行
	if err = dbbase.UserDbContext.RunExec(sql); err != nil {
		res.Err = err
		return res
	}

	return *sql.Respond
}

// DeleteDataById 通过ID删除某一条数据
func DeleteDataById(tableName string, id interface{}) dbbase.DbRespond {
	res := dbbase.DbRespond{
		RespondStatus: false,
	}

	// 生成约束条件
	filterFields := &[]dbbase.FilterField{}
	GetIdFilter(tableName, id, filterFields, &dbbase.UserDbContext)

	// 生成删除语句
	sql, err := dbbase.UserDbContext.GenerateDeleteSQL(tableName, *filterFields)
	if err != nil {
		res.Err = err
		return res
	}

	// 执行
	if err = dbbase.UserDbContext.RunExec(sql); err != nil {
		res.Err = err
		return res
	}

	return *sql.Respond
}

// GetSwaggerJson 获取该实例的Swagger Json
func GetSwaggerJson() string {
	model := make(map[string]map[string]string)
	for _, table := range dbbase.UserDbContext.AllTablesOfSchema {
		var fieldNames = make(map[string]string)
		for _, fieldValue := range dbbase.UserDbContext.AllFieldsOfTable[table] {
			fieldNames[fieldValue.FieldName] = fieldValue.DataType
		}
		model[table] = fieldNames
	}

	host := os.Getenv("Environment_Host")
	return swagger.ModelToSwaggerJson(model, host)
}

// GetIdFilter 在约束条件中添加一个 id = ？
func GetIdFilter(tableName string, id interface{}, filterFields *[]dbbase.FilterField, dbContent *dbbase.DbContent) {
	for index, field := range dbContent.AllFieldsOfTable[tableName] {
		if field.FieldName == "id" || field.FieldName == "Id" {
			*filterFields = append(*filterFields, dbbase.FilterField{
				CompareOption: "=",
				Field: dbbase.FieldCell{
					OriginData:   id,
					BelongColumn: &dbContent.AllFieldsOfTable[tableName][index],
				},
			})
			break
		}
	}
}

// 获取host
func GetHostUrl() string {
	return os.Getenv("Environment_Host")
}

// MapToRows 将传入的map数组转成需要的多行数据结构
func MapToRows(tableName string, fields []map[string]interface{}) []dbbase.Row {

	insertFields := make([]dbbase.Row, 0)

	for index, _ := range fields {
		var fieldCells []dbbase.FieldCell
		for tableIndex, tableValue := range dbbase.UserDbContext.AllFieldsOfTable[tableName] {
			for key, _ := range fields[index] {
				if strings.EqualFold(tableValue.FieldName, key) {
					belongColumn := &dbbase.UserDbContext.AllFieldsOfTable[tableName][tableIndex]
					originData := fields[index][key]
					if belongColumn.DataType == "datetime" {
						timeStamp, err := time.Parse(TimeTemplate, fmt.Sprintf("%v",originData))
						if err != nil{
							timeStamp, _ = time.Parse(time.RFC3339, fmt.Sprintf("%v",originData))
						}
						originData = timeStamp.UTC().Format(TimeTemplate)
					}
					fieldCells = append(fieldCells, dbbase.FieldCell{
						OriginData: originData,
						BelongColumn: belongColumn,
					})
				}
			}
		}
		insertFields = append(insertFields, dbbase.Row{
			FieldCells: fieldCells,
		})
	}

	return insertFields
}

func GetCondition(c *gin.Context)  map[string]string{
	topClause := c.Query("$top")
	skipClause := c.Query("$skip")
	selectClause := c.Query("$select")
	orderbyClause := c.Query("$orderby")
	filterClause := c.Query("$filter")
	return map[string]string{
		"top": topClause,
		"skip": skipClause,
		"select": selectClause,
		"orderby": orderbyClause,
		"filter": filterClause,
	}
}

//获取指定表的总的条数
func GetCount(tableName string) (dbbase.DbRespond,map[string]interface{}) {
	res := dbbase.DbRespond{
		RespondStatus: false,
	}
	sql, err := dbbase.UserDbContext.GenerateCountSQL(tableName)
	// 查询
	if err = dbbase.UserDbContext.RunQuery(sql); err != nil {
		res.Err = err
		return res, nil
	}
	res = *sql.Respond
	jsonResult, err := dbbase.DbRespondToMap(res)
	if err != nil {
		res.RespondStatus = false
		res.Err = err
		return res, nil
	}

	return res,jsonResult
}

//传来的tableName可能含有主键，需要判断，有的返回到map的v中
//如果带主键，那么形式是：people(12)，或者是people('red')
func GetPrimaryKey(tableName string) map[string]string{
	mapResult := map[string]string{}
	strArr := strings.Split(tableName,"(")
	if len(strArr) == 1 {
		//说明无主键
		mapResult[strArr[0]] = ""
	}else if len(strArr) == 2{
		//说明有主键，结构例如[people,34)],后面一个元素去)
		mapResult[strArr[0]] = strings.Trim(strArr[1],")")
	}
	return mapResult
}