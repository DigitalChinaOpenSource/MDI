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
	"dataapi/internal/kernel/db"
	"dataapi/internal/kernel/model"
	"dataapi/internal/pkg/utils"
	"strconv"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

//ProjectHandler 项目相关的处理业务
type ProjectHandler struct{}

//GetList 查询登录用户的项目列表
func (h *ProjectHandler) GetList(c *gin.Context) {
	sdb := db.GetSysDbContext()
	user := GetCurrentLoginUser(c)

	query := sdb.Table("project").Distinct("project.project_id").
		Joins("join project_user on project_user.project_id = project.project_id and project_user.user_login_name = ?", user).
		Joins("left join user_profile on project.creator = user_profile.login_name").
		Joins("left join project_star on project.project_id = project_star.project_id and project_star.user_login_name = ?", user).
		Order("project.create_on desc")
	dest := "project.*,user_profile.display_name,user_profile.avatar,project_star.project_star_id"
	result := utils.Paginate(c, query, dest)

	c.JSON(http.StatusOK, ResponseResult(true, "查询成功", result))
}

//GetProjectByID 根据id查询项目信息
func (h *ProjectHandler) GetProjectByID(c *gin.Context) {
	id := c.Param("id")
	sdb := db.GetSysDbContext()
	var proj model.Project
	result := sdb.Where(`project_id = ?`, id).First(&proj)
	if result.Error == nil {
		data := make(map[string]interface{})
		data["Project"] = proj
		data["Managers"] = h.getProjectUsers(id, 1)
		data["Members"] = h.getProjectUsers(id, 0)
		data["Envs"] = (&EnvHandler{}).GetProjectEnvs(id)
		c.JSON(http.StatusOK, ResponseResult(true, "查询成功", data))
		return
	}
	c.JSON(http.StatusBadRequest, ResponseResult(false, "查询失败", nil))
}

func (h *ProjectHandler) getProjectUsers(id string, isowner int) []model.UserProfile {
	sdb := db.GetSysDbContext()
	var users []model.UserProfile

	sdb.Model(model.ProjectUser{}).Select("user_profile.*").
		Joins("join user_profile on user_profile.login_name = project_user.user_login_name and project_user.project_id = ? and project_user.is_project_owner = ?", id, isowner).
		Find(&users)
	return users
}

//CreateProject 创建一个项目
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	sdb := db.GetSysDbContext()
	type body struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Managers    []string `json:"managers"`
		Members     []string `json:"members"`
		Icon        string   `json:"icon"`
	}
	creator := GetCurrentLoginUser(c)
	var p body
	if c.BindJSON(&p) == nil {
		//保存项目信息
		pro := model.Project{
			ProjectID:   utils.GetUUID(),
			Name:        p.Name,
			Description: p.Description,
			Icon:        p.Icon,
			Creator:     creator,
			CreateOn:    utils.Time(time.Now()),
		}
		res := sdb.Create(&pro)
		if res.Error == nil {
			//保存项目负责人和成员
			var members []model.ProjectUser
			if len(p.Managers) > 0 {
				for _, m := range p.Managers {
					members = append(members, model.ProjectUser{
						ProjectID:      pro.ProjectID,
						UserLoginName:  m,
						IsProjectOwner: true,
					})
				}
			}
			if len(p.Members) > 0 {
				for _, m := range p.Members {
					members = append(members, model.ProjectUser{
						ProjectID:      pro.ProjectID,
						UserLoginName:  m,
						IsProjectOwner: false,
					})
				}
			}
			sdb.Create(&members)
			//createDbInstance(pro.ProjectID)
			//创建默认环境
			createDefaultEnvs(pro.ProjectID, creator)
			c.JSON(http.StatusOK, ResponseResult(true, "创建成功", pro.ProjectID))
			return
		}
	}
	c.JSON(http.StatusBadRequest, ResponseResult(false, "创建失败", nil))
}

