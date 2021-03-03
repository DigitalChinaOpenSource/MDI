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

package webHelper

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

const CONTENT_TYPE = "application/json"
// Post请求到Rancher
func HttpToRancherWithBody(method string, url string, body *strings.Reader, token string) (map[string]interface{}, error) {
	// 忽略https证书
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	// post 请求
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// 设置请求头,添加token
	req.Header.Set("Authorization", token)
	req.Header.Set("Accept", CONTENT_TYPE)
	req.Header.Set("Content-Type", CONTENT_TYPE)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	//函数运行结束前关闭请求结果
	defer resp.Body.Close()

	returnMap, err := parseResponseToMap(resp)
	//http请求状态
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		return returnMap, nil
	} else {
		return nil, err
	}
}

// Get 和 Delete请求到Rancher
func HttpToRancher(method string, url string, token string) (map[string]interface{}, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Accept", CONTENT_TYPE)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	returnMap, err := parseResponseToMap(resp)

	//http请求状态
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		return returnMap, err
	} else {
		//打印错误信息，结束
		return nil, err
	}
}

//适用场景 { key: value}
func parseResponseToMap(response *http.Response) (map[string]interface{}, error) {
	var result map[string]interface{}
	body, err := ioutil.ReadAll(response.Body)
	if err == nil {
		err = json.Unmarshal(body, &result)
	}
	return result, err
}
