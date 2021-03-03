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
	"dataapi/cmd/kernel/utils"
	"dataapi/internal/kernel/db"
	"dataapi/internal/kernel/model"
	"dataapi/internal/kernel/tidb"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//SysHandler ..
type SysHandler struct{}

// GetOathkeeperRules ..
func (eventh *SysHandler) GetOathkeeperRules(c *gin.Context) {
	result, err := eventh.BuildOathkeeperRules()

	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, result)
	}
}

// BuildOathkeeperRules ..
func (eventh *SysHandler) BuildOathkeeperRules() ([]model.OathkeeperRuleForJSON, error) {
	var rules []model.OathkeeperRuleForJSON
	var list []model.OathkeeperRule

	dbContext := tidb.GetSysDbContext()
	result := dbContext.Find(&list)

	if result.Error != nil {
		return rules, result.Error
	} else {
		for _, r := range list {
			item := model.OathkeeperRuleForJSON{
				ID: r.ID,
				UpStream: model.OathkeeperUpstream{
					URL:          r.ToURL,
					PreserveHost: false,
					StripPath:    r.StripPath,
				},
				Match: model.OathkeeperMatch{
					Methods: strings.Split(r.Methods, ","),
					URL:     r.FromURL,
				},
				Authenticators: []model.OathkeeperHandler{
					model.OathkeeperHandler{
						Handler: r.Authenticator,
					},
				},
				Mutators: []model.OathkeeperHandler{
					model.OathkeeperHandler{
						Handler: "noop",
					},
				},
			}
			item.Authorizer = model.AuthorizerHandler{
				Handler: "allow",
			}
			rules = append(rules, item)
		}
		return rules, nil
	}
}

// CreateOathkeeperRules
func (eventh *SysHandler) CreateOathkeeperRules(c *gin.Context) {
	id := c.Param("agentid")
	result := utils.CreateOathkeeperRules(id)
	if result == false {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"result:": "success"})
		return
	}
}

// DeleteOathkeeperRules
func (eventh *SysHandler) DeleteOathkeeperRules(c *gin.Context) {
	id := c.Param("agentid")
	rule := model.OathkeeperRule{
		ID: id,
	}
	db := tidb.GetSysDbContext()
	db.Delete(&rule)
	return
}

//GetCDMEntityList 查询CDM实体列表
func (h *SysHandler) GetCDMEntityList(c *gin.Context) {
	key := c.Query("key")
	sdb := db.GetSysDbContext()
	var entities []model.CDMEntity
	query := sdb.Model(model.CDMEntity{})
	if len(key) > 0 {
		ws := "%" + key + "%"
		query = query.Where("schema_name like ?", ws).Or("display_name like ?", ws)
	}
	result := query.Order("sort").Find(&entities)
	if result.Error == nil {
		c.JSON(http.StatusOK, ResponseResult(true, "查询成功", entities))
	} else {
		c.JSON(http.StatusBadRequest, ResponseResult(false, "查询失败", nil))
	}
}
