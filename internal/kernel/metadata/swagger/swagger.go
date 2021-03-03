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

package swagger

type Swagger struct {
	SwaggerVersion string                 `json:"swagger"`
	Info           Info                   `json:"info"`
	Host           string                 `json:"host"`
	BasePath       string                 `json:"basePath"`
	Tags           []Tags                 `json:"tags,omitempty"`
	Schemas        []string               `json:"schemas"`
	Paths          map[string]EntityName  `json:"paths"`
	Definitions    map[string]Definitions `json:"definitions"`
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

type Tags struct {
	Name string `json:"name"`
	Descriptor string `json:"descriptor"`
}

type EntityName struct {
	Get Get `json:"get,omitempty"`
	Post *Post `json:"post,omitempty"`
	Put *Put `json:"put,omitempty"`
	Patch *Patch `json:"patch,omitempty"`
	Delete *Delete `json:"delete,omitempty"`
}

type Get struct {
	Tags []string `json:"tags"`
	Summary string `json:"summary"`
	Definitions string `json:"definitions"`
	OperationId string `json:"operationid"`
	Consumes []string `json:"consumes"`
	Produces []string `json:"produces"`
	Parameters []Parameters `json:"parameters"`
	Responses []Responses `json:"responses"`
}

type Post struct {
	Tags []string `json:"tags"`
	Summary string `json:"summary"`
	Definitions string `json:"definitions"`
	OperationId string `json:"operationid"`
	Consumes []string `json:"consumes"`
	Produces []string `json:"produces"`
	Parameters []Parameters `json:"parameters"`
	Responses []Responses `json:"responses"`
}

type Put struct {
	Tags []string `json:"tags"`
	Summary string `json:"summary"`
	Definitions string `json:"definitions"`
	OperationId string `json:"operationid"`
	Consumes []string `json:"consumes"`
	Produces []string `json:"produces"`
	Parameters []Parameters `json:"parameters"`
	Responses []Responses `json:"responses"`
}

type Patch struct {
	Tags []string `json:"tags"`
	Summary string `json:"summary"`
	Definitions string `json:"definitions"`
	OperationId string `json:"operationid"`
	Consumes []string `json:"consumes"`
	Produces []string `json:"produces"`
	Parameters []Parameters `json:"parameters"`
	Responses []Responses `json:"responses"`
}

type Delete struct {
	Tags []string `json:"tags"`
	Summary string `json:"summary"`
	Definitions string `json:"definitions"`
	OperationId string `json:"operationid"`
	Consumes []string `json:"consumes"`
	Produces []string `json:"produces"`
	Parameters []Parameters `json:"parameters"`
	Responses []Responses `json:"responses"`
}

type Parameters struct {
	Type string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Name string `json:"name,omitempty"`
	In string `json:"in,omitempty"`
	Required bool `json:"required,omitempty"`
	Schema *Schema `json:"schema,omitempty"`
}

type Schema struct {
	Ref string `json:"#ref,omitempty"`
}

type Responses struct {
	Code200 Code200 `json:"200,omitempty"`
	Code400 Code400 `json:"400,omitempty"`
	Code404 Code404 `json:"404,omitempty"`
	Code405 Code405 `json:"405,omitempty"`
}

type Code200 struct {
	Definition string `json:"definition,omitempty"`
}

type Code400 struct {
	Definition string `json:"definition,omitempty"`
}

type Code404 struct {
	Definition string `json:"definition,omitempty"`
}

type Code405 struct {
	Definition string `json:"definition,omitempty"`
}

type Definitions struct {
	Type string `json:"type"`
	Required []string `json:"required"`
	Properties map[string]Field `json:"properties"`
}

type Field struct {
	Type string `json:"type"`
}

type ErrorJson struct {
	SwaggerVersion string                 `json:"swagger"`
	Info           Info                   `json:"info"`
}