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

package middleware

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Authenticate 验证登录身份
func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		//读取用户信息
		// user := c.Request.Header.Get("X-User")
		// fmt.Println("X-User:" + user)

		// if len(user) > 0 {
		// 	sdb := db.GetSysDbContext()
		// 	var u model.UserProfile
		// 	result := sdb.Where("login_name = ?", user).First(&u)
		// 	if result.Error == nil {
		// 		c.Next()
		// 		return
		// 	}
		// }

		token := c.Request.Header.Get("Access-Token")
		if len(token) > 0 {
			url := "http://mdi-kratos-public.dev.wh.digitalchina.com/sessions/whoami"
			method := "POST"

			client := &http.Client{}
			req, err := http.NewRequest(method, url, nil)

			req.Header.Add("Authorization", "Bearer "+token)

			res, err := client.Do(req)
			if err == nil && res.StatusCode == http.StatusOK {
				defer res.Body.Close()
				body, _ := ioutil.ReadAll(res.Body)
				var tobj map[string]interface{}
				json.Unmarshal(body, &tobj)
				if tobj["active"] == true {
					identity := tobj["identity"].(map[string]interface{})
					traits := identity["traits"].(map[string]interface{})
					//fmt.Println("X-User:" + traits["email"].(string))
					c.Set("X-User", traits["email"].(string))
					c.Next()
					return
				}
			}
		}
		c.JSON(http.StatusUnauthorized, gin.H{"message": "身份验证失败"})
		c.Abort()
	}
}
