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
	"bytes"
	"dataapi/cmd/agent/router/handler/dbbase"
	"dataapi/internal/kernel/metadata/modeling"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"

	regexp2 "github.com/dlclark/regexp2"

	"github.com/wxnacy/wgo/arrays"

	"strings"
	//"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/handler"

	"github.com/graphql-go/graphql"
)

const (
	AND     = " AND "
	ORDERBY = " ORDER BY "
	WHERE   = " WHERE "
	OR      = " OR "
	LIMIT   = " LIMIT "
	OFFSET  = " OFFSET "
)

func GetSchema() graphql.Schema {
	//模型对象数组
	modelTypes := make([]*graphql.Object, len(dbbase.UserDbContext.AllTablesOfSchema))

	//查询对象
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Query",
		Fields: graphql.Fields{},
	})

	//突变对象
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Mutation",
		Fields: graphql.Fields{},
	})

	//schema对象
	schema := graphql.Schema{}

	//参数数组,一开始curd都用的一个参数数组。这样在开发过程中果然遇到了问题。所以需要分开，不同的行为用不同的参数数组。
	//argsArr := make([]map[string]graphql.FieldConfigArgument, 0)
	queryArgsArr := make([]map[string]graphql.FieldConfigArgument, 0)
	insertArgsArr := make([]map[string]graphql.FieldConfigArgument, 0)
	updateArgsArr := make([]map[string]graphql.FieldConfigArgument, 0)
	deleteArgsArr := make([]map[string]graphql.FieldConfigArgument, 0)
	//根据模型添加参数
	for tableIndex := range dbbase.UserDbContext.AllTablesOfSchema {
		tableValue := dbbase.UserDbContext.AllTablesOfSchema[tableIndex]
		//创建graphql实体模型对象,字段动态加入进去
		modelTypes[tableIndex] = graphql.NewObject(
			graphql.ObjectConfig{
				Name:   tableValue,
				Fields: graphql.Fields{},
			})

		//modelType循环添加映射字段
		for _, fieldValue := range dbbase.UserDbContext.AllFieldsOfTable[tableValue] {
			modelTypes[tableIndex].AddFieldConfig(fieldValue.FieldName, &graphql.Field{
				Type: TypeToGraphqlOutput(fieldValue.DataType),
			})
		}
	}

	//向查询方法的参数数组中添加参数，包括基本字段和与之关联的对象字段
	queryArgsArr = AddParamToQueryArr(queryArgsArr)

	//通过关系数组为模型对象添加关联属性
	AddRelationToModelType(modelTypes, queryArgsArr)

	//向插入方法的参数数组中添加参数，包括基本字段和与之关联的对象字段
	insertArgsArr = AddParamToInsertArr(insertArgsArr, modelTypes)
	updateArgsArr = AddParamToUpdateArr(updateArgsArr)
	deleteArgsArr = AddParamToDeleteArr(deleteArgsArr)

	//CURD方法数组
	queryFnArr := make([]graphql.FieldResolveFn, len(dbbase.UserDbContext.AllTablesOfSchema))
	insertFnArr := make([]graphql.FieldResolveFn, len(dbbase.UserDbContext.AllTablesOfSchema))
	updateFnArr := make([]graphql.FieldResolveFn, len(dbbase.UserDbContext.AllTablesOfSchema))
	deleteFnArr := make([]graphql.FieldResolveFn, len(dbbase.UserDbContext.AllTablesOfSchema))

	//依次生成方法并加入方法数组
	for tableIndex := range dbbase.UserDbContext.AllTablesOfSchema {
		tableName := dbbase.UserDbContext.AllTablesOfSchema[tableIndex]
		queryFnArr[tableIndex] = GenerateQueryFunc(tableName)
		insertFnArr[tableIndex] = GenerateInsertFunc(tableName)
		updateFnArr[tableIndex] = GenerateUpdateFunc(tableName)
		deleteFnArr[tableIndex] = GenerateDeleteFunc(tableName, modelTypes)
	}

	for tableIndex := range dbbase.UserDbContext.AllTablesOfSchema {
		tableValue := dbbase.UserDbContext.AllTablesOfSchema[tableIndex]
		//添加查询的Field
		queryType.AddFieldConfig(tableValue, &graphql.Field{
			Type:    graphql.NewList(modelTypes[tableIndex]),
			Args:    queryArgsArr[tableIndex][tableValue],
			Resolve: queryFnArr[tableIndex],
		})

		//添加突变的Field,目前有三个操作delete,insert,update
		//添加操作
		mutationType.AddFieldConfig("create_"+tableValue, &graphql.Field{
			Type:    modelTypes[tableIndex],
			Args:    insertArgsArr[tableIndex][tableValue],
			Resolve: insertFnArr[tableIndex],
		})

		//更新操作
		mutationType.AddFieldConfig("update_"+tableValue, &graphql.Field{
			Type:    modelTypes[tableIndex],
			Args:    updateArgsArr[tableIndex][tableValue],
			Resolve: updateFnArr[tableIndex],
		})

		//删除操作
		mutationType.AddFieldConfig("delete_"+tableValue, &graphql.Field{
			Type:    modelTypes[tableIndex],
			Args:    deleteArgsArr[tableIndex][tableValue],
			Resolve: deleteFnArr[tableIndex],
		})
	}

	//将查询和突变对象传入schema中
	schema, _ = graphql.NewSchema(
		graphql.SchemaConfig{
			Query:    queryType,
			Mutation: mutationType,
		})
	return schema
}

