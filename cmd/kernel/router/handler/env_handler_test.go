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

package handler

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/stretchr/testify"
)

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestGetCurrentLoginUser(t *testing.T) {
	c := &gin.Context{
		Request:  nil,
		Writer:   nil,
		Params:   nil,
		Keys:     nil,
		Errors:   nil,
		Accepted: nil,
	}
	c.Keys = map[string]interface{}{
		"X-User": "123",
	}
	result := GetCurrentLoginUser(c)
	if result != "123" {
		t.Errorf("\n result != 123 \n result:%s", result)
	}

}

func TestResponseResult(t *testing.T) {
	res := gin.H{
		"success": true,
		"message": "123",
		"data":    nil,
	}
	testRes := ResponseResult(true, "123", nil)
	if reflect.DeepEqual(res, testRes) != true {
		t.Errorf("\n res != testRes; \n res:%s \n test:%s", res, testRes)
	}
}
