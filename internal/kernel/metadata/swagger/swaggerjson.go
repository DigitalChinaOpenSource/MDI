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

import (
	"encoding/json"
)

const (
	CONTENT_TYPE = "application/json"
	DEFINITIONS = "#/definitions/"
)
func ModelToSwaggerJson(entityNames map[string]map[string]string ,host string)  string{
	model := Swagger{
		SwaggerVersion: "2.0",
		Info: Info{
			Title: "API",
			Description: "API for the current environment",
			Version: "1.0.0",
		},
		Host: host,
		BasePath:"/api/rest/",
		Tags: []Tags{
		},
		Schemas: []string{
			"http",
			"https",
		},
		Paths: map[string]EntityName{},
		Definitions: map[string]Definitions{},
	}

	paths := map[string]EntityName{}
	definitions := map[string]Definitions{}

	responses := []Responses{
		{
			Code200: Code200{
				Definition: "successful operation",
			},
			Code400: Code400{
				Definition: "Invalid tag value",
			},
			Code404: Code404{
				Definition: "Not found",
			},
			Code405: Code405{
				Definition: "Invalid input",
			},
		},
	}
	for entityName,field:= range entityNames {

		fields := map[string]Field{}

		model.Tags = append(model.Tags, Tags{Name: entityName,Descriptor: ""})
		paths["/"+entityName]= EntityName{
			Get: Get{
				Tags:        []string{entityName},
				Summary:     "",
				Definitions: "",
				OperationId: "",
				Consumes:    []string{CONTENT_TYPE},
				Produces:    []string{CONTENT_TYPE},
				Parameters:  []Parameters{
					{Type: "integer",Description: "指定返回记录的数量",Name: "limit",In: "query",Required: false},
					{Type: "integer",Description: "指定返回记录的开始位置",Name: "offset",In: "query",Required: false},
					{Type: "string",Description: "返回结果排序",Name: "sortby",In: "query",Required: false},
				},
				Responses: responses,
			},
			Post: &Post{
				Tags:        []string{entityName},
				Summary:     "",
				Definitions: "",
				OperationId: "",
				Consumes:    []string{CONTENT_TYPE},
				Produces:    []string{CONTENT_TYPE},
				Parameters:  []Parameters{
					{
						In:          "body",
						Name:        entityName,
						Description: "",
						Schema:      &Schema{Ref: DEFINITIONS+ entityName},
					},
				},
				Responses: responses,
			},
		}
		paths["/"+entityName+"/{id}"]=EntityName{
			Get: Get{
				Tags:        []string{entityName},
				Summary:     "",
				Definitions: "",
				OperationId: "",
				Consumes:    []string{CONTENT_TYPE},
				Produces:    []string{CONTENT_TYPE},
				Parameters:  []Parameters{
					{Type: "integer",Description: "记录的id",Name: "id",In: "path",Required: false},
				},
				Responses: responses,
			},
			Put: &Put{
				Tags:        []string{entityName},
				Summary:     "",
				Definitions: "",
				OperationId: "",
				Consumes:    []string{CONTENT_TYPE},
				Produces:    []string{CONTENT_TYPE},
				Parameters:  []Parameters{
					{Type: "integer",Description: "记录的id",Name: "id",In: "path",Required: false},
					{
						In:          "body",
						Name:        entityName,
						Description: "",
						Schema:      &Schema{Ref: DEFINITIONS+ entityName},
					},
				},
				Responses: responses,
			},
			Patch: &Patch{
				Tags:        []string{entityName},
				Summary:     "",
				Definitions: "",
				OperationId: "",
				Consumes:    []string{CONTENT_TYPE},
				Produces:    []string{CONTENT_TYPE},
				Parameters:  []Parameters{
					{Type: "integer",Description: "记录的id",Name: "id",In: "path",Required: false},
					{
						In:          "body",
						Name:        entityName,
						Description: "",
						Schema:      &Schema{Ref: DEFINITIONS+ entityName},
					},
				},
				Responses: responses,
			},
			Delete: &Delete{
				Tags:        []string{entityName},
				Summary:     "",
				Definitions: "",
				OperationId: "",
				Consumes:    []string{CONTENT_TYPE},
				Produces:    []string{CONTENT_TYPE},
				Parameters:  []Parameters{
					{Type: "integer",Description: "记录的id",Name: "id",In: "path",Required: false},
				},
				Responses: responses,
			},
		}

		for fieldName, fieldType := range field{
			fields[fieldName] = Field{
				Type: TypeMapping(fieldType),
			}
		}

		definitions[entityName] = Definitions{
			Type:       "object",
			Required:   nil,
			Properties: fields,
		}
	}

	model.Paths=paths
	model.Definitions = definitions

	data, err := json.MarshalIndent(model, "", "	")
	if err != nil {
		return ErrorSwaggerJson(err)
	}
	return string(data)
}

func ErrorSwaggerJson(err error)  string{
	if err != nil {
		model := ErrorJson{
			SwaggerVersion: "2.0",
			Info: Info{
				Title: "API",
				Description: "Error:"+err.Error(),
				Version: "1.0.0",
			},
		}
		data, err := json.MarshalIndent(model, "", "	")
		if err != nil {
			return "Json Marshaling Failed"
		}
		return string(data)
	}
	return "The error does not exist"
}

func TypeMapping(FieldType string) string{
	switch FieldType {
	case "integer","int2", "int4", "int8": return "integer"
	case "varchar", "text", "string","datetime", "date", "time", "timestamp", "timestamptz": return "string"
	case "decimal","float4", "float8": return "number"
	case "boolean", "bool": return "boolean"
	default:
		return "string"
	}
}
