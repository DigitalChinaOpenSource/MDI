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
	"github.com/gin-gonic/gin"
	"net/http"
)

type OdataHandler struct {}

func (h *OdataHandler) GetEntity(c *gin.Context)  {
	tableName := c.Param("entity")
	//获取路由中表名与其主键,如果不带主键，那么k为表名，v为空串。否则，v为其主键
	tableMap := ti.GetPrimaryKey(tableName)
	condition := ti.GetCondition(c)

	//根据tableMap中是否含有主键，分别执行不同得逻辑，有主键查询单个数据。没有则查询所有。
	for tableName,tablePK :=range tableMap {
		if tablePK == "" {
			//查询所有
			res,jsonResult := ti.GetOdataData(tableName,condition)
			if res.RespondStatus {
				c.JSON(http.StatusOK, jsonResult)
			} else {
				c.JSON(http.StatusOK, gin.H{
					"RespondStatus": res.RespondStatus,
					"Err":           res.Err.Error(),
				})
			}
		}else{
			//根据主键查一个
			res, jsonResult := ti.GetODataById(tableName, tablePK)
			if res.RespondStatus {
				c.JSON(http.StatusOK, jsonResult)
			} else {
				c.JSON(http.StatusOK, map[string]interface{}{
					"RespondStatus": res.RespondStatus,
					"Err":           res.Err.Error(),
				})
			}
		}
		//map中只有一个，第一次循环之后就退出
		break
	}
	return
}

func (h *OdataHandler) GetDataById(c *gin.Context) {
	tableName := c.Param("entity")
	//count := c.Param("$count")
	res, jsonResult:= ti.GetCount(tableName)
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