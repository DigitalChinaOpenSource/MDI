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

package validatefield

type validField struct {
	characters map[string]bool
}

func New(characters string) *validField {

	var obj = validField{make(map[string]bool, len(characters))}

	// build map
	for _, val := range characters {
		obj.characters[string(val)] = true
	}

	return &obj
}

func (v *validField) ValidateField(value string) bool {
	return v.characters[value]
}
