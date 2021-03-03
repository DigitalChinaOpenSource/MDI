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

package main

import (
	"dataapi/cmd/agent/router"
	"dataapi/cmd/agent/router/handler/dbbase"
	"dataapi/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
	_ "github.com/graphql-go/graphql"
	"net/http"
)

func main() {
	engine := gin.Default()
	engine.Use(middleware.Cors())

	// 初始化服务
	err := initService(engine)
	if err == nil {
		router.Register(engine)

		//initRestfulApi(engine)
	}
	engine.Run(":8081")
}

// initService 初始化服务
func initService(r *gin.Engine) error {
	// 初始化用户数据信息
	err := dbbase.UserDbContext.Refresh()
	// 服务状态
	status := true
	errMsg := "OK"
	if err != nil {
		status = false
		errMsg = err.Error()
	}
	// 访问根，用于测试是否能正常访问
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"RespondStatus": status,
			"Err":           errMsg,
		})
	})
	return err
}