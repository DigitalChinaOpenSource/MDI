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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

//UserHandler 用户相关的处理业务
type UserHandler struct{}

//GetUserByLoginName 使用登录名查询用户
func (h *UserHandler) GetUserByLoginName(c *gin.Context) {
	loginName := c.Param("loginname")
	sdb := db.GetSysDbContext()
	var user model.UserProfile
	result := sdb.Where("login_name = ?", loginName).First(&user)
	if result.Error == nil {
		c.JSON(http.StatusOK, ResponseResult(true, "查询成功", user))
	} else {
		c.JSON(http.StatusBadRequest, ResponseResult(false, "查询失败", nil))
	}
}

//GetUserList 查询所有的用户列表
func (h *UserHandler) GetUserList(c *gin.Context) {
	key := c.Query("key")
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 50
	}
	sdb := db.GetSysDbContext()
	var users []model.UserProfile
	query := sdb.Model(model.UserProfile{})
	if len(key) > 0 {
		ws := "%" + key + "%"
		query = query.Where("login_name like ?", ws).Or("display_name like ?", ws)
	}
	result := query.Limit(limit).Find(&users)
	if result.Error == nil {
		c.JSON(http.StatusOK, ResponseResult(true, "查询成功", users))
	} else {
		c.JSON(http.StatusBadRequest, ResponseResult(false, "查询失败", nil))
	}
}

//UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
	type user struct {
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	}
	var u user
	if c.BindJSON(&u) == nil {
		sdb := db.GetSysDbContext()
		result := sdb.Model(&model.UserProfile{}).Where("login_name = ?", GetCurrentLoginUser(c)).Updates(map[string]interface{}{"display_name": u.Name, "avatar": u.Avatar})
		if result.Error == nil {
			c.JSON(http.StatusOK, ResponseResult(true, "更新成功", nil))
			return
		}
	}
	c.JSON(http.StatusBadRequest, ResponseResult(false, "更新失败", nil))
}

func (h *UserHandler) SaveUser(c *gin.Context) {
	type userProfile struct {
		user string `json:"user"`
		name string `json:"name"`
	}
	d, _ := ioutil.ReadAll(c.Request.Body)
	var user map[string]string
	json.Unmarshal(d, &user)
	sdb := db.GetSysDbContext()
	model := model.UserProfile{LoginName: user["user"], DisplayName: user["name"]}
	s := sdb.Create(&model)
	println(s)
	c.JSON(http.StatusOK, ResponseResult(true, "", nil))
}
