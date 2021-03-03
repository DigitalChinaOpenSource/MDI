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
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
)

type SwaggerHandler struct {}

func (h *SwaggerHandler)GetSwaggerJson(c *gin.Context)  {
	jsonStr := ti.GetSwaggerJson()
	c.String(http.StatusOK, jsonStr)
}

func (h *SwaggerHandler)GetSwaggerIndexHtml() func(c *ginSwagger.Config) {
	url := ginSwagger.URL("https://" + ti.GetHostUrl() + "/data/swagger.json") // The url pointing to API definition
	return url
}