//Publish 发布项目从测试环境到生产环境
//原来设计的从开发环境合并到测试环境流程有点问题，因为可能出现合并冲突的问题，
//使用DevOps接口来实现的话没办法实现异常处理，所以这里只做成单向合并，也就是从测试到生产一对一模式，
//可以选择从页面上手动从开发合并到测试（或者导入模型），测试到生产可以手动有也可以自动
func (h *ProjectHandler) Publish(c *gin.Context) {
	id := c.Param("id")
	sdb := db.GetSysDbContext()
	var tenv model.Environment
	result := sdb.Where(`project_id = ? and type = 1`, id).First(&tenv)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, ResponseResult(false, "无法识别的项目", nil))
		return
	}
	var penv model.Environment
	result = sdb.Where(`project_id = ? and type = 2`, id).First(&penv)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, ResponseResult(false, "无法识别的项目", nil))
		return
	}
	sdb.Model(&penv).Updates(map[string]interface{}{"metadata_current": tenv.MetadataPublished})
	cvt := (&EnvHandler{}).SubmitMigration(penv.EnvironmentID, "")
	c.JSON(http.StatusOK, gin.H{"successed": cvt.SuccessedCnt, "failed": cvt.FailedCnt})
}

func (h *ProjectHandler) currentUserIsOwner(id string, c *gin.Context) bool {
	var count int64
	db.GetSysDbContext().Model(&model.ProjectUser{}).Where("project_id = ? and user_login_name = ? and is_project_owner = 1", id, GetCurrentLoginUser(c)).Count(&count)
	return count > 0
}

// AddProjectUser 添加项目负责人或参与成员
func (h *ProjectHandler) AddProjectUser(c *gin.Context) {
	id := c.Param("id")
	if h.currentUserIsOwner(id, c) == false {
		c.JSON(http.StatusBadRequest, ResponseResult(false, "没有修改权限", nil))
		return
	}
	type AddUser struct {
		IsOwner   bool     `json:"is_owner"`
		LoginName []string `json:"loginname"`
	}
	var user AddUser
	c.BindJSON(&user)
	var memers []model.ProjectUser
	if len(user.LoginName) > 0 {
		for _, m := range user.LoginName {
			memers = append(memers, model.ProjectUser{
				ProjectID:      id,
				UserLoginName:  m,
				IsProjectOwner: user.IsOwner,
			})
		}
	}
	result := db.GetSysDbContext().Create(&memers)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, ResponseResult(false, "添加失败", nil))
		return
	}
	c.JSON(http.StatusOK, ResponseResult(true, "添加成功", nil))
}

//RemoveProjectUser 移除项目负责人或参与成员
func (h *ProjectHandler) RemoveProjectUser(c *gin.Context) {
	id := c.Param("id")
	if h.currentUserIsOwner(id, c) == false {
		c.JSON(http.StatusBadRequest, ResponseResult(false, "没有修改权限", nil))
		return
	}
	type RemoveUser struct {
		LoginName string `json:"loginname"`
	}
	var user RemoveUser
	c.BindJSON(&user)
	result := db.GetSysDbContext().Where("project_id = ? and user_login_name = ?", id, user.LoginName).Delete(&model.ProjectUser{})
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, ResponseResult(false, "删除失败", nil))
		return
	}
	c.JSON(http.StatusOK, ResponseResult(true, "删除成功", nil))
}

// GetLatestStarList 查询我最近关注的项目
func (h *ProjectHandler) GetLatestStarList(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 6
	}
	sdb := db.GetSysDbContext()
	var data []map[string]interface{}

	result := sdb.Table("project").Select("project.project_id,project.name").
		Joins("join project_star on project_star.project_id = project.project_id and project_star.user_login_name = ?", GetCurrentLoginUser(c)).
		Order("project_star.create_on desc").
		Limit(limit).
		Find(&data)
	if result.Error == nil {
		c.JSON(http.StatusOK, ResponseResult(true, "查询成功", data))
		return
	}
	c.JSON(http.StatusBadRequest, ResponseResult(false, "查询失败", nil))
}

