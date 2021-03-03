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

package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaginationResult struct {
	Size  int         `json:"size"`
	Page  int         `json:"page"`
	Rows  interface{} `json:"rows"`
	Total int64       `json:"total"`
}

// Paginate 基于gin和gorm实现的分页器
func Paginate(c *gin.Context, db *gorm.DB, dest interface{}) PaginationResult {
	result := PaginationResult{}
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	result.Page = page
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	switch {
	case pageSize > 100:
		pageSize = 100
	case pageSize <= 0:
		pageSize = 10
	}
	result.Size = pageSize
	db.Count(&result.Total)

	var data []map[string]interface{}

	offset := (page - 1) * pageSize
	query := db.Select(dest).Offset(offset).Limit(pageSize).Find(&data)
	if query.Error == nil {
		result.Rows = data
	}
	return result
}
