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

package tidb

import (
	"dataapi/internal/kernel/metadata/modeling"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
)

/*
QueryAllTables 查询数据库中的所有表
*/
func (e *EnvConverter) queryAllTables() []modeling.Entity {

	result := []modeling.Entity{}
	rows, err := e.Db.Query(fmt.Sprintf("select tidb_table_id,table_name,table_comment from information_schema.tables where table_schema = '%s' and table_type='BASE TABLE' ", e.Connector.DatabaseName) + "and table_name not like 'sys_%'")
	//fmt.Println("---------------")
	for rows.Next() {
		var entity modeling.Entity
		var id int
		var desc sql.NullString
		err = rows.Scan(&id, &entity.SchemaNameAttr, &desc)

		if err != nil {
			panic(err)
		}

		var extra = e.parseDescription(1, entity.SchemaNameAttr)
		if extra != nil {
			entity.DisplayNameAttr = extra.DisplayName
			entity.IsOriginalAttr = extra.IsOriginal
		}
		//fmt.Printf("%d | %s | %s \n", id, value.String, desc.String)

		entity.Fields = e.queryTableFields(entity.SchemaNameAttr)

		entity.Indexes = e.queryTableIndexes(entity.SchemaNameAttr)

		entity.UniqueConstraints = e.queryTableUniqueConstraint(entity.SchemaNameAttr)

		result = append(result, entity)
	}
	return result
}

/*
QueryTableFields 查询表中的字段
*/
func (e *EnvConverter) queryTableFields(table string) []modeling.Field {
	result := []modeling.Field{}
	sqlStr := fmt.Sprintf(`select column_name as ColumnName,ordinal_position as Colorder,
	case is_nullable when 'NO' then 0 else 1 end as CanNull,extra,
	data_type as TypeName, column_comment,
	coalesce(character_maximum_length,numeric_precision,-1) as Length,numeric_scale as Scale,
	column_default as DefaultVal
	from information_schema.columns
	where table_schema='%s' and table_name='%s' order by ordinal_position asc`, e.Connector.DatabaseName, table)
	rows, err := e.Db.Query(sqlStr)
	//fmt.Println("	****************Fields*************")
	for rows.Next() {
		var field modeling.Field
		var num int
		var extra sql.NullString
		var typename sql.NullString
		var defaultValue sql.NullString
		var scale sql.NullInt32
		var length int
		var desc sql.NullString
		err = rows.Scan(&field.SchemaNameAttr, &num, &field.IsNullAttr, &extra, &typename, &desc, &length, &scale, &defaultValue)

		if err != nil {
			panic(err)
		}
		var comment = e.parseDescription(2, table+"_"+field.SchemaNameAttr)
		if comment != nil {
			field.DisplayNameAttr = comment.DisplayName
			field.IsOriginalAttr = comment.IsOriginal
		}

		typeOption := &modeling.TypeOption{}
		if defaultValue.Valid {
			field.DefaultConstraint = &modeling.DefaultConstraint{
				//SchemaNameAttr: field.SchemaNameAttr,
				ValueAttr: defaultValue.String,
			}
		}
		typeOption.AutoIncrementAttr = strings.Contains(extra.String, "auto_increment")
		switch typename.String {
		case "bit":
			field.DataTypeAttr = "boolean"
		case "smallint":
			field.DataTypeAttr = "integer"
			typeOption.LengthAttr = 16
		case "int":
			field.DataTypeAttr = "integer"
			typeOption.LengthAttr = 32
		case "bigint":
			field.DataTypeAttr = "integer"
			typeOption.LengthAttr = 64
		case "decimal":
			typeOption.LengthAttr = length
			typeOption.PrecisionAttr = int(scale.Int32)
			if length == 15 && typeOption.PrecisionAttr == 2 {
				field.DataTypeAttr = "money"
			} else {
				field.DataTypeAttr = "decimal"
			}
		case "datetime":
			field.DataTypeAttr = "datetime"
		case "varchar", "text":
			field.DataTypeAttr = "string"
			typeOption.LengthAttr = length
		}
		field.TypeOption = typeOption

		result = append(result, field)
	}
	rows.Close()
	return result
}

/*
QueryTableIndexes 查询表中的索引项
*/
func (e *EnvConverter) queryTableIndexes(table string) *[]modeling.Index {
	result := []modeling.Index{}
	uniqueFilter := "and not exists(select 0 from information_schema.table_constraints tc where  tc.table_name = ss.table_name and tc.table_schema = ss.table_schema and ss.index_name = tc.constraint_name and constraint_type='UNIQUE')"
	sqlStr := fmt.Sprintf("select index_name from information_schema.statistics ss where table_schema='%s' and table_name='%s' "+uniqueFilter+" group by index_name", e.Connector.DatabaseName, table)
	rows, err := e.Db.Query(sqlStr)
	//fmt.Println("	******************Indexes**********")
	for rows.Next() {
		var index modeling.Index
		err = rows.Scan(&index.SchemaNameAttr)

		if err != nil {
			panic(err)
		}

		index.IsPrimaryAttr = index.SchemaNameAttr == "PRIMARY"

		var comment = e.parseDescription(3, index.SchemaNameAttr)
		if comment != nil {
			index.DisplayNameAttr = comment.DisplayName
		}

		col := []modeling.ColumnDirection{}
		rws, _ := e.Db.Query(fmt.Sprintf("select column_name,collation from information_schema.statistics where table_schema='%s' and table_name='%s' and index_name='%s'", e.Connector.DatabaseName, table, index.SchemaNameAttr))
		for rws.Next() {
			var cd modeling.ColumnDirection
			var collation string
			rws.Scan(&cd.ColumnAttr, &collation)
			cd.DirectionASCAttr = collation == "A"
			col = append(col, cd)
		}
		index.Columns = col
		if index.IsPrimaryAttr {
			index.SchemaNameAttr = ""
			index.DisplayNameAttr = ""
		}
		result = append(result, index)
	}
	rows.Close()
	return &result
}