func AddParamToInsertArr(argsArr []map[string]graphql.FieldConfigArgument, modelTypes []*graphql.Object) []map[string]graphql.FieldConfigArgument {
	for tableIndex := range dbbase.UserDbContext.AllTablesOfSchema {
		tableValue := dbbase.UserDbContext.AllTablesOfSchema[tableIndex]
		//argsArr数组是输入参数数组
		args := make(map[string]graphql.FieldConfigArgument)
		tempArgs := graphql.FieldConfigArgument{}
		//添加基本参数
		tempArgs = AddBasicParam(tempArgs, tableValue)
		//添加关联对象参数，比如A表中某条数据外键关联b表的多条数据，插入时我们想一次性插入A以及与它相关联的多个B。以免我先插A，然后插多条B
		AddForeignObj(modelTypes, tempArgs, tableValue)
		args[tableValue] = tempArgs
		argsArr = append(argsArr, args)
	}
	return argsArr
}

func AddParamToQueryArr(argsArr []map[string]graphql.FieldConfigArgument) []map[string]graphql.FieldConfigArgument {
	for tableIndex := range dbbase.UserDbContext.AllTablesOfSchema {
		tableValue := dbbase.UserDbContext.AllTablesOfSchema[tableIndex]
		//argsArr数组是输入参数数组
		args := make(map[string]graphql.FieldConfigArgument)
		tempArgs := graphql.FieldConfigArgument{}
		tempArgs = AddBasicParam(tempArgs, tableValue)
		tempArgs["where"] = &graphql.ArgumentConfig{
			Type: GenerateWhereType(tableValue),
		}
		//tempArgs["test"] = &graphql.ArgumentConfig{
		//	Type:graphql.NewScalar(graphql.ScalarConfig{
		//		Name: "test",
		//		Description: "test",
		//		Serialize: func(value interface{}) interface{} {
		//
		//		},
		//	}),
		//}
		tempArgs["order_by"] = &graphql.ArgumentConfig{
			Type: GenerateOrderType(tableValue),
		}
		tempArgs["limit"] = &graphql.ArgumentConfig{
			Type: graphql.Int,
		}
		tempArgs["offset"] = &graphql.ArgumentConfig{
			Type: graphql.Int,
		}
		args[tableValue] = tempArgs
		argsArr = append(argsArr, args)
	}
	return argsArr
}

func AddParamToUpdateArr(argsArr []map[string]graphql.FieldConfigArgument) []map[string]graphql.FieldConfigArgument {
	for tableIndex := range dbbase.UserDbContext.AllTablesOfSchema {
		tableValue := dbbase.UserDbContext.AllTablesOfSchema[tableIndex]
		//argsArr数组是输入参数数组
		args := make(map[string]graphql.FieldConfigArgument)
		tempArgs := graphql.FieldConfigArgument{}
		tempArgs = AddBasicParam(tempArgs, tableValue)
		//添加Where条件输入参数类型
		tempArgs["where"] = &graphql.ArgumentConfig{
			Type: GenerateWhereType(tableValue),
		}
		args[tableValue] = tempArgs
		argsArr = append(argsArr, args)
	}
	return argsArr
}

func AddParamToDeleteArr(argsArr []map[string]graphql.FieldConfigArgument) []map[string]graphql.FieldConfigArgument {
	for tableIndex := range dbbase.UserDbContext.AllTablesOfSchema {
		tableValue := dbbase.UserDbContext.AllTablesOfSchema[tableIndex]
		//argsArr数组是输入参数数组
		args := make(map[string]graphql.FieldConfigArgument)
		tempArgs := graphql.FieldConfigArgument{}
		tempArgs["where"] = &graphql.ArgumentConfig{
			Type: GenerateWhereType(tableValue),
		}
		args[tableValue] = tempArgs
		argsArr = append(argsArr, args)
	}
	return argsArr
}

func AddBasicParam(args graphql.FieldConfigArgument, tableValue string) graphql.FieldConfigArgument {
	for _, fieldValue := range dbbase.UserDbContext.AllFieldsOfTable[tableValue] {
		args[fieldValue.FieldName] = &graphql.ArgumentConfig{
			Type: TypeToGraphqlInput(fieldValue.DataType),
		}
	}
	return args
}

func GenerateWhereType(tableValue string) *graphql.InputObject {
	whereType := graphql.NewInputObject(
		graphql.InputObjectConfig{
			Name:   "where",
			Fields: graphql.InputObjectConfigFieldMap{},
		},
	)
	for index := range dbbase.UserDbContext.AllFieldsOfTable[tableValue] {
		fieldValue := dbbase.UserDbContext.AllFieldsOfTable[tableValue][index]
		if fieldValue.DataType == "integer" {
			whereType.AddFieldConfig(
				fieldValue.FieldName, &graphql.InputObjectFieldConfig{
					Type: GetWhereInputObjectByType("integer"),
				})
		} else if fieldValue.DataType == "string" {
			whereType.AddFieldConfig(
				fieldValue.FieldName, &graphql.InputObjectFieldConfig{
					Type: GetWhereInputObjectByType("string"),
				})
		} else if fieldValue.DataType == "datetime" {
			whereType.AddFieldConfig(
				fieldValue.FieldName, &graphql.InputObjectFieldConfig{
					Type: GetWhereInputObjectByType("datetime"),
				})
		}
	}
	return whereType
}

