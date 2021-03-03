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

package tidb

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestConnection(t *testing.T) {
	// db, err := sql.Open("mysql", "root:111111@tcp(10.3.70.134:4000)/test?charset=utf8")
	// if err != nil {
	// 	fmt.Printf("connect tidb failed ! [%s]", err)
	// } else {
	// 	fmt.Println("connect to tidb successed")
	// }

	// rows, err := db.Query("select sasas,asas from tablea")
	// if err != nil {
	// 	fmt.Printf("select fail [%s]", err)
	// 	return
	// }

	// for rows.Next() {
	// 	var id string
	// 	var username string
	// 	rows.Columns()
	// 	err := rows.Scan(&id, &username)
	// 	if err != nil {
	// 		fmt.Printf("get user info error [%s]", err)
	// 	}
	// }
}

// func TestSchemaToXML(t *testing.T) {
// 	converter := &EnvConverter{
// 		Connector: config.DbConnector{
// 			Host:         "10.3.70.132",
// 			Port:         4000,
// 			UserName:     "root",
// 			Password:     "111111",
// 			DatabaseName: "test",
// 		},
// 	}
// 	converter.SchemaToModelXML()
// }

// func TestMigrationToSQL(t *testing.T) {
// 	absPath, _ := filepath.Abs("../metadata/migration/Migration.xml")
// 	file, err := os.Open(absPath)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	data, err := ioutil.ReadAll(file)

// 	converter := EnvConverter{
// 		Connector: config.DbConnector{
// 			Host:         "10.3.70.132",
// 			Port:         4000,
// 			UserName:     "root",
// 			Password:     "111111",
// 			DatabaseName: "test",
// 		},
// 	}
// 	converter.MigrationToSQL(data)
// }
