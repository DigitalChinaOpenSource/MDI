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
	"dataapi/cmd/agent/router/handler/ti"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

type RestfulHandler struct{}

func (h *RestfulHandler) GetEntity(c *gin.Context) {
	tableName := c.Param("entity")
	limit := c.Query("limit")
	offset := c.Query("offset")
	orderBy := c.Query("sortby")
	condition := map[string]string{
		"limit":   limit,
		"offset":  offset,
		"orderBy": orderBy,
	}

	res, jsonResult := ti.GetData(tableName, condition)
	if res.RespondStatus {
		c.JSON(http.StatusOK, jsonResult)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"RespondStatus": res.RespondStatus,
			"Err":           res.Err.Error(),
		})
	}
	return
}

func (h *RestfulHandler) PostEntity(c *gin.Context) {
	tableName := c.Param("entity")
	data, _ := ioutil.ReadAll(c.Request.Body)
	jsonStr := string(data)
	var jsonMap []map[string]interface{}

	if err := json.Unmarshal([]byte(jsonStr), &jsonMap); err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{
			"RespondStatus": false,
			"Err":           err.Error(),
		})
	}

	res := ti.CreateData(tableName, jsonMap)

	c.JSON(http.StatusOK, map[string]interface{}{
		"RespondStatus": res.RespondStatus,
		"Err":           res.Err.Error(),
	})
	return
}

func (h *RestfulHandler) GetDataById(c *gin.Context) {
	tableName := c.Param("entity")
	id := c.Param("id")
	res, jsonResult := ti.GetDataById(tableName, id)
	if res.RespondStatus {
		c.JSON(http.StatusOK, jsonResult)
	} else {
		c.JSON(http.StatusOK, map[string]interface{}{
			"RespondStatus": res.RespondStatus,
			"Err":           res.Err.Error(),
		})
	}
	return
}

func (h *RestfulHandler) PutDataById(c *gin.Context) {
	tableName := c.Param("entity")
	id := c.Param("id")
	data, _ := ioutil.ReadAll(c.Request.Body)
	jsonStr := string(data)
	var jsonMap map[string]interface{}

	if err := json.Unmarshal([]byte(jsonStr), &jsonMap); err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{
			"RespondStatus": false,
			"Err":           err.Error(),
		})
	}

	res := ti.UpdateDataById(tableName, id, jsonMap)
	c.JSON(http.StatusOK, map[string]interface{}{
		"RespondStatus": res.RespondStatus,
		"Err":           res.Err.Error(),
	})
	return
}

func (h *RestfulHandler) DeleteDataById(c *gin.Context)  {
	tableName := c.Param("entity")
	id := c.Param("id")
	res := ti.DeleteDataById(tableName, id)

	c.JSON(http.StatusOK, map[string]interface{}{
		"RespondStatus": res.RespondStatus,
		"Err":           res.Err.Error(),
	})
	return
}
