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
	"time"
)

type OauthClient struct {
	ClientID      string    `gorm:"primaryKey;column:client_id"`
	ClientName    string    `gorm:"column:client_name"`
	ClientSecret  string    `gorm:"column:client_secret"`
	Owner         string    `gorm:"column:owner"`
	Scope         string    `gorm:"column:scope"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
	AccessToken   string    `gorm:"column:access_token"`
	OauthToken    string    `gorm:"column:oauth_token"`
	ProjectId     string    `gorm:"column:project_id"`
	EnvironmentId string    `gorm:"column:environment_id"`
}

//TableName 数据库表名映射
func (OauthClient) TableName() string {
	return "oauth_client"
}

type OauthClientJson struct {
	ClientID                          string   `json:"client_id"`
	ClientSecret                      string   `json:"client_secret"`
	ClientName                        string   `json:"client_name"`
	ClientSecretExpiresAt             int      `json:"client_secret_expires_at"`
	FrontchannelLogoutSessionRequired bool     `json:"frontchannel_logout_session_required"`
	Scope                             string   `json:"scope"`
	TokenEndpointAuthMethod           string   `json:"token_endpoint_auth_method"`
	UserinfoSignedResponseAlg         string   `json:"userinfo_signed_response_alg"`
	GrantTypes                        []string `json:"grant_types"`
	ResponseTypes                     []string `json:"response_types"`
}

type ClientJson struct {
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	ClientName     string `json:"client_name"`
	Owner          string `json:"owner"`
	EnvironmentId  string `json:"env_id"`
	ProjectId      string `json:"project_id"`
}

type TokenInfo struct {
	ClientID   string    `json:"client_id"`
	Owner      string    `json:"owner"`
	UpdateTime string `json:"update_time"`
}

type AccessToken struct {
	UserToken string `json:"user_token"`
}