func GenerateOrderType(tableValue string) *graphql.InputObject {
	orderType := graphql.NewInputObject(
		graphql.InputObjectConfig{
			Name:   "order_by",
			Fields: graphql.InputObjectConfigFieldMap{},
		})
	for index := range dbbase.UserDbContext.AllFieldsOfTable[tableValue] {
		fieldValue := dbbase.UserDbContext.AllFieldsOfTable[tableValue][index]
		orderType.AddFieldConfig(
			fieldValue.FieldName, &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			})
	}
	return orderType
}

func AddForeignObj(modelTypes []*graphql.Object, args graphql.FieldConfigArgument, tableName string) {
	relations := dbbase.UserDbContext.DbEntity.ForeignKeys
	if relations != nil {
		for index := range *relations {
			relation := (*relations)[index]
			if relation.FromEntityAttr == tableName {
				fieldName := relation.FromEntityAttr + "_" + relation.FromEntityRelation + "_" + relation.ForeignEntityRelation + "_" + relation.ForeignEntityAttr
				objType := graphql.NewInputObject(
					graphql.InputObjectConfig{
						Name:   fieldName,
						Fields: graphql.InputObjectConfigFieldMap{},
					})
				obj := GetModelTypeByName(modelTypes, relation.ForeignEntityAttr)
				GenerateArgs(obj, objType, modelTypes, relation.ForeignEntityAttr, relation.FromEntityAttr)
				args[fieldName] = &graphql.ArgumentConfig{
					Type: objType,
				}
			}
		}
	}
}

//递归地生成对象类型地参数对象
func GenerateArgs(obj *graphql.Object, objType *graphql.InputObject, modelTypes []*graphql.Object, foreignEntityAttr string, fromEntityAttr string) *graphql.InputObject {
	//obj := GetModelTypeByName(modelTypes,foreignEntityAttr)
	for k, v := range obj.Fields() {
		if v.Type.Name() == "Int" {
			objType.AddFieldConfig(k,
				&graphql.InputObjectFieldConfig{
					Type: graphql.Int,
				})
		} else if v.Type.Name() == "String" {
			objType.AddFieldConfig(k,
				&graphql.InputObjectFieldConfig{
					Type: graphql.String,
				})
		} else {
			//对象属性，比如test里面可以插people，people里面可以插school
			//此时需要根据foreignEntityAttr再去relations中看foreignEntityAttr还有没有关系
			arr := GetRelationAttrByFromEntityAttr(foreignEntityAttr)
			for index := range arr {
				relation := arr[index]
				// 双外键退出，避免无线递归
				if relation.ForeignEntityAttr == fromEntityAttr {
					continue
				}
				objParam := GetModelTypeByName(modelTypes, relation.ForeignEntityAttr)
				fieldName := relation.FromEntityAttr + "_" + relation.FromEntityRelation + "_" + relation.ForeignEntityRelation + "_" + relation.ForeignEntityAttr
				objTypeParam := graphql.NewInputObject(
					graphql.InputObjectConfig{
						Name:   fieldName,
						Fields: graphql.InputObjectConfigFieldMap{},
					})
				objType.AddFieldConfig(k,
					&graphql.InputObjectFieldConfig{
						Type: GenerateArgs(objParam, objTypeParam, modelTypes, relation.ForeignEntityAttr, relation.FromEntityAttr),
					})
			}
		}
	}
	return objType
}

//通过fromEntityAttr获取对应的relation
func GetRelationAttrByFromEntityAttr(fromEntityAttr string) []modeling.ForeignKey {
	arr := make([]modeling.ForeignKey, 0)
	relations := dbbase.UserDbContext.DbEntity.ForeignKeys
	if relations != nil {
		for index := range *relations {
			relation := (*relations)[index]
			if relation.FromEntityAttr == fromEntityAttr {
				arr = append(arr, relation)
			}
		}
	}
	return arr
}

//通过关系数组为模型对象添加关联属性
func AddRelationToModelType(modelTypes []*graphql.Object, argsArr []map[string]graphql.FieldConfigArgument) {
	relations := dbbase.UserDbContext.DbEntity.ForeignKeys
	if relations != nil {
		for index := range *relations {
			relation := (*relations)[index]
			for index, modelType := range modelTypes {
				if relation.FromEntityAttr == modelType.PrivateName {
					configArgs := GetArgsByTargetName(argsArr, relation.ForeignEntityAttr)
					modelTypes[index].AddFieldConfig(relation.ForeignEntityAttr+"__"+relation.ForeignEntityRelation,
						&graphql.Field{
							Type: graphql.NewList(GetModelTypeByName(modelTypes, relation.ForeignEntityAttr)),
							Args: configArgs,
							Resolve: func(p graphql.ResolveParams) (interface{}, error) {
								sql := GenerateQuerySql(relation.ForeignEntityAttr, p.Args)
								//根据关系拼接关联字段的条件
								sql = AddConditionToSql(p, relation.FromFieldAttr, relation.ForeignFieldAttr, sql)
								dbSQL := dbbase.DbSQL{
									TableName: relation.ForeignEntityAttr,
									SqlStr:    sql,
								}
								dbbase.UserDbContext.RunQuery(&dbSQL)
								dbModelMaps, _ := dbbase.GetDataArr(*dbSQL.Respond)
								return dbModelMaps, nil
							},
						})
					break
				}
			}
		}
	}
}

