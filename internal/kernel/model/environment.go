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

package model

import (
	"dataapi/internal/pkg/utils"
)

//Environment ..
type Environment struct {
	EnvironmentID     string `gorm:"primaryKey;column:environment_id"`
	ProjectID         string `gorm:"column:project_id"`
	Type              uint16 `gorm:"column:type"`
	Owner             string `gorm:"column:owner"`
	GraphCurrent      string `gorm:"column:graph_current"`
	MetadataCurrent   string `gorm:"column:metadata_current"`
	MetadataPublished string `gorm:"column:metadata_published"`
	SQLHost           string `gorm:"column:sql_host"`
	SQLPort           int    `gorm:"column:sql_port"`
	SQLUser           string `gorm:"column:sql_user"`
	SQLPassword       string `gorm:"column:sql_password"`
	SQLDBName         string `gorm:"column:sql_dbname"`
	SQLSchema         string `gorm:"column:sql_schema"`
	AgentDomain       string `gorm:"column:agent_domain"`
	AgentKey          string `gorm:"column:agent_key"`
}

//TableName 数据库表名映射
func (Environment) TableName() string {
	return "environment"
}

//EnvironmentUser ..
type EnvironmentUser struct {
	EnvironmentUserID string `gorm:"primaryKey;autoIncrement;column:environment_user_id"`
	EnvironmentID     string `gorm:"column:environment_id"`
	UserLoginName     string `gorm:"column:user_login_name"`
	Permission        uint32 `gorm:"column:permission"`
}

//TableName 数据库表名映射
func (EnvironmentUser) TableName() string {
	return "environment_user"
}

//EnvironmentToken ..
type EnvironmentToken struct {
	TokenID       uint64     `gorm:"primaryKey;autoIncrement;column:token_id"`
	Name          string     `gorm:"column:name"`
	EnvironmentID string     `gorm:"column:environment_id"`
	UserLoginName string     `gorm:"column:user_login_name"`
	Token         string     `gorm:"column:token"`
	CreateOn      utils.Time `gorm:"column:create_on"`
	ExpiredOn     utils.Time `gorm:"column:expired_on"`
	Remark        string     `gorm:"column:remark"`
}

//TableName 数据库表名映射
func (EnvironmentToken) TableName() string {
	return "environment_token"
}

//EnvironmentHistory ..
type EnvironmentHistory struct {
	EnvironmentHistoryID uint64     `gorm:"primaryKey;autoIncrement;column:environment_history_id"`
	EnvironmentID        string     `gorm:"column:environment_id"`
	Publisher            string     `gorm:"column:publisher"`
	PublishOn            utils.Time `gorm:"column:publish_on"`
	MetadataSource       string     `gorm:"column:metadata_source"`
	MetadataTarget       string     `gorm:"column:metadata_target"`
	ActionResult         string     `gorm:"column:action_result"`
	Actions              uint32     `gorm:"column:actions"`
	FailedActions        uint32     `gorm:"column:failed_actions"`
}

//TableName 数据库表名映射
func (EnvironmentHistory) TableName() string {
	return "environment_history"
}

//EnvironmentAccessHistory ..
type EnvironmentAccessHistory struct {
	AccessHistoryID uint64     `gorm:"primaryKey;autoIncrement;column:access_history_id"`
	EnvironmentID   string     `gorm:"column:environment_id"`
	UserLoginName   string     `gorm:"column:user_login_name"`
	CreateOn        utils.Time `gorm:"column:create_on"`
	APIMethod       string     `gorm:"column:api_method"`
	APIUrl          string     `gorm:"column:api_url"`
	HTTPBody        string     `gorm:"column:http_body"`
}

//TableName 数据库表名映射
func (EnvironmentAccessHistory) TableName() string {
	return "environment_access_history"
}
