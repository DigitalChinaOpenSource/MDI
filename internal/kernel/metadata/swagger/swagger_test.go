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
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModelToSwaggerJson(t *testing.T) {
	entityNames := map[string]map[string]string{
		"test":{"name":"string","age":"int8"},
	}
	result := ModelToSwaggerJson(entityNames,"127.0.0.1")
	var x = "{\n\t\"swagger\": \"2.0\",\n\t\"info\": {\n\t\t\"title\": \"API\",\n\t\t\"description\": \"API for the current environment\",\n\t\t\"version\": \"1.0.0\"\n\t},\n\t\"host\": \"127.0.0.1\",\n\t\"basePath\": \"/api/rest/\",\n\t\"tags\": [\n\t\t{\n\t\t\t\"name\": \"test\",\n\t\t\t\"descriptor\": \"\"\n\t\t}\n\t],\n\t\"schemas\": [\n\t\t\"http\",\n\t\t\"https\"\n\t],\n\t\"paths\": {\n\t\t\"/test\": {\n\t\t\t\"get\": {\n\t\t\t\t\"tags\": [\n\t\t\t\t\t\"test\"\n\t\t\t\t],\n\t\t\t\t\"summary\": \"\",\n\t\t\t\t\"definitions\": \"\",\n\t\t\t\t\"operationid\": \"\",\n\t\t\t\t\"consumes\": [\n\t\t\t\t\t\"application/json\"\n\t\t\t\t],\n\t\t\t\t\"produces\": [\n\t\t\t\t\t\"application/json\"\n\t\t\t\t],\n\t\t\t\t\"parameters\": [\n\t\t\t\t\t{\n\t\t\t\t\t\t\"type\": \"integer\",\n\t\t\t\t\t\t\"description\": \"指定返回记录的数量\",\n\t\t\t\t\t\t\"name\": \"limit\",\n\t\t\t\t\t\t\"in\": \"query\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"type\": \"integer\",\n\t\t\t\t\t\t\"description\": \"指定返回记录的开始位置\",\n\t\t\t\t\t\t\"name\": \"offset\",\n\t\t\t\t\t\t\"in\": \"query\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"type\": \"string\",\n\t\t\t\t\t\t\"description\": \"返回结果排序\",\n\t\t\t\t\t\t\"name\": \"sortby\",\n\t\t\t\t\t\t\"in\": \"query\"\n\t\t\t\t\t}\n\t\t\t\t],\n\t\t\t\t\"responses\": [\n\t\t\t\t\t{\n\t\t\t\t\t\t\"200\": {\n\t\t\t\t\t\t\t\"definition\": \"successful operation\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"400\": {\n\t\t\t\t\t\t\t\"definition\": \"Invalid tag value\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"404\": {\n\t\t\t\t\t\t\t\"definition\": \"Not found\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"405\": {\n\t\t\t\t\t\t\t\"definition\": \"Invalid input\"\n\t\t\t\t\t\t}\n\t\t\t\t\t}\n\t\t\t\t]\n\t\t\t},\n\t\t\t\"post\": {\n\t\t\t\t\"tags\": [\n\t\t\t\t\t\"test\"\n\t\t\t\t],\n\t\t\t\t\"summary\": \"\",\n\t\t\t\t\"definitions\": \"\",\n\t\t\t\t\"operationid\": \"\",\n\t\t\t\t\"consumes\": [\n\t\t\t\t\t\"application/json\"\n\t\t\t\t],\n\t\t\t\t\"produces\": [\n\t\t\t\t\t\"application/json\"\n\t\t\t\t],\n\t\t\t\t\"parameters\": [\n\t\t\t\t\t{\n\t\t\t\t\t\t\"name\": \"test\",\n\t\t\t\t\t\t\"in\": \"body\",\n\t\t\t\t\t\t\"schema\": {\n\t\t\t\t\t\t\t\"#ref\": \"#/definitions/test\"\n\t\t\t\t\t\t}\n\t\t\t\t\t}\n\t\t\t\t],\n\t\t\t\t\"responses\": [\n\t\t\t\t\t{\n\t\t\t\t\t\t\"200\": {\n\t\t\t\t\t\t\t\"definition\": \"successful operation\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"400\": {\n\t\t\t\t\t\t\t\"definition\": \"Invalid tag value\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"404\": {\n\t\t\t\t\t\t\t\"definition\": \"Not found\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"405\": {\n\t\t\t\t\t\t\t\"definition\": \"Invalid input\"\n\t\t\t\t\t\t}\n\t\t\t\t\t}\n\t\t\t\t]\n\t\t\t}\n\t\t},\n\t\t\"/test/{id}\": {\n\t\t\t\"get\": {\n\t\t\t\t\"tags\": [\n\t\t\t\t\t\"test\"\n\t\t\t\t],\n\t\t\t\t\"summary\": \"\",\n\t\t\t\t\"definitions\": \"\",\n\t\t\t\t\"operationid\": \"\",\n\t\t\t\t\"consumes\": [\n\t\t\t\t\t\"application/json\"\n\t\t\t\t],\n\t\t\t\t\"produces\": [\n\t\t\t\t\t\"application/json\"\n\t\t\t\t],\n\t\t\t\t\"parameters\": [\n\t\t\t\t\t{\n\t\t\t\t\t\t\"type\": \"integer\",\n\t\t\t\t\t\t\"description\": \"记录的id\",\n\t\t\t\t\t\t\"name\": \"id\",\n\t\t\t\t\t\t\"in\": \"path\"\n\t\t\t\t\t}\n\t\t\t\t],\n\t\t\t\t\"responses\": [\n\t\t\t\t\t{\n\t\t\t\t\t\t\"200\": {\n\t\t\t\t\t\t\t\"definition\": \"successful operation\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"400\": {\n\t\t\t\t\t\t\t\"definition\": \"Invalid tag value\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"404\": {\n\t\t\t\t\t\t\t\"definition\": \"Not found\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"405\": {\n\t\t\t\t\t\t\t\"definition\": \"Invalid input\"\n\t\t\t\t\t\t}\n\t\t\t\t\t}\n\t\t\t\t]\n\t\t\t},\n\t\t\t\"put\": {\n\t\t\t\t\"tags\": [\n\t\t\t\t\t\"test\"\n\t\t\t\t],\n\t\t\t\t\"summary\": \"\",\n\t\t\t\t\"definitions\": \"\",\n\t\t\t\t\"operationid\": \"\",\n\t\t\t\t\"consumes\": [\n\t\t\t\t\t\"application/json\"\n\t\t\t\t],\n\t\t\t\t\"produces\": [\n\t\t\t\t\t\"application/json\"\n\t\t\t\t],\n\t\t\t\t\"parameters\": [\n\t\t\t\t\t{\n\t\t\t\t\t\t\"type\": \"integer\",\n\t\t\t\t\t\t\"description\": \"记录的id\",\n\t\t\t\t\t\t\"name\": \"id\",\n\t\t\t\t\t\t\"in\": \"path\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"name\": \"test\",\n\t\t\t\t\t\t\"in\": \"body\",\n\t\t\t\t\t\t\"schema\": {\n\t\t\t\t\t\t\t\"#ref\": \"#/definitions/test\"\n\t\t\t\t\t\t}\n\t\t\t\t\t}\n\t\t\t\t],\n\t\t\t\t\"responses\": [\n\t\t\t\t\t{\n\t\t\t\t\t\t\"200\": {\n\t\t\t\t\t\t\t\"definition\": \"successful operation\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"400\": {\n\t\t\t\t\t\t\t\"definition\": \"Invalid tag value\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"404\": {\n\t\t\t\t\t\t\t\"definition\": \"Not found\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"405\": {\n\t\t\t\t\t\t\t\"definition\": \"Invalid input\"\n\t\t\t\t\t\t}\n\t\t\t\t\t}\n\t\t\t\t]\n\t\t\t},\n\t\t\t\"patch\": {\n\t\t\t\t\"tags\": [\n\t\t\t\t\t\"test\"\n\t\t\t\t],\n\t\t\t\t\"summary\": \"\",\n\t\t\t\t\"definitions\": \"\",\n\t\t\t\t\"operationid\": \"\",\n\t\t\t\t\"consumes\": [\n\t\t\t\t\t\"application/json\"\n\t\t\t\t],\n\t\t\t\t\"produces\": [\n\t\t\t\t\t\"application/json\"\n\t\t\t\t],\n\t\t\t\t\"parameters\": [\n\t\t\t\t\t{\n\t\t\t\t\t\t\"type\": \"integer\",\n\t\t\t\t\t\t\"description\": \"记录的id\",\n\t\t\t\t\t\t\"name\": \"id\",\n\t\t\t\t\t\t\"in\": \"path\"\n\t\t\t\t\t},\n\t\t\t\t\t{\n\t\t\t\t\t\t\"name\": \"test\",\n\t\t\t\t\t\t\"in\": \"body\",\n\t\t\t\t\t\t\"schema\": {\n\t\t\t\t\t\t\t\"#ref\": \"#/definitions/test\"\n\t\t\t\t\t\t}\n\t\t\t\t\t}\n\t\t\t\t],\n\t\t\t\t\"responses\": [\n\t\t\t\t\t{\n\t\t\t\t\t\t\"200\": {\n\t\t\t\t\t\t\t\"definition\": \"successful operation\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"400\": {\n\t\t\t\t\t\t\t\"definition\": \"Invalid tag value\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"404\": {\n\t\t\t\t\t\t\t\"definition\": \"Not found\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"405\": {\n\t\t\t\t\t\t\t\"definition\": \"Invalid input\"\n\t\t\t\t\t\t}\n\t\t\t\t\t}\n\t\t\t\t]\n\t\t\t},\n\t\t\t\"delete\": {\n\t\t\t\t\"tags\": [\n\t\t\t\t\t\"test\"\n\t\t\t\t],\n\t\t\t\t\"summary\": \"\",\n\t\t\t\t\"definitions\": \"\",\n\t\t\t\t\"operationid\": \"\",\n\t\t\t\t\"consumes\": [\n\t\t\t\t\t\"application/json\"\n\t\t\t\t],\n\t\t\t\t\"produces\": [\n\t\t\t\t\t\"application/json\"\n\t\t\t\t],\n\t\t\t\t\"parameters\": [\n\t\t\t\t\t{\n\t\t\t\t\t\t\"type\": \"integer\",\n\t\t\t\t\t\t\"description\": \"记录的id\",\n\t\t\t\t\t\t\"name\": \"id\",\n\t\t\t\t\t\t\"in\": \"path\"\n\t\t\t\t\t}\n\t\t\t\t],\n\t\t\t\t\"responses\": [\n\t\t\t\t\t{\n\t\t\t\t\t\t\"200\": {\n\t\t\t\t\t\t\t\"definition\": \"successful operation\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"400\": {\n\t\t\t\t\t\t\t\"definition\": \"Invalid tag value\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"404\": {\n\t\t\t\t\t\t\t\"definition\": \"Not found\"\n\t\t\t\t\t\t},\n\t\t\t\t\t\t\"405\": {\n\t\t\t\t\t\t\t\"definition\": \"Invalid input\"\n\t\t\t\t\t\t}\n\t\t\t\t\t}\n\t\t\t\t]\n\t\t\t}\n\t\t}\n\t},\n\t\"definitions\": {\n\t\t\"test\": {\n\t\t\t\"type\": \"object\",\n\t\t\t\"required\": null,\n\t\t\t\"properties\": {\n\t\t\t\t\"age\": {\n\t\t\t\t\t\"type\": \"integer\"\n\t\t\t\t},\n\t\t\t\t\"name\": {\n\t\t\t\t\t\"type\": \"string\"\n\t\t\t\t}\n\t\t\t}\n\t\t}\n\t}\n}"
	assert.Equal(t,x,result)
}

func TestErrorSwaggerJson(t *testing.T) {
	err := errors.New("123")
	result := ErrorSwaggerJson(err)
	x := "{\n\t\"swagger\": \"2.0\",\n\t\"info\": {\n\t\t\"title\": \"API\",\n\t\t\"description\": \"Error:123\",\n\t\t\"version\": \"1.0.0\"\n\t}\n}"
	assert.Equal(t,x,result)
}

func TestTypeMapping(t *testing.T) {
	cases := map[string][]string{
		"integer":{"integer","int2", "int4", "int8"},
		"string" :{"varchar", "text", "string", "datetime", "date", "time", "timestamp", "timestamptz"},
		"number": {"decimal", "float4", "float8"},
		"boolean":{"boolean", "bool"},
	}
	for key, value := range cases {
		for _, c := range value{
			t.Run(c, func(t *testing.T) {
				assert.Equal(t,key,TypeMapping(c))
			})
		}
	}
}