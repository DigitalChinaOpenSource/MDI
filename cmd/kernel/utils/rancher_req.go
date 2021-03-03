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
	"fmt"
	"io/ioutil"

	"net/http"
	"strings"
)

const (
	RancherHost        = "https://rancher.wh.digitalchina.com/v3/project/c-wsxm5:p-29rmg"
	RancherBearerToken = "Bearer token-qw8fn:z6j59xn7nr2gfh6lzk2wxd87w759b2mjb9mp72gdck6hqrc88d5xll"
	RancherProjectId   = "c-wsxm5:p-29rmg"
)

//SendRancherRequest 向rancher发送HTTP请求
func SendRancherRequest(method string, route string, payload *strings.Reader) (statusCode int, content string) {
	url := RancherHost + route
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Authorization", RancherBearerToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err == nil {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)

		fmt.Println(string(body))

		return res.StatusCode, string(body)
	}
	return 400, ""
}
