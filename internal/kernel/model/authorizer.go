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

type AuthorizerRoleJson struct {
	Id         string `json:"id"`
	Descriptor string `json:"descriptor"`
	Members    []string `json:"members"`
}

type AuthorizerPolicyJson struct {
	Actions      []string          `json:"actions"`
	Conditions   map[string]string `json:"conditions"`
	Description  string            `json:"description"`
	Effect       string            `json:"effect"`
	Id           string            `json:"id"`
	Resources    []string          `json:"resources"`
	Subjects     []string          `json:"subjects"`
}

type EnvAccessControlPolicy struct {
	Read   []string `json:"read"`
	Write  []string `json:"write"`
	Modify []string `json:"modify"`
}