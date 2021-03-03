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

package router

//const accessToken = "gIX3H578xPi1L65K7pHCJRGM6FNR2NJy"
//
//func TestRegister(t *testing.T) {
//	engine := gin.Default()
//	engine.Use(middleware.Cors())
//	Register(engine)
//	w := httptest.NewRecorder()
//	req := CustomNewRequest("GET","/mdi/api/system/rules.json", nil,accessToken)
//	engine.ServeHTTP(w, req)
//	assert.Equal(t, 200, w.Code)
//}
//
//func CustomNewRequest(method string,url string,body io.Reader,accessToken string)(*http.Request){
//	req, _ := http.NewRequest(method, url,body)
//	req.Header.Add("access-token",accessToken)
//	return req
//}