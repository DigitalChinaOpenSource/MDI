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
	"bytes"
	"dataapi/cmd/kernel/utils"
	"dataapi/internal/kernel/model"
	"dataapi/internal/kernel/tidb"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type TokenHandler struct {}

//获取所有客户端
func (eventh *TokenHandler) GetClient(c *gin.Context) {
	statusCode,content := utils.SendHydraRequest(utils.AdminUrl ,"clients","GET",nil)
	c.String(statusCode,content)
	return
}

//根据客户端id获取客户端
func (eventh *TokenHandler) GetUserClient(c *gin.Context) {
	clientID := c.Param("clientId")
	oauthClient := model.OauthClient{
		ClientID:      clientID,
	}
	db := tidb.GetSysDbContext()
	result := db.Where("client_id=?",clientID).Find(&oauthClient)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest,ResponseResult(false,"查询失败",nil))
		return
	}
	c.JSON(http.StatusOK,ResponseResult(true,"查询成功",oauthClient))
	return
}

//创建客户端
func (eventh *TokenHandler) CreateClient(c *gin.Context) {
	client := model.ClientJson{}
	if err := c.ShouldBindJSON(&client); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		hydraClient := model.OauthClientJson{
			ClientID:                          client.ClientID,
			ClientSecret:                      client.ClientSecret,
			ClientName:                        client.ClientName,
			ClientSecretExpiresAt:             0,
			FrontchannelLogoutSessionRequired: false,
			Scope:                             "offline offline_access",
			TokenEndpointAuthMethod:           "client_secret_post",
			UserinfoSignedResponseAlg:         "none",
			GrantTypes:                        []string{"client_credentials"},
			ResponseTypes:                     []string{"code"},
		}
		reader, _ := json.Marshal(hydraClient)
		statusCode,content := utils.SendHydraRequest(utils.AdminUrl ,"clients","POST",bytes.NewBuffer(reader))

		tokenInfo := model.TokenInfo{
			ClientID:      client.ClientID,
			Owner:         client.Owner,
			UpdateTime:    time.Now().String(),
		}
		encryptCode,_ := json.Marshal(tokenInfo)
		orig := string(encryptCode)
		accessToken := utils.AesEncrypt(orig, utils.Key)

		oauthClient := model.OauthClient{
			ClientID:      client.ClientID,
			ClientName:    client.ClientName,
			ClientSecret:  client.ClientSecret,
			Owner:         client.Owner,
			Scope:         "offline",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			AccessToken:   accessToken,
			OauthToken:    "",
			ProjectId:     client.ProjectId,
			EnvironmentId: client.EnvironmentId,
		}
		db := tidb.GetSysDbContext()
		result := db.Create(&oauthClient)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": result.Error})
			return
		}else {
			c.String(statusCode,content)
			return
		}
	}
}

//删除客户端
func (eventh *TokenHandler) DeleteClient(c *gin.Context) {
	utils.SendHydraRequest(utils.AdminUrl ,"clients/"+c.Param("clientId"),"DELETE",nil)
	db := tidb.GetSysDbContext()
	data := model.OauthClient{
		ClientID: c.Param("clientId"),
	}
	result := db.Delete(data)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error})
		return
	}
	return
}

//更新客户端
func (eventh *TokenHandler) UpdateClient(c *gin.Context) {
	client := model.ClientJson{}
	if err := c.ShouldBindJSON(&client); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		hydraClient := model.OauthClientJson{
			ClientSecret:                      client.ClientSecret,
			ClientName:                        client.ClientName,
			ClientSecretExpiresAt:             0,
			FrontchannelLogoutSessionRequired: false,
			Scope:                             "offline offline_access",
			TokenEndpointAuthMethod:           "client_secret_post",
			UserinfoSignedResponseAlg:         "none",
			GrantTypes:                        []string{"client_credentials"},
			ResponseTypes:                     []string{"code"},
		}
		reader, _ := json.Marshal(hydraClient)
		statusCode,content := utils.SendHydraRequest(utils.AdminUrl ,"clients/"+client.ClientID,"PUT",bytes.NewBuffer(reader))

		tokenInfo := model.TokenInfo{
			ClientID:   client.ClientID,
			Owner:      client.Owner,
			UpdateTime: time.Now().String(),
		}

		encryptCode,_ := json.Marshal(tokenInfo)
		orig := string(encryptCode)
		accessToken := utils.AesEncrypt(orig, utils.Key)

		oauthClient := model.OauthClient{
			ClientID:      client.ClientID,
			ClientName:    client.ClientName,
			ClientSecret:  client.ClientSecret,
			Owner:         client.Owner,
			UpdatedAt:     time.Now(),
			EnvironmentId: client.EnvironmentId,
			ProjectId:     client.ProjectId,
			AccessToken:   accessToken,
		}

		db := tidb.GetSysDbContext()
		result := db.Model(&oauthClient).Updates(oauthClient)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": result.Error})
			return
		}else {
			c.String(statusCode,content)
			return
		}
	}
}

//更新accesstoken
func (eventh *TokenHandler) UpdateAccessToken(c *gin.Context) {
	tokenInfo := model.TokenInfo{}
	if err := c.ShouldBindJSON(&tokenInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		tokenInfo.UpdateTime = time.Now().String()
		encryptCode,_ := json.Marshal(tokenInfo)
		orig := string(encryptCode)
		accessToken := utils.AesEncrypt(orig, utils.Key)

		oauthClient := model.OauthClient{
			ClientID:      tokenInfo.ClientID,
			AccessToken:   accessToken,
			UpdatedAt:     time.Now(),
		}

		db := tidb.GetSysDbContext()
		result := db.Model(&oauthClient).Updates(oauthClient)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": result.Error})
			return
		}else {
			c.JSON(http.StatusOK,gin.H{"AccessToken":accessToken})
			return
		}
	}
}