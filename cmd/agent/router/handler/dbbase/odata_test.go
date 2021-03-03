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

package dbbase

import (
	"net/url"
	"testing"
)

func TestODataSQLFilter(t *testing.T) {
	query := url.Values{}
	query["$filter"] = []string{"floor(height) lt 190"}
	filter, _ := ODataSQLFilter(query)
	if filter == "FLOOR(height) < 190" {
		t.Log(filter)
	} else {
		t.Fatal(filter + " unexpected sql fragments")
	}

	query["$filter"] = []string{"ceiling(weight) ge 80"}
	filter, _ = ODataSQLFilter(query)
	if filter == "CEILING(weight) >= 80" {
		t.Log(filter)
	} else {
		t.Fatal(filter + " unexpected sql fragments")
	}

	query["$filter"] = []string{"name eq 'll' or contains(addr,'bridge')"}
	filter, _ = ODataSQLFilter(query)
	if filter == "name = 'll' or addr LIKE '%bridge%'" {
		t.Log(filter)
	} else {
		t.Fatal(filter + " unexpected sql fragments")
	}

	query["$filter"] = []string{"replace(name,'p','l') eq 'll' and substring(name,1,4) eq 'lory'"}
	filter, _ = ODataSQLFilter(query)
	if filter == "REPLACE(name,'p','l') = 'll' and SUBSTRING(name,1,4) = 'lory'" {
		t.Log(filter)
	} else {
		t.Fatal(filter + " unexpected sql fragments")
	}

	query["$filter"] = []string{"year(birth) eq 1997 or length(addr) le 9"}
	filter, _ = ODataSQLFilter(query)
	if filter == "YEAR(birth) = 1997 or LENGTH(addr) <= 9" {
		t.Log(filter)
	} else {
		t.Fatal(filter + " unexpected sql fragments")
	}

	query["$filter"] = []string{"indexof(addr,'street') eq 9"}
	filter, _ = ODataSQLFilter(query)
	if filter == "INSTR(addr,'street') = 9" {
		t.Log(filter)
	} else {
		t.Fatal(filter + " unexpected sql fragments")
	}

	//不存在的关键字
	query["$filter"] = []string{"myself(name,'n')"}
	filter, _ = ODataSQLFilter(query)
	if filter == "" {
		t.Log(filter)
	} else {
		t.Fatal(filter + " unexpected sql fragments")
	}

	query["$filter"] = []string{"substring(name,6) eq 'you'"}
	filter, _ = ODataSQLFilter(query)
	if filter == "SUBSTRING(name,6) = 'you'" {
		t.Log(filter)
	} else {
		t.Fatal(filter + " unexpected sql fragments")
	}
}
