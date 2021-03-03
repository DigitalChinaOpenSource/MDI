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

package db

import (
	"dataapi/internal/kernel/config"
	"dataapi/internal/kernel/postgres"
	"dataapi/internal/kernel/tidb"
	"database/sql"

	"gorm.io/gorm"
)

// GetSysDbContext ...
func GetSysDbContext() *gorm.DB {
	if config.KernelDbDriver == "tidb" {
		return tidb.GetSysDbContext()
	}
	if config.KernelDbDriver == "postgres" {
		return postgres.GetSysDbContext()
	}
	panic("DbDriver type is must to be assigned! please set for environment variable [KERNEL_DB_DRIVER]. ")
}

//GetEnvDbContext ..
func GetEnvDbContext(connector config.DbConnector) *sql.DB {
	if config.KernelDbDriver == "tidb" {
		return tidb.GetEnvDbContext(connector)
	}
	if config.KernelDbDriver == "postgres" {
		return postgres.GetEnvDbContext(connector)
	}
	panic("DbDriver type is must to be assigned! please set for environment variable [KERNEL_DB_DRIVER]. ")
}
