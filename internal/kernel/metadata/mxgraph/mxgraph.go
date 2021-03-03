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

package mxgraph

import "encoding/xml"

type MxGraphModel struct {
	XMLName xml.Name `xml:"mxGraphModel"`
	Cells   []MxCell `xml:"root>mxCell"`
}

type MxCell struct {
	ID          string      `xml:"id,attr"`
	Parent      string      `xml:"parent,attr,omitempty"`
	Style       string      `xml:"style,attr,omitempty"`
	Vertex      string      `xml:"vertex,attr,omitempty"`
	Connectable string      `xml:"connectable,attr,omitempty"`
	Edge        string      `xml:"edge,attr,omitempty"`
	Source      string      `xml:"source,attr,omitempty"`
	Target      string      `xml:"target,attr,omitempty"`
	D           *MxCellD    `xml:"D,omitempty"`
	MxGeometry  *MxGeometry `xml:"mxGeometry,omitempty"`
}

type MxCellD struct {
	As    string `xml:"as,attr"`
	Value string `xml:",innerxml"`
}

type MxGeometry struct {
	X           int          `xml:"x,attr,omitempty"`
	Y           int          `xml:"y,attr,omitempty"`
	Width       int          `xml:"width,attr,omitempty"`
	Height      int          `xml:"height,attr,omitempty"`
	As          string       `xml:"as,attr,omitempty"`
	Relative    string       `xml:"relative,attr,omitempty"`
	MxRectangle *MxRectangle `xml:"mxRectangle,omitempty"`
}

type MxRectangle struct {
	Width  int    `xml:"width,attr,omitempty"`
	Height int    `xml:"height,attr,omitempty"`
	As     string `xml:"as,attr,omitempty"`
}

type MxCellValue struct {
	SchemaName        string            `json:"schemaName"`
	DisplayName       string            `json:"displayName"`
	IsOriginal        bool              `json:"isOriginal"`
	IsNull            bool              `json:"isNull,omitempty"`
	DataType          string            `json:"dataType,omitempty"`
	PrimaryKey        bool              `json:"primaryKey,omitempty"`
	Unique            bool              `json:"unique,omitempty"`
	Clustered         string            `json:"clustered,omitempty"`
	StringOption      *MxCellTypeOption `json:"stringOption"`
	IntegerOption     *MxCellTypeOption `json:"integerOption"`
	DecimalOption     *MxCellTypeOption `json:"decimalOption"`
	UniqueConstraints []MxCellIndex     `json:"uniqueConstraints"`
	Indexes           []MxCellIndex     `json:"indexes"`
}

type MxCellTypeOption struct {
	AutoIncrement bool `json:"autoIncrement,omitempty"`
	Length        int  `json:"length,omitempty"`
	Precision     int  `json:"precision,omitempty"`
}

type MxCellIndex struct {
	SchemaName  string         `json:"schemaName,omitempty"`
	DisplayName string         `json:"displayName,omitempty"`
	IsPrimary   bool           `json:"isPrimary,omitempty"`
	Columns     []MxCellColumn `json:"columns"`
}
type MxCellColumn struct {
	Column       string `json:"column"`
	DirectionASC bool   `json:"directionASC,omitempty"`
}
