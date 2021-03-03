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

package migration

import (
	"dataapi/internal/kernel/metadata/modeling"
	"encoding/xml"
)

// Migration ...
type Migration struct {
	XMLName          xml.Name                  `xml:"migration"`
	RemoveForeignKey *[]RemoveForeignKeyAction `xml:"removeForeignKey>action"`
	CreateForeignKey *[]CreateForeignKeyAction `xml:"createForeignKey>action"`
	ChangeForeignKey *[]ChangeForeignKeyAction `xml:"changeForeignKey>action"`
	CreateEntity     *[]CreateEntityAction     `xml:"createEntity>action"`
	ChangeEntity     *[]ChangeEntityAction     `xml:"changeEntity>action"`
	RemoveEntity     *[]RemoveItem             `xml:"removeEntity>action"`
	RenameActions    *[]modeling.RenameAction  `xml:"renameActions>rename"`
}

// RemoveForeignKeyAction ...
type RemoveForeignKeyAction struct {
	ForeignEntityAttr string        `xml:"foreignEntity,attr"`
	SchemaNameAttr    string        `xml:"schemaName,attr"`
	Result            *ActionResult `xml:"result,omitempty"`
}

// CreateForeignKeyAction
type CreateForeignKeyAction struct {
	SchemaNameAttr        string        `xml:"schemaName,attr"`
	DisplayNameAttr       string        `xml:"displayName,attr"`
	ForeignEntityAttr     string        `xml:"foreignEntity,attr"`
	ForeignFieldAttr      string        `xml:"foreignField,attr"`
	FromEntityAttr        string        `xml:"fromEntity,attr"`
	FromFieldAttr         string        `xml:"fromField,attr"`
	CascadeOptionAttr     string        `xml:"cascadeOption,attr"`
	ForeignEntityRelation string        `xml:"foreignEntityRelation,attr"`
	FromEntityRelation    string        `xml:"fromEntityRelation,attr"`
	Result                *ActionResult `xml:"result,omitempty"`
}

// ChangeForeignKey ...
type ChangeForeignKeyAction struct {
	SchemaNameAttr        string        `xml:"schemaName,attr"`
	DisplayNameAttr       string        `xml:"displayName,attr"`
	CascadeOptionAttr     string        `xml:"cascadeOption,attr"`
	ForeignEntityRelation string        `xml:"foreignEntityRelation,attr"`
	FromEntityRelation    string        `xml:"fromEntityRelation,attr"`
	Result                *ActionResult `xml:"result,omitempty"`
}

// CreateEntityAction
type CreateEntityAction struct {
	DisplayNameAttr   string              `xml:"displayName,attr"`
	IsOriginalAttr    bool                `xml:"isOriginal,attr"`
	SchemaNameAttr    string              `xml:"schemaName,attr"`
	ClusteredAttr     string              `xml:"Clustered,attr"`
	Fields            []Field             `xml:"fields>field"`
	UniqueConstraints *[]UniqueConstraint `xml:"uniqueConstraints>unique"`
	Indexes           *[]Index            `xml:"indexes>index"`
	Result            *ActionResult       `xml:"result,omitempty"`
}

// ChangeEntityAction ..
type ChangeEntityAction struct {
	SchemaNameAttr          string              `xml:"schemaName,attr"`
	NewSchemaNameAttr       string              `xml:"newSchemaName,attr"`
	IgnoreExistsDataAttr    bool                `xml:"ignoreExistsData,attr"`
	DisplayNameAttr         string              `xml:"displayName,attr"`
	NewDisplayNameAttr      string              `xml:"newDisplayName,attr"`
	IsOriginalAttr          bool                `xml:"isOriginal,attr"`
	NewFields               *[]Field            `xml:"newFields>field"`
	NewUniqueConstraints    *[]UniqueConstraint `xml:"newUniqueConstraints>unique"`
	NewIndexes              *[]Index            `xml:"newIndexes>index"`
	RemoveFields            *[]RemoveItem       `xml:"removeFields>field"`
	RemoveUniqueConstraints *[]RemoveItem       `xml:"removeUniqueConstraints>unique"`
	RemoveIndexes           *[]RemoveItem       `xml:"removeIndexes>index"`
	ModifyFields            *[]ModifyField      `xml:"modifyFields>field"`
	Result                  *ActionResult       `xml:"result,omitempty"`
}

// ModifyField ...
type ModifyField struct {
	DisplayNameAttr    string                      `xml:"displayName,attr"`
	NewDisplayNameAttr string                      `xml:"newDisplayName,attr"`
	SchemaNameAttr     string                      `xml:"schemaName,attr"`
	NewSchemaNameAttr  string                      `xml:"newSchemaName,attr"`
	IsOriginalAttr     bool                        `xml:"isOriginal,attr"`
	IsNullAttr         bool                        `xml:"isNull,attr"`
	DataTypeAttr       string                      `xml:"dataType,attr"`
	TypeOption         *modeling.TypeOption        `xml:"typeOption"`
	DefaultConstraint  *modeling.DefaultConstraint `xml:"defaultConstraint"`
	Result             *ActionResult               `xml:"result,omitempty"`
}

// RemoveItem ...
type RemoveItem struct {
	SchemaNameAttr string        `xml:"schemaName,attr"`
	Result         *ActionResult `xml:"result,omitempty"`
}

// ActionResult ...
type ActionResult struct {
	SuccessAttr  bool   `xml:"success,attr"`
	DurationAttr uint16 `xml:"duration,attr"`
	Description  string `xml:",innerxml"`
}

type Field struct {
	SchemaNameAttr    string                      `xml:"schemaName,attr"`
	IsNullAttr        bool                        `xml:"isNull,attr"`
	DisplayNameAttr   string                      `xml:"displayName,attr"`
	IsOriginalAttr    bool                        `xml:"isOriginal,attr"`
	DataTypeAttr      string                      `xml:"dataType,attr"`
	TypeOption        *modeling.TypeOption        `xml:"typeOption"`
	DefaultConstraint *modeling.DefaultConstraint `xml:"defaultConstraint,omitempty"`
	Result            *ActionResult               `xml:"result,omitempty"`
}

type UniqueConstraint struct {
	SchemaNameAttr  string                     `xml:"schemaName,attr"`
	DisplayNameAttr string                     `xml:"displayName,attr"`
	Columns         []modeling.ColumnDirection `xml:"for"`
	Result          *ActionResult              `xml:"result,omitempty"`
}

type Index struct {
	SchemaNameAttr  string                     `xml:"schemaName,attr"`
	DisplayNameAttr string                     `xml:"displayName,attr"`
	IsPrimaryAttr   bool                       `xml:"isPrimary,attr"`
	Columns         []modeling.ColumnDirection `xml:"for"`
	Result          *ActionResult              `xml:"result,omitempty"`
}
