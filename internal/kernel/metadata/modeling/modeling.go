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

package modeling

import (
	"encoding/xml"
)

type Model struct {
	XMLName             xml.Name        `xml:"model"`
	CollationAttr       string          `xml:"collation,attr"`
	ModelingVersionAttr string          `xml:"modelingVersion,attr"`
	OwnerAttr           string          `xml:"owner,attr"`
	Entities            []Entity        `xml:"entities>entity"`
	ForeignKeys         *[]ForeignKey   `xml:"foreignKeys>foreignKey"`
	RenameActions       *[]RenameAction `xml:"renameActions>rename"`
}

type ForeignKey struct {
	SchemaNameAttr        string `xml:"schemaName,attr"`
	DisplayNameAttr       string `xml:"displayName,attr"`
	ForeignFieldAttr      string `xml:"foreignField,attr"`
	ForeignEntityAttr     string `xml:"foreignEntity,attr"`
	FromFieldAttr         string `xml:"fromField,attr"`
	FromEntityAttr        string `xml:"fromEntity,attr"`
	CascadeOptionAttr     string `xml:"cascadeOption,attr"`
	ForeignEntityRelation string `xml:"foreignEntityRelation,attr"`
	FromEntityRelation    string `xml:"fromEntityRelation,attr"`
}

type RenameAction struct {
	TargetAttr    string `xml:"target,attr"`
	IDAttr        string `xml:"id,attr"`
	CreatedOnAttr string `xml:"createdOn,attr"`
	BeforeAttr    string `xml:"before,attr"`
	AfterAttr     string `xml:"after,attr"`
	TableNameAttr string `xml:"tableName,attr,omitempty"`
}

type Entity struct {
	SchemaNameAttr    string              `xml:"schemaName,attr"`
	ClusteredAttr     string              `xml:"clustered,attr"`
	DisplayNameAttr   string              `xml:"displayName,attr"`
	IsOriginalAttr    bool                `xml:"isOriginal,attr"`
	Fields            []Field             `xml:"fields>field"`
	UniqueConstraints *[]UniqueConstraint `xml:"uniqueConstraints>unique"`
	Indexes           *[]Index            `xml:"indexes>index"`
}

type Field struct {
	SchemaNameAttr    string             `xml:"schemaName,attr"`
	IsNullAttr        bool               `xml:"isNull,attr"`
	DisplayNameAttr   string             `xml:"displayName,attr"`
	IsOriginalAttr    bool               `xml:"isOriginal,attr"`
	DataTypeAttr      string             `xml:"dataType,attr"`
	TypeOption        *TypeOption        `xml:"typeOption"`
	DefaultConstraint *DefaultConstraint `xml:"defaultConstraint,omitempty"`
}

type DefaultConstraint struct {
	//SchemaNameAttr string `xml:"schemaName,attr"`
	ValueAttr string `xml:"value,attr"`
}

type TypeOption struct {
	AutoIncrementAttr bool `xml:"autoIncrement,attr,omitempty"`
	LengthAttr        int  `xml:"length,attr,omitempty"`
	PrecisionAttr     int  `xml:"precision,attr,omitempty"`
}

type UniqueConstraint struct {
	SchemaNameAttr  string            `xml:"schemaName,attr"`
	DisplayNameAttr string            `xml:"displayName,attr"`
	Columns         []ColumnDirection `xml:"for"`
}

type ColumnDirection struct {
	ColumnAttr       string `xml:"column,attr"`
	DirectionASCAttr bool   `xml:"directionASC,attr"`
}

type Index struct {
	SchemaNameAttr  string            `xml:"schemaName,attr"`
	DisplayNameAttr string            `xml:"displayName,attr"`
	IsPrimaryAttr   bool              `xml:"isPrimary,attr"`
	Columns         []ColumnDirection `xml:"for"`
}

type DescriptionExtra struct {
	IsOriginal  bool
	DisplayName string
	Foreign     *ForeignExtra
}

type ForeignExtra struct {
	ForeignEntityRelation string
	FromEntityRelation    string
	CascadeOptionAttr     string
}
