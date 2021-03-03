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

package postgres

import (
	"dataapi/internal/kernel/config"
	"dataapi/internal/kernel/model"
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var sysdb *gorm.DB

//GetSysDbContext ..
func GetSysDbContext() *gorm.DB {
	if sysdb == nil {
		var err error
		sysdb, err = gorm.Open(postgres.Open(config.KernelDbConnection), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
	}

	return sysdb
}

var envdbMap map[string]*sql.DB

//GetEnvDbContext ..
func GetEnvDbContext(connector config.DbConnector) *sql.DB {
	if envdbMap == nil {
		envdbMap = make(map[string]*sql.DB)
	}

	db, ok := envdbMap[connector.ID]
	if ok {
		return db
	} else {
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", connector.Host, connector.Port, connector.UserName, connector.Password, connector.DatabaseName)
		db, err := sql.Open("postgres", connStr)
		//db.SetConnMaxLifetime(time.Hour)

		if err != nil {
			panic(err)
		}

		err = db.Ping()
		if err != nil {
			panic(err)
		}
		fmt.Println("Successfully connected!")

		envdbMap[connector.ID] = db
		return db
	}
}

// GetEnvDbContextByID ..
func GetEnvDbContextByID(eid string) *sql.DB {
	db := GetSysDbContext()
	env := model.Environment{}
	result := db.Where(`environment_id = ?`, eid).First(&env)
	if result.Error == nil {
		edb := GetEnvDbContext(config.DbConnector{
			ID:           env.EnvironmentID,
			Host:         env.SQLHost,
			Port:         env.SQLPort,
			UserName:     env.SQLUser,
			Password:     env.SQLPassword,
			DatabaseName: env.SQLDBName,
		})
		return edb
	}
	return nil
}
