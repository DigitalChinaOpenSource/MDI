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

package router

import (
	"dataapi/cmd/agent/router/handler"
	"dataapi/cmd/agent/router/handler/ti"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func Register(engine *gin.Engine){
	authGroup := engine.Group("")

	data := authGroup.Group("/data")
	{
		handler := &handler.SwaggerHandler{}

		data.GET("/swagger.json",handler.GetSwaggerJson)

		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, handler.GetSwaggerIndexHtml()))
	}

	graphql := authGroup.Group("/api/graphql")
	{
		handler := &handler.GraphQLHandler{}

		graphql.POST("/",ti.GraphqlHandler(handler.GetGraphqlSchema()))
	}

	restful := authGroup.Group("/api/rest")
	{
		handler := &handler.RestfulHandler{}

		restful.GET("/:entity",handler.GetEntity)

		restful.POST("/:entity",handler.PostEntity)

		restful.GET("/:entity/:id",handler.GetDataById)

		restful.PUT("/:entity/:id",handler.PutDataById)

		restful.PATCH("/:entity/:id",handler.PutDataById)

		restful.DELETE("/:entity/:id",handler.DeleteDataById)
	}

	odata := authGroup.Group("/api/odata")
	{
		handler := &handler.OdataHandler{}

		odata.GET("/:entity",handler.GetEntity)

		odata.GET("/:entity/:id",handler.GetDataById)

	}
}