func AddConditionToSql(p graphql.ResolveParams, fromFieldAttr string, foreignFieldAttr string, sql string) string {
	source := (p.Source).(map[string]interface{})
	sourceAttr := source[fromFieldAttr]
	sqlStrArr := make([]string, 0)
	if len(p.Args) == 0 {
		//如果没有传任何参数，那么sql就是select * from table，接下来的工作很简单，加个“WHERE”，接着拼接条件就行
		sql += WHERE
		switch sourceAttr.(type) {
		case int, int8, int16, int32, int64:
			sql += foreignFieldAttr + "=" + strconv.Itoa(sourceAttr.(int))
		case string:
			sql += foreignFieldAttr + "=" + "'" + sourceAttr.(string) + "'"
		}
	} else {
		//如果传了参数，那么sql可能是select * from table where a=a。或者是 select * from table where b=b order by a
		//亦或者是select * from table order by b
		//只有当条件中不存在order_by,limit,offset时，才能直接拼AND
		condition := p.Args["order_by"] == nil && p.Args["limit"] == nil && p.Args["offset"] == nil
		onlyOrderLimitOffset := true
		for k := range p.Args {
			if k != "order_by" && k != "limit" && k != "offset" {
				onlyOrderLimitOffset = false
				break
			}
		}
		if condition {
			//没有order_by,limit,offset条件,直接拼接“AND”,原sql是这种形式：select * from table where a=a AND b=b ...
			sql += AND
		} else {

			if onlyOrderLimitOffset {
				//sql : select * from table order by id desc limit 2 offset 2
				if strings.Index(sql, ORDERBY) != -1 {
					sqlStrArr = strings.Split(sql, ORDERBY)
					sqlStrArr[0] += WHERE
					sqlStrArr[1] = ORDERBY + sqlStrArr[1]
				} else if strings.Index(sql, LIMIT) != -1 {
					sqlStrArr = strings.Split(sql, LIMIT)
					sqlStrArr[0] += WHERE
					sqlStrArr[1] = LIMIT + sqlStrArr[1]
				} else if strings.Index(sql, OFFSET) != -1 {
					sqlStrArr = strings.Split(sql, OFFSET)
					sqlStrArr[0] += WHERE
					sqlStrArr[1] = OFFSET + sqlStrArr[1]
				}
			} else {
				// sql : select * from table where name = 'ss' order by id desc
				if strings.Index(sql, ORDERBY) != -1 {
					sqlStrArr = strings.Split(sql, ORDERBY)
					sqlStrArr[0] += AND
					sqlStrArr[1] = ORDERBY + sqlStrArr[1]
				} else if strings.Index(sql, LIMIT) != -1 {
					sqlStrArr = strings.Split(sql, LIMIT)
					sqlStrArr[0] += AND
					sqlStrArr[1] = LIMIT + sqlStrArr[1]
				} else if strings.Index(sql, OFFSET) != -1 {
					sqlStrArr = strings.Split(sql, OFFSET)
					sqlStrArr[0] += AND
					sqlStrArr[1] = OFFSET + sqlStrArr[1]
				}
			}
			switch sourceAttr.(type) {
			case int, int8, int16, int32, int64:
				sqlStrArr[0] += foreignFieldAttr + "=" + strconv.Itoa(sourceAttr.(int))
			case string:
				sqlStrArr[0] += foreignFieldAttr + "=" + "'" + sourceAttr.(string) + "'"
			}
			sql = sqlStrArr[0] + sqlStrArr[1]
		}
	}
	sql = strings.Trim(strings.Trim(sql, " "), "AND")
	return sql
}

func GetWhereInputObjectByType(datatype string) *graphql.InputObject {
	if "integer" == datatype {
		return graphql.NewInputObject(
			graphql.InputObjectConfig{
				Name: "IntObjOp",
				Fields: graphql.InputObjectConfigFieldMap{
					"gt": &graphql.InputObjectFieldConfig{
						Type: graphql.Int,
					},
					"lt": &graphql.InputObjectFieldConfig{
						Type: graphql.Int,
					},
					"eq": &graphql.InputObjectFieldConfig{
						Type: graphql.Int,
					},
				},
			},
		)
	} else if "string" == datatype {
		return graphql.NewInputObject(
			graphql.InputObjectConfig{
				Name: "StrObjOp",
				Fields: graphql.InputObjectConfigFieldMap{
					"eq": &graphql.InputObjectFieldConfig{
						Type: graphql.String,
					},
				},
			},
		)
	} else if "datetime" == datatype {
		return graphql.NewInputObject(
			graphql.InputObjectConfig{
				Name: "datetimeOpObj",
				Fields: graphql.InputObjectConfigFieldMap{
					"gt": &graphql.InputObjectFieldConfig{
						Type: graphql.String,
					},
					"lt": &graphql.InputObjectFieldConfig{
						Type: graphql.String,
					},
					"eq": &graphql.InputObjectFieldConfig{
						Type: graphql.String,
					},
				},
			},
		)
	}
	return graphql.NewInputObject(graphql.InputObjectConfig{})
}

