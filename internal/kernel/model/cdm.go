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

package model

//CDMEntity ..
type CDMEntity struct {
	EntityID    string `gorm:"primaryKey;column:entity_id"`
	SchemaName  string `gorm:"column:schema_name"`
	DisplayName string `gorm:"column:display_name"`
	IsEditable  bool   `gorm:"column:is_editable"`
	Sort        int    `gorm:"column:sort"`
}

//TableName 数据库表名映射
func (CDMEntity) TableName() string {
	return "cdm_entity"
}

//CDMField ..
type CDMField struct {
	EntityID    string `gorm:"primaryKey;column:field_id"`
	EntityName  string `gorm:"column:entity_name"`
	SchemaName  string `gorm:"column:schema_name"`
	DisplayName string `gorm:"column:display_name"`
	IsPrimary   bool   `gorm:"column:is_primary"`
	IsNullable  bool   `gorm:"column:is_nullable"`
	DataType    string `gorm:"column:data_type"`
	IsAutoIncr  bool   `gorm:"column:is_autoincr"`
	Length      int    `gorm:"column:length"`
	Precision   int    `gorm:"column:precision"`
	IsEditable  bool   `gorm:"column:is_editable"`
	Sort        int    `gorm:"column:sort"`
}

//TableName 数据库表名映射
func (CDMField) TableName() string {
	return "cdm_field"
}