/*
QueryTableUniqueConstraint 查询表中的唯一性约束
*/
func (e *EnvConverter) queryTableUniqueConstraint(table string) *[]modeling.UniqueConstraint {
	result := []modeling.UniqueConstraint{}
	sqlStr := fmt.Sprintf("select constraint_name from information_schema.table_constraints where table_schema='%s' and table_name='%s' and constraint_type='UNIQUE'", e.Connector.DatabaseName, table)
	rows, err := e.Db.Query(sqlStr)
	//fmt.Println("	****************UniqueConstraint************")
	for rows.Next() {
		var unique modeling.UniqueConstraint
		err = rows.Scan(&unique.SchemaNameAttr)

		if err != nil {
			panic(err)
		}
		var comment = e.parseDescription(4, unique.SchemaNameAttr)
		if comment != nil {
			unique.DisplayNameAttr = comment.DisplayName
		}
		col := []modeling.ColumnDirection{}
		rws, _ := e.Db.Query(fmt.Sprintf("select column_name from information_schema.key_column_usage where table_schema='%s' and table_name='%s' and constraint_name='%s'", e.Connector.DatabaseName, table, unique.SchemaNameAttr))
		for rws.Next() {
			var cd modeling.ColumnDirection
			rws.Scan(&cd.ColumnAttr)
			col = append(col, cd)
		}
		unique.Columns = col
		result = append(result, unique)
	}
	rows.Close()
	return &result
}

/*
QueryAllForeignKeys 查询数据库中的外键
*/
func (e *EnvConverter) queryAllForeignKeys() *[]modeling.ForeignKey {
	result := []modeling.ForeignKey{}
	sqlStr := fmt.Sprintf(`select constraint_name,table_name,column_name,referenced_table_name,referenced_column_name 
		from information_schema.key_column_usage
		where referenced_table_name is not null and table_schema='%s'`, e.Connector.DatabaseName)
	rows, err := e.Db.Query(sqlStr)
	//fmt.Println("******************ForeignKeys**********")
	for rows.Next() {
		var foreign modeling.ForeignKey

		err = rows.Scan(&foreign.SchemaNameAttr, &foreign.ForeignEntityAttr, &foreign.ForeignFieldAttr, &foreign.FromEntityAttr, &foreign.FromFieldAttr)
		if err != nil {
			panic(err)
		}

		var desc string
		e.Db.QueryRow(fmt.Sprintf("select column_comment from information_schema.columns where table_schema='%s' and table_name='%s' and column_name='%s'", e.Connector.DatabaseName, foreign.ForeignEntityAttr, foreign.ForeignFieldAttr)).Scan(&desc)

		var extra = e.parseDescription(5, foreign.SchemaNameAttr)
		if extra != nil && extra.Foreign != nil {
			foreign.DisplayNameAttr = extra.DisplayName
			foreign.ForeignEntityRelation = extra.Foreign.ForeignEntityRelation
			foreign.FromEntityRelation = extra.Foreign.FromEntityRelation
			foreign.CascadeOptionAttr = extra.Foreign.CascadeOptionAttr
		}
		result = append(result, foreign)
	}
	rows.Close()
	return &result
}

/*
ParseDescription 解析描述字段
cate:1-table 2-column 3-index 4-unique 5-fk
*/
func (e *EnvConverter) parseDescription(cate int, name string) *modeling.DescriptionExtra {
	var isOriginal, displayName string
	var extension sql.NullString
	e.Db.QueryRow(fmt.Sprintf("select is_original,display_name,extension from sys_schemainfo where schema_cate=%d and schema_name='%s'", cate, name)).Scan(&isOriginal, &displayName, &extension)
	var res = &modeling.DescriptionExtra{
		IsOriginal:  isOriginal == "1",
		DisplayName: displayName,
	}
	if cate == 5 && extension.Valid {
		var fk modeling.ForeignExtra
		json.Unmarshal([]byte(extension.String), &fk)
		res.Foreign = &fk
	}
	return res
}

/*
ParseDescription 解析描述字段
*/
// func (e *EnvConverter) parseDescription(desc string) *modeling.DescriptionExtra {
// 	if len(desc) == 0 {
// 		return nil
// 	}
// 	var extra modeling.DescriptionExtra
// 	json.Unmarshal([]byte(desc), &extra)
// 	return &extra
// }

/*
SchemaToModelXML 将数据库结构转化为指定格式的XML
*/
func (e *EnvConverter) SchemaToModelXML() string {
	v := &modeling.Model{
		CollationAttr:       "Chinese_PRC_CI_AS",
		ModelingVersionAttr: "1.0",
		OwnerAttr:           e.Owner,
	}

	e.Db = GetEnvDbContext(e.Connector)
	//defer e.Db.Close()

	v.Entities = e.queryAllTables()

	v.ForeignKeys = e.queryAllForeignKeys()

	output, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	//os.Stdout.Write([]byte(xml.Header))
	//os.Stdout.Write(output)
	return xml.Header + string(output)
	// f, err := os.Create("d:/model.xml")
	// f.WriteString(xml.Header)
	// f.Write(output)
	// f.Close()
}