func GetArgsByTargetName(argsArr []map[string]graphql.FieldConfigArgument, target string) graphql.FieldConfigArgument {
	if len(argsArr) == 0 {
		return nil
	}
	for index := range argsArr {
		for argK, argV := range argsArr[index] {
			if argK == target {
				return argV
			}
		}
	}
	return nil
}

//通过名字返回所需的graphql.Object对象
func GetModelTypeByName(modelTypeArr []*graphql.Object, targetName string) *graphql.Object {
	for index, model := range modelTypeArr {
		if model.Name() == targetName {
			return modelTypeArr[index]
		}
	}
	return nil
}

//检查graphql.object对象的属性中有没有其他对象，即嵌套关系，有则以 关系名：实际表名 的键值对形式返回
func CheckSubObjectsExists(objects []*graphql.Object, targetName string) map[string]string {
	object := GetModelTypeByName(objects, targetName)
	if object == nil {
		return map[string]string{}
	}
	map1 := make(map[string]string)
	for k, v := range object.Fields() {
		for _, tableName := range dbbase.UserDbContext.AllTablesOfSchema {
			if v.Type.Name() == tableName {
				map1[k] = tableName
			}
		}
	}
	return map1
}

//用GetDataArr从数据库查出的对象只有基础数据，对象类型的字段需要根据graphql.object的属性来添加、
func AddObjectTypeAttr(maps []map[string]interface{}, objects []*graphql.Object, tableName string) []map[string]interface{} {
	if len(maps) == 0 {
		return nil
	}
	currentModel := GetModelTypeByName(objects, tableName)
	fields := currentModel.Fields()

	fieldKeys := make([]string, 0)
	for fieldKey := range fields {
		fieldKeys = append(fieldKeys, fieldKey)
	}
	keys := make([]string, 0)
	for key := range maps[0] {
		keys = append(keys, key)
	}
	for index := range maps {
		for f := range fieldKeys {
			if arrays.ContainsString(keys, fieldKeys[f]) == -1 {
				maps[index][fieldKeys[f]] = nil
			}
		}
	}
	return maps
}

//递归地设置类型为[]map[string]interface的对象的属性值
func SetMapData(maps []map[string]interface{}, objects []*graphql.Object, tableName string) []map[string]interface{} {
	for index := range maps {
		currentMap := maps[index]
		for k := range currentMap {
			result := CheckSubObjectsExists(objects, strings.Split(k, "_")[0])
			if len(result) != 0 {
				for k1, v1 := range result {
					condition := strings.Split(k1, "_")
					sql := "SELECT * FROM " + v1 + " WHERE " + strings.ToLower(condition[4]) + "="
					val := currentMap[strings.ToLower(condition[2])]
					switch val.(type) {
					case int:
						sql += strconv.Itoa(val.(int))
					case string:
						sql += "'" + val.(string) + "'"
					}
					dbSql1 := dbbase.DbSQL{
						TableName: v1,
						SqlStr:    sql,
					}
					dbbase.UserDbContext.RunQuery(&dbSql1)
					tempMaps1, _ := dbbase.GetDataArr(*dbSql1.Respond)
					tempMaps1 = AddObjectTypeAttr(tempMaps1, objects, v1)
					tempMaps1 = SetMapData(tempMaps1, objects, tableName)
					currentMap[k1] = tempMaps1
				}
			}
		}
	}
	return maps
}

//生成查询方法，需要考虑关联查询
func GenerateQueryFunc(tableName string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		sql := GenerateQuerySql(tableName, p.Args)
		dbSQL := dbbase.DbSQL{
			TableName: tableName,
			SqlStr:    sql,
		}
		dbbase.UserDbContext.RunQuery(&dbSQL)
		dbModelMaps, _ := dbbase.GetDataArr(*dbSQL.Respond)
		return dbModelMaps, nil
	}
}

func GenerateInsertFunc(tableName string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		sql := GenerateCreateSql(tableName, p.Args)
		dbSQL := dbbase.DbSQL{
			TableName: tableName,
			SqlStr:    sql,
		}
		//执行插入操作
		dbbase.UserDbContext.RunExec(&dbSQL)
		//查询结果集
		querySql := GenerateQuerySql(tableName, p.Args)
		dbSQL1 := dbbase.DbSQL{
			TableName: tableName,
			SqlStr:    querySql,
		}
		dbbase.UserDbContext.RunQuery(&dbSQL1)
		dbModelMaps, _ := dbbase.GetDataArr(*dbSQL1.Respond)
		return dbModelMaps[0], nil
	}
}

