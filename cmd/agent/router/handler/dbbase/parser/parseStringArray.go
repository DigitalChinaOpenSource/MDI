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

package parser

import (
	"dataapi/cmd/agent/router/handler/dbbase/parser/validatefield"
	"errors"
	"strings"
)

// OrderItem holds order key information
type OrderItem struct {
	Field string
	Order string
}

func parseStringArray(value *string) ([]string, error) {
	result := strings.Split(*value, ",")

	// trim out space
	for idx, resultNoSpace := range result {
		result[idx] = strings.TrimSpace(resultNoSpace)
	}

	if len(result) == 0 {
		return nil, errors.New("cannot parse zero length string")
	}

	return result, nil
}

func ParseOrderArray(value *string) ([]OrderItem, error) {
	parsedArray, err := parseStringArray(value)
	if err != nil {
		return nil, err
	}

	// Validate values for special characters
	valid := validatefield.New("~!@#$%^&*()_+-")
	for _, val := range parsedArray {
		if valid.ValidateField(val) || val == "" {
			return nil, errors.New("Cannot support field " + val)
		}
	}

	result := make([]OrderItem, len(parsedArray))

	for i, v := range parsedArray {
		timmedString := strings.TrimSpace(v)
		compressedSpaces := strings.Join(strings.Fields(timmedString), " ")
		s := strings.Split(compressedSpaces, " ")

		if len(s) > 2 {
			return nil, errors.New("Cannot have more than 2 items in orderby query")
		}

		if len(s) > 1 {
			if s[1] != "asc" &&
				s[1] != "desc" {
				return nil, errors.New("Second value in orderby needs to be asc or desc")
			}
			result[i] = OrderItem{s[0], s[1]}
			continue
		}
		result[i] = OrderItem{compressedSpaces, "asc"}
	}
	return result, nil
}