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

package handler

import (
	"bytes"
	"dataapi/cmd/kernel/utils"
	"dataapi/internal/kernel/model"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthorizerHandler struct {}

//获取环境访问策略
func (eventh *AuthorizerHandler) GetAccessControlPolicy(c *gin.Context) {
	env := c.Param("env")
	data := model.EnvAccessControlPolicy{}
	result := [3]model.AuthorizerRoleJson{}

	statusCode,content := utils.SendKetoRequest("roles/"+env+"-read","GET",nil)
	if statusCode!=400 {
		err := json.Unmarshal([]byte(content),&result[0])
		if err != nil {
			c.JSON(http.StatusBadRequest,gin.H{"error:":err.Error()})
		}
		data.Read = result[0].Members
	}

	statusCode,content = utils.SendKetoRequest("roles/"+env+"-write","GET",nil)
	if statusCode!=400 {
		err := json.Unmarshal([]byte(content),&result[1])
		if err != nil {
			c.JSON(http.StatusBadRequest,gin.H{"error:":err.Error()})
		}
		data.Write = result[1].Members
	}

	statusCode,content = utils.SendKetoRequest("roles/"+env+"-modify","GET",nil)
	if statusCode!=400 {
		err := json.Unmarshal([]byte(content),&result[2])
		if err != nil {
			c.JSON(http.StatusBadRequest,gin.H{"error:":err.Error()})
		}
		data.Modify = result[2].Members
	}
	c.JSON(http.StatusOK,data)
}

//更新环境访问策略
func (eventh *AuthorizerHandler) PutAccessControlPolicy(c *gin.Context) {
	data := model.EnvAccessControlPolicy{}
	env := c.Param("env")
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		authorizer := [3]model.AuthorizerRoleJson{
			{
				Id:         env+"-read",
				Descriptor: "",
				Members:    data.Read,
			},
			{
				Id:         env+"-write",
				Descriptor: "",
				Members:    data.Write,
			},
			{
				Id:         env+"-modify",
				Descriptor: "",
				Members:    data.Modify,
			},
		}

		reader, _ := json.Marshal(authorizer[0])
		statusCode,content := utils.SendKetoRequest("roles","PUT",bytes.NewBuffer(reader))
		if statusCode == 400 {
			c.JSON(http.StatusBadRequest,content)
		}
		reader, _ = json.Marshal(authorizer[1])
		statusCode,content = utils.SendKetoRequest("roles","PUT",bytes.NewBuffer(reader))
		if statusCode == 400 {
			c.JSON(http.StatusBadRequest,content)
		}
		reader, _ = json.Marshal(authorizer[2])
		statusCode,content = utils.SendKetoRequest("roles","PUT",bytes.NewBuffer(reader))
		if statusCode == 400 {
			c.JSON(http.StatusBadRequest,content)
		}
		c.JSON(http.StatusOK,"success")
	}
}
