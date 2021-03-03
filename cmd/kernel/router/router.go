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
	"dataapi/cmd/kernel/router/handler"
	"dataapi/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

//Register ..
func Register(engine *gin.Engine) {

	authGroup := engine.Group("")
	authGroup.Use(middleware.Authenticate())

	sys := authGroup.Group("/mdi/api/system")
	{
		handler := &handler.SysHandler{}

		//获取rules.json
		sys.GET("/rules.json", handler.GetOathkeeperRules)

		//创建一条规则
		sys.POST("/rules/:agentid", handler.CreateOathkeeperRules)

		//删除一条规则
		sys.DELETE("/rules/:agentid", handler.DeleteOathkeeperRules)

		sys.GET("/cdm", handler.GetCDMEntityList)
	}

	proj := authGroup.Group("/mdi/api/project")
	{
		handler := &handler.ProjectHandler{}

		//获取所有项目列表
		proj.GET("/", handler.GetList)

		//获取项目信息
		proj.GET("/g/:id", handler.GetProjectByID)

		//创建项目
		proj.POST("/", handler.CreateProject)

		//获取项目下的环境列表
		//proj.GET("/:id/env", handler.GetProjectEnvs)

		//添加项目关联的用户
		proj.POST("/:id/user", handler.AddProjectUser)

		//发布项目到生产环境
		proj.POST("/:id/publish", handler.Publish)

		//删除项目关联的用户
		proj.PUT("/:id/user", handler.RemoveProjectUser)

		//我关注的项目
		proj.GET("/stars", handler.GetStarList)

		//我关注的项目
		proj.GET("/stars/latest", handler.GetLatestStarList)

		//添加项目关注
		proj.POST("/:id/star", handler.AddProjectStar)

		//删除项目关注
		proj.DELETE("/:id/star", handler.RemoveProjectStar)
	}

	env := authGroup.Group("/mdi/api/env")
	{
		handler := &handler.EnvHandler{}

		//创建环境metadata和画布xml
		env.POST("/", handler.CreateEnv)

		//获取环境的配置
		env.GET("/:id", handler.GetCurrent)

		//删除环境
		env.DELETE("/:id", handler.DeleteEnv)

		//删除关联的用户
		env.POST("/:id/user", handler.RemoveEnvironmentUser)

		//获取环境未发布的metadata
		env.GET("/:id/modeling", handler.GetCurrentModel)

		//保存编辑中的metadata
		env.POST("/:id/modeling", handler.SaveModel)

		//获取环境已发布的metadata
		env.GET("/:id/modeling/published", handler.GetPublishedModel)

		//导出模型
		env.GET("/:id/modeling/export", handler.ExportModel)

		//导入模型
		env.POST("/:id/modeling/import", handler.ImportModel)

		//预发布
		env.PATCH("/:sourceid/modeling/*targetid", handler.PrePublish)

		//发布
		env.PUT("/:sourceid/modeling/*targetid", handler.Publish)

		//复制环境
		// env.POST("/:sourceid/copy/*targetid", handler.CopyEnv)

		// //重置环境
		// env.POST("/:id/reset", handler.ResetEnv)
	}

	user := authGroup.Group("/mdi/api/user")
	{
		handler := &handler.UserHandler{}

		user.GET("/list", handler.GetUserList)

		user.POST("/profile", handler.UpdateUser)

		user.GET("/query/:loginname", handler.GetUserByLoginName)

	}

	token := authGroup.Group("/mdi/api/oauth")
	{
		handler := &handler.TokenHandler{}

		//获取所有客户端
		token.GET("/client", handler.GetClient)

		//根据id获取客户端
		token.GET("/client/:clientId", handler.GetUserClient)

		//创建客户端
		token.POST("/client", handler.CreateClient)

		//删除客户端
		token.DELETE("/client/:clientId", handler.DeleteClient)

		//更新客户端
		token.PATCH("/client/:clientId", handler.UpdateClient)

		//更新access token
		token.POST("/accesstoken", handler.UpdateAccessToken)

	}

	domain := authGroup.Group("/mdi/api/domain")
	{
		handler := &handler.EnvHandler{}

		//获取用户所有的domain
		domain.GET("/", handler.GetAgentDomain)
	}

	authorizer := authGroup.Group("/mdi/api/authorizer")
	{
		handler := &handler.AuthorizerHandler{}

		//获取环境访问控制策略
		authorizer.GET("/:env", handler.GetAccessControlPolicy)

		//更新环境访问控制策略
		authorizer.PUT("/:env", handler.PutAccessControlPolicy)
	}

	//匿名路由处理放在engine下面
	engine.POST("/mdi/api/user/", (&handler.UserHandler{}).SaveUser)
}