func GenerateUpdateFunc(tableName string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		sql := GenerateUpdateSql(tableName, p.Args)
		dbSQL := dbbase.DbSQL{
			TableName: tableName,
			SqlStr:    sql,
		}
		//执行更新操作
		dbbase.UserDbContext.RunExec(&dbSQL)

		//查询结果集
		querySql := GenerateQuerySql(tableName, p.Args)
		dbSQL1 := dbbase.DbSQL{
			TableName: tableName,
			SqlStr:    querySql,
		}
		//查询数据库最新数据
		dbbase.UserDbContext.RunQuery(&dbSQL1)
		dbModelMaps, _ := dbbase.GetDataArr(*dbSQL1.Respond)
		//更新子表的数据
		UpdateSubTable(dbModelMaps, tableName)
		return dbModelMaps[0], nil
	}
}

func UpdateSubTable(maps []map[string]interface{}, tableName string) {
	sql := ""
	if len(maps) != 0 {
		relations := dbbase.UserDbContext.DbEntity.ForeignKeys
		if relations != nil {
			for index := range *relations {
				relation := (*relations)[index]
				if relation.FromEntityAttr == tableName {
					for index := range maps {
						data := maps[index]
						sql += " UPDATE " + relation.ForeignEntityAttr + " SET " + relation.ForeignFieldAttr + "="
						field := data[relation.FromFieldAttr]
						switch field.(type) {
						case int:
							sql += strconv.Itoa(field.(int))
						case string:
							sql += "'" + field.(string) + "'"
						}
						sql += ";"
					}
				}
			}
		}
	}
	//根据生成的sql更新
	dbSQL1 := dbbase.DbSQL{
		TableName: tableName,
		SqlStr:    sql,
	}
	//查询数据库最新数据
	dbbase.UserDbContext.RunQuery(&dbSQL1)
}

func GenerateDeleteFunc(tableName string, modelTypes []*graphql.Object) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		//查询本表结果集
		querySql := GenerateQuerySql(tableName, p.Args)
		dbSQL := dbbase.DbSQL{
			TableName: tableName,
			SqlStr:    querySql,
		}
		dbbase.UserDbContext.RunQuery(&dbSQL)
		dbModelMaps, _ := dbbase.GetDataArr(*dbSQL.Respond)
		dbModelMaps = AddObjectTypeAttr(dbModelMaps, modelTypes, tableName)
		dbModelMaps = SetMapData(dbModelMaps, modelTypes, tableName)
		sql := GenerateDeleteSql(tableName, p.Args, dbModelMaps)
		dbSQL1 := dbbase.DbSQL{
			TableName: tableName,
			SqlStr:    sql,
		}
		//执行更新操作
		dbbase.UserDbContext.RunExec(&dbSQL1)
		return dbModelMaps[0], nil
	}
}

//go基础类型转graphql输出类型
func TypeToGraphqlOutput(t string) graphql.Output {
	map1 := map[string]graphql.Output{
		"integer":  graphql.Int,
		"string":   graphql.String,
		"decimal":  graphql.Int,
		"datetime": graphql.String,
		"boolean":  graphql.Boolean,
		"money":    graphql.Float,
	}
	return map1[t]
}

//go基础类型转graphql输入类型
func TypeToGraphqlInput(t string) graphql.Input {
	map1 := map[string]graphql.Output{
		"integer":  graphql.Int,
		"string":   graphql.String,
		"decimal":  graphql.Int,
		"datetime": graphql.String,
		"boolean":  graphql.Boolean,
		"money":    graphql.Float,
	}
	return map1[t]
}

//根据条件参数生成查询sql
func GenerateQuerySql(tableName string, args map[string]interface{}) string {
	sqlStr := "SELECT * FROM " + tableName
	if len(args) == 0 {
		return sqlStr
	} else {
		//有参数，但是需要判断情况
		//1.只有普通参数，如name:"ll" sqlStr拼接"WHERE"
		//2.只有条件参数，比如where：{id:{gt:3}},order_by,limit等,在有where条件时，也要加"WHERE"，其他的不加。
		//3.二者皆有,sqlStr拼接"WHERE"
		sqlStr = ConcatWhereToSql(args, sqlStr)
	}
	orderByStr := ORDERBY
	limitStr := LIMIT
	offsetStr := OFFSET
	for argK, argV := range args {
		//条件
		if argK == "where" {
			mapV := argV.(map[string]interface{})
			if len(mapV) != 0 {
				for paramK, paramV := range mapV {
					conditionV := paramV.(map[string]interface{})
					sqlStr = ConcatEqLtGtToSql(sqlStr, paramK, conditionV)
				}
			}
		} else if argK == "order_by" {
			//order by 字段必须放在普通条件后面，所以这里把排序的sql片段先单独拼接出来，等循环结束之后再拼接到sql的最后面。
			mapV := argV.(map[string]interface{})
			if len(mapV) != 0 {
				for factorK, factorV := range mapV {
					orderByStr += factorK + " " + factorV.(string) + ","
				}
			}
			//去掉最后一个逗号
			orderByStr = orderByStr[0 : len(orderByStr)-1]
		} else if argK == "limit" {
			//limit,offset更要放在普通条件和order by后面。所以也是先拼接条件串，最后再拼到主串上。
			limitStr += strconv.Itoa(argV.(int))
		} else if argK == "offset" {
			offsetStr += strconv.Itoa(argV.(int))
		} else {
			//普通参数
			sqlStr = ConcatParamToSql(sqlStr, argK, argV)
		}
	}
	//判断字符串后面有没有多的"AND"，去除sql语句中多余的”AND“，考虑不完善，可能也有OR
	if strings.HasSuffix(strings.Trim(sqlStr, " "), "AND") {
		sqlStr = sqlStr[0 : len(sqlStr)-4]
	}
	//拼接排序,limit,offset语句,
	return ConcatConditionToSql(orderByStr, limitStr, offsetStr, sqlStr)
}

