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
	"dataapi/internal/kernel/model"
	"dataapi/internal/kernel/tidb"
)

// CreateOathkeeperRules
func CreateOathkeeperRules(agentid string) bool {
	to := "http://" + agentid + ".dev.wh.digitalchina.com"
	from := "http://mdi-oathkeeper-proxy.dev.wh.digitalchina.com/" + agentid
	authenticator := "oauth2_introspection"
	rule := model.OathkeeperRule{
		ID:            agentid,
		ToURL:         to,
		FromURL:       from + "/<.*>",
		Methods:       "GET,POST,PATCH,PUT,DELETE",
		Group:         "agent",
		Authenticator: authenticator,
		StripPath:     "/" + agentid + "/",
	}
	db := tidb.GetSysDbContext()
	result := db.Create(&rule)
	if result.Error != nil {
		return false
	} else {
		return true
	}
}

// DeleteOathkeeperRules
func DeleteOathkeeperRules(agentid string) bool {
	rule := model.OathkeeperRule{
		ID: agentid,
	}
	db := tidb.GetSysDbContext()
	result := db.Delete(&rule)
	if result.Error != nil {
		return false
	} else {
		return true
	}
}
