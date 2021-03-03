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

package postgres

import (
	"dataapi/internal/kernel/metadata/modeling"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

/*
QueryAllTables 查询数据库中的所有表
*/
func (e *EnvConverter) queryAllTables() []modeling.Entity {

	result := []modeling.Entity{}
	rows, err := e.Db.Query(`select c.oid,c.relname,d.description from pg_class c 
	join pg_namespace n on c.relnamespace=n.oid 
	left join pg_description d on c.oid=d.objoid and d.objsubid=0
	where c.relkind='r' and n.nspname='public' `)
	//fmt.Println("---------------")
	for rows.Next() {
		var entity modeling.Entity
		var id int
		//var value sql.NullString
		var desc sql.NullString
		err = rows.Scan(&id, &entity.SchemaNameAttr, &desc)

		if err != nil {
			panic(err)
		}

		var extra = e.parseDescription(desc.String)
		if extra != nil {
			entity.DisplayNameAttr = extra.DisplayName
			entity.IsOriginalAttr = extra.IsOriginal
		}
		//fmt.Printf("%d | %s | %s \n", id, value.String, desc.String)

		entity.Fields = e.queryTableFields(entity.SchemaNameAttr)

		entity.Indexes = e.queryTableIndexes(id)

		entity.UniqueConstraints = e.queryTableUniqueConstraint(id)

		result = append(result, entity)
	}
	return result
}

/*
QueryTableFields 查询表中的字段
*/
func (e *EnvConverter) queryTableFields(table string) []modeling.Field {
	result := []modeling.Field{}
	sqlStr := `select column_name as ColumnName,ordinal_position as Colorder,
	case is_nullable when 'NO' then 0 else 1 end as CanNull,is_identity as IsIdentity,
	udt_name as TypeName, c.DeText,
	coalesce(character_maximum_length,numeric_precision,-1) as Length,numeric_scale as Scale,
	column_default as DefaultVal
	from information_schema.columns 
	left join (
		select attname,description as DeText from pg_class
		left join pg_attribute pg_attr on pg_attr.attrelid= pg_class.oid
		left join pg_description pg_desc on pg_desc.objoid = pg_attr.attrelid and pg_desc.objsubid=pg_attr.attnum
		where pg_attr.attnum>0 and pg_attr.attrelid=pg_class.oid and pg_class.relname='` + table + `'
	)c on c.attname = information_schema.columns.column_name
	where table_schema='public' and table_name='` + table + `' order by ordinal_position asc`
	rows, err := e.Db.Query(sqlStr)
	//fmt.Println("	****************Fields*************")
	for rows.Next() {
		var field modeling.Field
		var num int
		var identity sql.NullString
		var typename sql.NullString
		var defaultValue sql.NullString
		var scale sql.NullInt32
		var length int
		var desc sql.NullString
		err = rows.Scan(&field.SchemaNameAttr, &num, &field.IsNullAttr, &identity, &typename, &desc, &length, &scale, &defaultValue)

		if err != nil {
			panic(err)
		}
		var extra = e.parseDescription(desc.String)
		if extra != nil {
			field.DisplayNameAttr = extra.DisplayName
			field.IsOriginalAttr = extra.IsOriginal
		}

		typeOption := &modeling.TypeOption{}
		if defaultValue.Valid {
			field.DefaultConstraint = &modeling.DefaultConstraint{
				//SchemaNameAttr: field.SchemaNameAttr,
				ValueAttr: strings.Split(strings.TrimPrefix(defaultValue.String, "'"), "'")[0],
			}
		}
		if identity.String == "YES" {
			typeOption.AutoIncrementAttr = true
		}
		switch typename.String {
		case "bool":
			field.DataTypeAttr = "boolean"
		case "int2", "int4", "int8":
			field.DataTypeAttr = "integer"
			typeOption.LengthAttr = length
		case "numeric":
			field.DataTypeAttr = "decimal"
			typeOption.LengthAttr = length
			typeOption.PrecisionAttr = int(scale.Int32)
		case "money":
			field.DataTypeAttr = "money"
		case "timestamp":
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
func (e *EnvConverter) queryTableIndexes(table int) *[]modeling.Index {
	result := []modeling.Index{}
	sqlStr := `select c.oid,c.relname,i.indkey,i.indisprimary,i.indoption,d.description from pg_class c
	join pg_namespace n on c.relnamespace=n.oid and n.nspname='public'
	join pg_index i on c.oid=i.indexrelid
	left join pg_description d on c.oid=d.objoid 
	where c.relkind='i' and i.indisprimary=i.indisunique and i.indrelid=` + strconv.Itoa(table)
	rows, err := e.Db.Query(sqlStr)
	//fmt.Println("	******************Indexes**********")
	for rows.Next() {
		var index modeling.Index
		var id int
		//var name string
		var fkey string
		//var isunique bool
		//var isprimary bool
		var indoption string
		var desc sql.NullString
		err = rows.Scan(&id, &index.SchemaNameAttr, &fkey, &index.IsPrimaryAttr, &indoption, &desc)

		if err != nil {
			panic(err)
		}
		optionArry := strings.Fields(indoption)
		col := []modeling.ColumnDirection{}
		for i, v := range strings.Fields(fkey) {
			var field string
			e.Db.QueryRow(`select attname from  pg_attribute where attnum=` + v + ` and attrelid=` + strconv.Itoa(table)).Scan(&field)
			cd := modeling.ColumnDirection{ColumnAttr: field}
			switch optionArry[i] {
			case "0":
				cd.DirectionASCAttr = true
			case "1":
				cd.DirectionASCAttr = false
			}
			col = append(col, cd)
		}
		index.Columns = col

		result = append(result, index)
	}
	rows.Close()
	return &result
}

/*
QueryTableUniqueConstraint 查询表中的唯一性约束
*/
func (e *EnvConverter) queryTableUniqueConstraint(table int) *[]modeling.UniqueConstraint {
	result := []modeling.UniqueConstraint{}
	sqlStr := `select c.oid,c.relname,i.indkey,i.indoption,d.description from pg_class c
	join pg_namespace n on c.relnamespace=n.oid and n.nspname='public'
	join pg_index i on c.oid=i.indexrelid
	join pg_constraint ct on c.oid=ct.conindid
	left join pg_description d on c.oid=d.objoid 
	where c.relkind='i' and i.indisprimary!=i.indisunique and i.indrelid=` + strconv.Itoa(table)
	rows, err := e.Db.Query(sqlStr)
	//fmt.Println("	****************UniqueConstraint************")
	for rows.Next() {
		var unique modeling.UniqueConstraint
		var id int
		//var name string
		var ckey string
		var indoption string
		var desc sql.NullString
		err = rows.Scan(&id, &unique.SchemaNameAttr, &ckey, &indoption, &desc)

		if err != nil {
			panic(err)
		}
		optionArry := strings.Fields(indoption)
		col := []modeling.ColumnDirection{}
		for i, v := range strings.Fields(ckey) {
			var field string
			e.Db.QueryRow(`select attname from  pg_attribute where attnum=` + v + ` and attrelid=` + strconv.Itoa(table)).Scan(&field)
			cd := modeling.ColumnDirection{ColumnAttr: field}
			switch optionArry[i] {
			case "0":
				cd.DirectionASCAttr = true
			case "1":
				cd.DirectionASCAttr = false
			}
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
	sqlStr := `select c.oid,c.conname,c.conkey,c.confkey,c.conrelid,c1.relname,c.confrelid,c2.relname,c.confupdtype,d.description from pg_constraint c
		left join pg_class c1 on c.conrelid=c1.oid
		left join pg_class c2 on c.confrelid=c2.oid
		left join pg_description d on c.oid=d.objoid 
		where c.contype='f'	`
	rows, err := e.Db.Query(sqlStr)
	//fmt.Println("******************ForeignKeys**********")
	for rows.Next() {
		var foreign modeling.ForeignKey
		var id int
		//var name string
		var ctable int
		var ckey string
		var ftable int
		var fkey string
		var optype string
		var desc sql.NullString
		err = rows.Scan(&id, &foreign.SchemaNameAttr, &ckey, &fkey, &ctable, &foreign.ForeignEntityAttr, &ftable, &foreign.FromEntityAttr, &optype, &desc)

		if err != nil {
			panic(err)
		}
		switch optype {
		case "c":
			foreign.CascadeOptionAttr = "CASCADE"
		case "n":
			foreign.CascadeOptionAttr = "SET NULL"
		case "a":
			foreign.CascadeOptionAttr = "NO ACTION"
		default:
			foreign.CascadeOptionAttr = "NO ACTION"
		}
		var extra = e.parseForeignExtra(desc.String)
		if extra != nil {
			//foreign.DisplayNameAttr = extra.DisplayName
			foreign.ForeignEntityRelation = extra.ForeignEntityRelation
			foreign.FromEntityRelation = extra.FromEntityRelation
		}
		e.Db.QueryRow(`select attname from  pg_attribute where attnum=` + strings.TrimSuffix(strings.TrimPrefix(ckey, "{"), "}") + ` and attrelid=` + strconv.Itoa(ctable)).Scan(&foreign.ForeignFieldAttr)
		e.Db.QueryRow(`select attname from  pg_attribute where attnum=` + strings.TrimSuffix(strings.TrimPrefix(fkey, "{"), "}") + ` and attrelid=` + strconv.Itoa(ftable)).Scan(&foreign.FromFieldAttr)

		result = append(result, foreign)
	}
	rows.Close()
	return &result
}

/*
ParseDescription 解析描述字段
*/
func (e *EnvConverter) parseForeignExtra(desc string) *modeling.ForeignExtra {
	if len(desc) == 0 {
		return nil
	}
	var extra modeling.ForeignExtra
	json.Unmarshal([]byte(desc), &extra)
	return &extra
}

/*
ParseDescription 解析描述字段
*/
func (e *EnvConverter) parseDescription(desc string) *modeling.DescriptionExtra {
	if len(desc) == 0 {
		return nil
	}
	var extra modeling.DescriptionExtra
	json.Unmarshal([]byte(desc), &extra)
	return &extra
}

/*
SchemaToModelingXML 将数据库结构转化为指定格式的XML
*/
func (e *EnvConverter) SchemaToModelingXML() string {
	v := &modeling.Model{
		CollationAttr:       "Chinese_PRC_CI_AS",
		ModelingVersionAttr: "1.0",
		OwnerAttr:           "heao@dc.com",
	}

	e.Db = GetEnvDbContext(e.Connector)
	//defer e.Db.Close()

	v.Entities = e.queryAllTables()

	v.ForeignKeys = e.queryAllForeignKeys()

	output, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	os.Stdout.Write([]byte(xml.Header))
	os.Stdout.Write(output)
	return xml.Header + string(output)
	// f, err := os.Create("d:/model.xml")
	// f.WriteString(xml.Header)
	// f.Write(output)
	// f.Close()
}