func ConcatWhereToSql(args map[string]interface{}, sqlStr string) string {
	condition := args["where"] != nil || args["order_by"] != nil || args["limit"] != nil || args["offset"] != nil
	//where，order_by这样的参数的数量
	conditionCount := 0
	for k := range args {
		if k == "where" || k == "order_by" || k == "limit" || k == "offset" {
			conditionCount++
		}
	}
	if !condition {
		//全部都是普通参数
		sqlStr += WHERE
	} else {
		if conditionCount != len(args) {
			//二者皆有
			sqlStr += WHERE
		} else {
			if args["where"] != nil {
				sqlStr += WHERE
			}
		}
	}
	return sqlStr
}

//拼接普通参数
func ConcatParamToSql(sqlStr string, argK string, argV interface{}) string {
	switch argV.(type) {
	case int:
		sqlStr += argK + " = " + strconv.Itoa(argV.(int)) + AND
	case string:
		sqlStr += argK + " = " + "'" + argV.(string) + "'" + AND
	}
	return sqlStr
}

//拼接order by，limit,offset的字符串到sql里
func ConcatConditionToSql(order string, limit string, offset string, sqlStr string) string {
	if order != ORDERBY {
		sqlStr += order
	}
	if limit != LIMIT {
		sqlStr += limit
	}
	if offset != OFFSET {
		sqlStr += offset
	}
	return sqlStr
}

func ConcatEqLtGtToSql(sqlStr string, paramK string, condition map[string]interface{}) string {
	if len(condition) != 0 {
		for factorK, factorV := range condition {
			switch factorV.(type) {
			case int:
				if factorK == "gt" {
					sqlStr += paramK + " > " + strconv.Itoa(factorV.(int)) + AND
				} else if factorK == "lt" {
					sqlStr += paramK + " < " + strconv.Itoa(factorV.(int)) + AND
				} else if factorK == "eq" {
					sqlStr += paramK + " = " + strconv.Itoa(factorV.(int)) + AND
				}
			case string:
				if factorK == "gt" {
					sqlStr += paramK + " > " + "'" + factorV.(string) + "'" + AND
				} else if factorK == "lt" {
					sqlStr += paramK + " < " + "'" + factorV.(string) + "'" + AND
				} else if factorK == "eq" {
					sqlStr += paramK + " = " + "'" + factorV.(string) + "'" + AND
				}
			}
		}
	}
	if strings.HasSuffix(strings.Trim(sqlStr, " "), "AND") {
		sqlStr = sqlStr[0 : len(sqlStr)-4]
	}
	return sqlStr
}

//生成插入语句的sql
func GenerateCreateSql(tableName string, args map[string]interface{}) string {
	sql := AddObjectParam(args, tableName)
	return strings.Trim(sql, ";")
}

//对于插入方法传来的对象参数，需要判断是哪个对象。然后生成sql，拼接条件。注意需要递归实现。
func AddObjectParam(args map[string]interface{}, tableName string) string {
	sql := ";INSERT INTO  " + tableName + "("
	value := "VALUES("
	subSql := ""
	for argK, argV := range args {
		switch reflect.TypeOf(argV).Kind() {
		case reflect.Map:
			keys := strings.Split(argK, "_")
			if len(keys) != 0 {
				subSqlStr := ";INSERT INTO " + keys[5] + "("
				subValue := " VALUES( "
				mapData := argV.(map[string]interface{})
				//mapData里既有简单字段，比如id:33,name:"kimmy"。也可能有嵌套的map对象。对于简单字段，直接拼接到insert语句中。
				//而对于对象类型的数据，需要重新再写一个insert语句来插入到不同的表中去。
				for mapK, mapV := range mapData {
					switch mapV.(type) {
					case int:
						subSqlStr += mapK + ","
						subValue += strconv.Itoa(mapV.(int)) + ","
					case string:
						subSqlStr += mapK + ","
						subValue += "'" + mapV.(string) + "'" + ","
					default:
						//这里就是对象类型字段的情况，需要注意在拼接上一个insert语句时，不能在中间拼接。要把下一级递归生成的insert语句拼接在这一层insert语句的前面。
						//此时的mapV也是一个map，递归就在这里进行。insert sql不能中间拼串。
						subSql = AddObjectParam(mapV.(map[string]interface{}), strings.Split(mapK, "_")[5])
					}
				}
				subSqlStr = subSqlStr[0:len(subSqlStr)-1] + ") "
				subValue = subValue[0:len(subValue)-1] + ")"
				subSql += subSqlStr + subValue
			}
		case reflect.Int:
			sql += argK + ","
			value += strconv.Itoa(argV.(int)) + ","
		case reflect.String:
			sql += argK + ","
			value += "'" + argV.(string) + "'" + ","
		}
	}
	sql = sql[0:len(sql)-1] + ") "
	value = value[0:len(value)-1] + ")"
	return sql + value + subSql
}

