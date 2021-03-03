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

//OathkeeperRule ..
type OathkeeperRule struct {
	ID            string `gorm:"primaryKey;column:id"`
	Authenticator string `gorm:"column:authenticator"`
	ToURL         string `gorm:"column:to_url"`
	FromURL       string `gorm:"column:from_url"`
	Methods       string `gorm:"column:methods"`
	Group         string `gorm:"column:group"`
	StripPath     string `gorm:"column:strip_path"`
	KetoAction    string `gorm:"column:keto_action"`
}

//TableName 数据库表名映射
func (OathkeeperRule) TableName() string {
	return "oathkeeper_rule"
}

// OathkeeperRuleForJSON ..
type OathkeeperRuleForJSON struct {
	ID             string              `json:"id"`
	UpStream       OathkeeperUpstream  `json:"upstream"`
	Match          OathkeeperMatch     `json:"match"`
	Authenticators []OathkeeperHandler `json:"authenticators"`
	Authorizer     AuthorizerHandler   `json:"authorizer"`
	Mutators       []OathkeeperHandler `json:"mutators"`
}

// OathkeeperUpstraem ..
type OathkeeperUpstream struct {
	URL          string `json:"url"`
	PreserveHost bool   `json:"preserve_host"`
	StripPath    string `json:"strip_path"`
}

// OathkeeperMatch ..
type OathkeeperMatch struct {
	Methods []string `json:"methods"`
	URL     string   `json:"url"`
}

// OathkeeperHandler ..
type OathkeeperHandler struct {
	Handler string `json:"handler"`
}

type AuthorizerHandler struct {
	Handler string           `json:"handler"`
}