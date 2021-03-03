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
	"bytes"
	"context"
	"dataapi/internal/kernel/model"
	"dataapi/internal/kernel/tidb"
	"encoding/json"
	"golang.org/x/oauth2/clientcredentials"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	AdminUrl = "http://mdi-hydra-admin.dev.wh.digitalchina.com/"
	PublicUrl = "http://mdi-hydra-public.dev.wh.digitalchina.com/"
	Key = "123456781234567812345678"
)

func SendHydraRequest(baseurl string ,route string ,method string ,payload io.Reader) (statusCode int, content string) {
	url := baseurl+route
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return 400,err.Error()
	}

	httpclient:=&http.Client{}
	response, resErr := httpclient.Do(req)
	if resErr != nil {
		return 400,resErr.Error()
	}

	defer response.Body.Close()
	body,bodyErr := ioutil.ReadAll(response.Body)
	if bodyErr != nil {
		return 400,bodyErr.Error()
	}
	return 200,string(body)
}

//获取oauth2.0 token
func GetToken (conf clientcredentials.Config) string{
	RevokeToken(conf.ClientID)
	ctx := context.TODO()
	token,_ := conf.Token(ctx)
	return token.AccessToken
}

//注销oauth2.0 token
func RevokeToken (ClientID string) (statusCode int, content string) {
	statusCode, content = SendHydraRequest(AdminUrl, "oauth2/tokens?client_id="+ClientID, "DELETE", nil)
	return statusCode,content
}

func CreateClient(client model.OauthClient) (statusCode int, content string){
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
	statusCode,content = SendHydraRequest(AdminUrl ,"clients","POST",bytes.NewBuffer(reader))
	tokenInfo := model.TokenInfo{
		ClientID:      client.ClientID,
		Owner:         client.Owner,
		UpdateTime:    time.Now().String(),
	}
	encryptCode,_ := json.Marshal(tokenInfo)
	orig := string(encryptCode)
	accessToken := AesEncrypt(orig, Key)

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
		return statusCode,result.Error.Error()
	}else {
		return statusCode,content
	}
}

func DeleteClient(clientID string) (statusCode int, content string) {
	statusCode,content = SendHydraRequest(AdminUrl ,"clients/"+clientID,"DELETE",nil)
	db := tidb.GetSysDbContext()
	data := model.OauthClient{
		ClientID: clientID,
	}
	result := db.Delete(data)
	if result.Error != nil {
		return statusCode,result.Error.Error()
	}
	return statusCode,content
}