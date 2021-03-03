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

package utils

import (
	"bytes"
	"dataapi/internal/kernel/model"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	Url = "http://mdi-keto.dev.wh.digitalchina.com/engines/acp/ory/exact/"
)

//对keto服务器发起请求
func SendKetoRequest(route string ,method string ,payload io.Reader) (statusCode int, content string) {
	url := Url+route
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return 400,err.Error()
	}

	httpclient:=&http.Client{}
	response, resErr := httpclient.Do(req)
	if resErr != nil {
		return 400,resErr.Error()
	}

	defer response.Body.Close()
	body,bodyErr := ioutil.ReadAll(response.Body)
	if bodyErr != nil {
		return 400,bodyErr.Error()
	}
	return 200,string(body)
}

//创建初始策略
func CreateAuthorizerPolicy(agent string) {
	authorizer := [3]model.AuthorizerPolicyJson{
		{
			Actions:     []string{"read"},
			Conditions:  nil,
			Description: "",
			Effect:      "allow",
			Id:          agent+"-read",
			Resources:   []string{agent},
			Subjects:    []string{agent+"-read"},
		},
		{
			Actions:     []string{"write"},
			Conditions:  nil,
			Description: "",
			Effect:      "allow",
			Id:          agent+"-write",
			Resources:   []string{agent},
			Subjects:    []string{agent+"-write"},
		},
		{
			Actions:     []string{"modify"},
			Conditions:  nil,
			Description: "",
			Effect:      "allow",
			Id:          agent+"-modify",
			Resources:   []string{agent},
			Subjects:    []string{agent+"-modify"},
		},
	}

	reader, _ := json.Marshal(authorizer[0])
	SendKetoRequest("policies","PUT",bytes.NewBuffer(reader))
	reader, _ = json.Marshal(authorizer[1])
	SendKetoRequest("policies","PUT",bytes.NewBuffer(reader))
	reader, _ = json.Marshal(authorizer[2])
	SendKetoRequest("policies","PUT",bytes.NewBuffer(reader))
	return
}

//创建初始角色
func CreateAuthorizerRole(agent string) {
	authorizer := [3]model.AuthorizerRoleJson{
		{
			Id:         agent+"-read",
			Descriptor: "",
			Members:    nil,
		},
		{
			Id:         agent+"-write",
			Descriptor: "",
			Members:    nil,
		},
		{
			Id:         agent+"-modify",
			Descriptor: "",
			Members:    nil,
		},
	}
	reader, _ := json.Marshal(authorizer[0])
	SendKetoRequest("roles","PUT",bytes.NewBuffer(reader))
	reader, _ = json.Marshal(authorizer[1])
	SendKetoRequest("roles","PUT",bytes.NewBuffer(reader))
	reader, _ = json.Marshal(authorizer[2])
	SendKetoRequest("roles","PUT",bytes.NewBuffer(reader))
	return
}

//删除环境策略
func DeleteAuthorizerPolicy(agent string) {
	SendKetoRequest("policies/"+agent+"-read","DELETE",nil)
	SendKetoRequest("policies/"+agent+"-write","DELETE",nil)
	SendKetoRequest("policies/"+agent+"-modify","DELETE",nil)
	return
}

//删除环境角色
func DeleteAuthorizerRole(agent string) {
	SendKetoRequest("roles/"+agent+"-read","DELETE",nil)
	SendKetoRequest("roles/"+agent+"-write","DELETE",nil)
	SendKetoRequest("roles/"+agent+"-modify","DELETE",nil)
	return
}
