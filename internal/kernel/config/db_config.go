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

package config

import "os"

// const (
// 	Dbdrive      = 1
// 	PgConnection = "host=10.3.70.128 port=5432 user=postgres password=111111 dbname=MDI sslmode=disable"
// 	TiConnection = "root:111111@tcp(10.3.70.134:4000)/MDI?charset=utf8mb4&parseTime=true"
// )

// KernelDbDriver 数据库类型
var KernelDbDriver = os.Getenv("KERNEL_DB_DRIVER")

// KernelDbConnection 数据库连接串
var KernelDbConnection = os.Getenv("KERNEL_DB_CONNECTION")

// AgentDbHost agent数据库主机地址
var AgentDbHost = os.Getenv("AGENT_DB_HOST")

// AgentDbPort agent数据库主机端口
var AgentDbPort = os.Getenv("AGENT_DB_PORT")

var AgentDbUser = os.Getenv("AGENT_DB_USER")

var AgentDbPassword = os.Getenv("AGENT_DB_PASSWORD")

// DbConnector ..
type DbConnector struct {
	ID           string
	Host         string // 主机名
	Port         int    // 端口
	UserName     string // 用户名
	Password     string // 密码
	DatabaseName string // 数据库名
	Charset      string // 字符集
	TimeZone     string // 时区
}
