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

// Project ..
type Project struct {
	ProjectID   string     `gorm:"primaryKey;column:project_id"`
	Name        string     `gorm:"column:name"`
	Description string     `gorm:"column:description"`
	Creator     string     `gorm:"column:creator"`
	CreateOn    utils.Time `gorm:"column:create_on"`
	Icon        string     `gorm:"column:icon"`
}

//TableName 数据库表名映射
func (Project) TableName() string {
	return "project"
}

//ProjectUser ..
type ProjectUser struct {
	ProjectUserID  uint32 `gorm:"primaryKey;autoIncrement;column:project_user_id"`
	ProjectID      string `gorm:"column:project_id"`
	UserLoginName  string `gorm:"column:user_login_name"`
	IsProjectOwner bool   `gorm:"column:is_project_owner"`
}

//TableName 数据库表名映射
func (ProjectUser) TableName() string {
	return "project_user"
}

//ProjectStar ..
type ProjectStar struct {
	ProjectStarID uint32     `gorm:"primaryKey;column:project_star_id"`
	ProjectID     string     `gorm:"column:project_id"`
	UserLoginName string     `gorm:"column:user_login_name"`
	CreateOn      utils.Time `gorm:"column:create_on"`
}

//TableName 数据库表名映射
func (ProjectStar) TableName() string {
	return "project_star"
}