//生成更新数据的sql
func GenerateUpdateSql(tableName string, args map[string]interface{}) string {
	sqlStr := "UPDATE " + tableName + " SET "
	conditions := WHERE
	for k, v := range args {
		//if k == "id" {
		//	//过滤掉id信息，因为id是判断条件而不是设置参数
		//	continue
		//}
		switch v.(type) {
		case int:
			sqlStr += k + "=" + strconv.Itoa(v.(int)) + ","
		case string:
			sqlStr += k + " = " + "'" + v.(string) + "'" + " ,"
		}
	}
	sqlStr = sqlStr[0 : len(sqlStr)-1]
	//根据where参数拼接插入条件
	whereArg := args["where"]
	if whereArg != nil {
		params := whereArg.(map[string]interface{})
		if len(params) != 0 {
			for paramK, paramV := range params {
				conditionV := paramV.(map[string]interface{})
				conditions = ConcatEqLtGtToSql(conditions, paramK, conditionV)
			}
		}
	}
	return sqlStr + conditions + ";"
}

//生成删除数据sql,要考虑级联的情况
func GenerateDeleteSql(tableName string, args map[string]interface{}, dbModelMaps []map[string]interface{}) string {
	//主表中删除数据
	sql := "DELETE FROM " + tableName + WHERE
	for argK, argV := range args {
		//条件
		if argK == "where" {
			mapV := argV.(map[string]interface{})
			if len(mapV) != 0 {
				for paramK, paramV := range mapV {
					conditionV := paramV.(map[string]interface{})
					sql = ConcatEqLtGtToSql(sql, paramK, conditionV) + ";"
				}
			}
		}
	}
	//考虑级联，同时删除外键表中相关的数据。
	return RecursivelyGenerateDeleteSql(dbModelMaps, sql)
}

//递归地生成删除外键表数据的语句，例如A有多个B，C.B有多个F.C有多个G，。那么总的生成sql包含：Delete A，Delete,B，Delete C，Delete F，Delete G.
func RecursivelyGenerateDeleteSql(dbModelMaps []map[string]interface{}, sqlStr string) string {
	//tableName string,modelTypes []*graphql.Object,
	for index := range dbModelMaps {
		model := dbModelMaps[index]
		for k, v := range model {
			switch reflect.TypeOf(v).Kind() {
			case reflect.Slice, reflect.Array:
				subModel := v.([]map[string]interface{})
				sqlStr += RecursivelyGenerateDeleteSql(subModel, sqlStr)
				kArr := strings.Split(k, "_")
				fromEntity := model[strings.ToLower(kArr[2])]
				sqlPart := " DELETE FROM " + kArr[5] + WHERE + kArr[4] + "="
				switch fromEntity.(type) {
				case int:
					sqlStr += sqlPart + strconv.Itoa(fromEntity.(int)) + ";"
				case string:
					sqlStr += sqlPart + "'" + fromEntity.(string) + "'" + ";"
				}
			}
		}
	}
	return sqlStr
}

//GraphqlHandler 整合进gin框架的处理方法
func GraphqlHandler(schema graphql.Schema) gin.HandlerFunc {
	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: false,
	})
	// 只需要通过Gin简单封装即可
	return func(c *gin.Context) {
		SetIgnoreCase(c.Request)
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// SetIgnoreCase 设置忽略大小写
func SetIgnoreCase(req *http.Request) {
	bodyBytes, _ := ioutil.ReadAll(req.Body)
	type queryParam struct {
		Query string `json:"query"`
	}
	var param queryParam
	json.Unmarshal(bodyBytes, &param)
	queryStr := param.Query
	// fmt.Println(queryStr)
	//匹配所有字符串（带引号的）
	strMatches, err := regexp.Compile("\"[^\"]+\"")
	//匹配所有单元
	allMatches, err := regexp.Compile("\"?\\w+\"?")
	// fmt.Println(err)
	if err != nil {
		return
	}
	strResults := strMatches.FindAllString(queryStr, -1)
	allResults := allMatches.FindAllString(queryStr, -1)
	for _, entityOrField := range allResults {
		isStr := false
		for _, str := range strResults {
			if str == entityOrField {
				//筛选掉字符串的单元
				isStr = true
				break
			}
		}
		if !isStr {
			//将字段和表名称统一转为小写
			compileStr := "(?<!'|\")" + entityOrField + "(?!='|\")"
			replaceCompile, _ := regexp2.Compile(compileStr, 0)
			queryStr, err = replaceCompile.Replace(queryStr, strings.ToLower(entityOrField), 0, len(entityOrField))
		}
	}

	param.Query = queryStr
	newParamByteArr, _ := json.MarshalIndent(param, "", " ")

	//重写查询请求
	buff := bytes.NewBuffer(newParamByteArr)
	req.Body = ioutil.NopCloser(buff)
}