// GetStarList 查询我已关注的项目
func (h *ProjectHandler) GetStarList(c *gin.Context) {
	user := GetCurrentLoginUser(c)

	sdb := db.GetSysDbContext()
	query := sdb.Table("project").Distinct("project.project_id").
		Joins("join project_user on project_user.project_id = project.project_id and project_user.user_login_name = ?", user).
		Joins("join project_star on project.project_id = project_star.project_id and project_star.user_login_name = ?", user).
		Joins("left join user_profile on project.creator = user_profile.login_name").
		Order("project_star.create_on desc")
	dest := "project.*,user_profile.display_name,user_profile.avatar,project_star.project_star_id"
	result := utils.Paginate(c, query, dest)

	c.JSON(http.StatusOK, ResponseResult(true, "查询成功", result))
}

// AddProjectStar 关注项目
func (h *ProjectHandler) AddProjectStar(c *gin.Context) {
	star := model.ProjectStar{
		ProjectID:     c.Param("id"),
		UserLoginName: GetCurrentLoginUser(c),
		CreateOn:      utils.Time(time.Now()),
	}
	sdb := db.GetSysDbContext()
	var count int64
	sdb.Model(&model.ProjectStar{}).Where("project_id = ? and user_login_name = ?", star.ProjectID, star.UserLoginName).Count(&count)
	if count > 0 {
		c.JSON(http.StatusOK, ResponseResult(true, "关注成功", nil))
		return
	}
	result := sdb.Create(&star)
	if result.Error == nil {
		c.JSON(http.StatusOK, ResponseResult(true, "关注成功", nil))
		return
	}
	c.JSON(http.StatusBadRequest, ResponseResult(false, "关注失败", nil))
}

// RemoveProjectStar 取消关注
func (h *ProjectHandler) RemoveProjectStar(c *gin.Context) {
	result := db.GetSysDbContext().Where("project_id = ? and user_login_name = ?", c.Param("id"), GetCurrentLoginUser(c)).Delete(&model.ProjectStar{})
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, ResponseResult(false, "取消失败", nil))
		return
	}
	c.JSON(http.StatusOK, ResponseResult(true, "取消成功", nil))
}

//在rancher中创建项目的数据库实例
func createDbInstance(pid string) {
	// payload := strings.NewReader(`{
	// 	"appRevisionId": "",
	// 	"externalId": "catalog://?catalog=alibaba-app-hub&template=postgresql-ha&version=1.4.3",
	// 	"multiClusterAppId": "",
	// 	"name": "` + pid + `",
	// 	"namespaceId": "",
	// 	"projectId": "` + utils.RancherProjectId + `",
	// 	"prune": false,
	// 	"targetNamespace": "mdi-agent-db",
	// 	"timeout": 300,
	// 	"wait": false,
	// 	"answers": {
	// 	"pgpoolImage.debug": "true",
	// 	"postgresql.password": "111111",
	// 	"postgresql.replicaCount": "3",
	// 	"postgresql.repmgrPassword": "111111"
	// 	}
	// 	}`)
	// code, content := utils.SendRancherRequest("POST", "/apps", payload)
	// if code != 400 {
	// 	fmt.Println(content)
	// }
}

//给项目创建默认的环境
func createDefaultEnvs(pid string, owner string) {
	CreateEnv(model.Environment{
		ProjectID:         pid,
		Type:              2, //生产环境
		Owner:             owner,
		MetadataCurrent:   "",
		MetadataPublished: "",
	})
	CreateEnv(model.Environment{
		ProjectID:         pid,
		Type:              1, //测试环境
		Owner:             owner,
		MetadataCurrent:   "",
		MetadataPublished: "",
	})
}

//ProjectEnvView ..
type ProjectEnvView struct {
	ID     string
	Name   string
	Type   uint16
	Domain string
	Users  []model.UserProfile
}

//AgentDomainView ..
type AgentDomainView struct {
	DisplayName string
	AgentDomain string
}
