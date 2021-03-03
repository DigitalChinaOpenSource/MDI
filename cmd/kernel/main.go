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

package main

import (
	"dataapi/cmd/kernel/router"
	"dataapi/cmd/kernel/utils"
	"dataapi/internal/kernel/model"
	"dataapi/internal/kernel/tidb"
	"dataapi/internal/pkg/middleware"
	"encoding/json"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	engine := gin.Default()
	engine.Use(middleware.Cors())
	router.Register(engine)
	engine.GET("/", func(c *gin.Context) {
		c.String(200, "MDI-Kernel Server.")
	})

	//换令牌不需要身份认证
	engine.POST("/mdi/api/token", func(c *gin.Context) {
		accessToken := model.AccessToken{}
		if err := c.ShouldBindJSON(&accessToken); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		} else {
			var tokenInfo model.TokenInfo
			_ = json.Unmarshal([]byte(utils.AesDecrypt(accessToken.UserToken, utils.Key)), &tokenInfo)

			client := model.OauthClient{
				ClientID:      tokenInfo.ClientID,
			}
			db := tidb.GetSysDbContext()
			result := db.First(&client)
			if result.Error != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": result.Error})
				return
			}else {
				if accessToken.UserToken == client.AccessToken{
					conf := clientcredentials.Config{
						ClientID:       client.ClientID,
						ClientSecret:   client.ClientSecret,
						TokenURL:       utils.PublicUrl+"oauth2/token",
						Scopes:         []string{
							"offline",
							"offline_access",
						},
						EndpointParams: nil,
						AuthStyle:      1,
					}

					token := utils.GetToken(conf)

					oauthClient := model.OauthClient{
						ClientID:      client.ClientID,
						UpdatedAt:     time.Now(),
						OauthToken:    token,
					}

					db := tidb.GetSysDbContext()
					result := db.Model(&oauthClient).Updates(oauthClient)
					if result.Error != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": result.Error})
						return
					}else {
						c.JSON(http.StatusOK,gin.H{"token:":token})
						return
					}
				}else {
					c.JSON(http.StatusOK,gin.H{"error:":"无效令牌"})
				}
			}
		}
	})

	return engine
}
func main() {
	engine := setupRouter()

	utils.ReDeployAgent()
	engine.Run(":8080")
}
