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
	"errors"
	"net/url"
	"strings"
)

// Odata keywords
const (
	Select      = "$select"
	Top         = "$top"
	Skip        = "$skip"
	Count       = "$count"
	OrderBy     = "$orderby"
	InlineCount = "$inlinecount"
	Filter      = "$filter"
)

// ParseURLValues parses url values in odata format into a map of interfaces for the DB adapters to translate
func ParseURLValues(query url.Values) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var parseErrors []string

	result[Count] = false
	result[InlineCount] = "none"

	if isCountAndInlineCountSet(query) {
		parseErrors = append(parseErrors, "$count and $inlinecount cannot be set in the same odata query")
	}

	for queryParam, queryValues := range query {
		var parseResult interface{}
		var err error

		if len(queryValues) > 1 {
			parseErrors = append(parseErrors, "Duplicate keyword '"+queryParam+"' found in odata query")
			continue
		}
		value := query.Get(queryParam)
		if value == "" && queryParam != Count {
			parseErrors = append(parseErrors, "No value was set for keyword '"+queryParam+"'")
			continue
		}

		switch queryParam {
		case Select:
			parseResult, err = parseStringArray(&value)
		case Top:
			parseResult, err = parseInt(&value)
		case Skip:
			parseResult, err = parseInt(&value)
		case Count:
			parseResult = true
		case OrderBy:
			parseResult, err = ParseOrderArray(&value)
		case InlineCount:
			if !isValidInlineCountValue(value) {
				parseErrors = append(parseErrors, "Inline count value needs to be allpages or none")
			}
			parseResult = value
		case Filter:
			parseResult, err = ParseFilterString(value)
		default:
			parseErrors = append(parseErrors, "Keyword '"+queryParam+"' is not valid")
		}

		if err != nil {
			parseErrors = append(parseErrors, err.Error())
		}
		result[queryParam] = parseResult
	}
	if len(parseErrors) > 0 {
		return nil, errors.New(strings.Join(parseErrors[:], ";"))
	}
	return result, nil
}

func isValidInlineCountValue(value string) bool {
	valueNoSpace := strings.TrimSpace(value)
	if valueNoSpace != "allpages" && valueNoSpace != "none" {
		return false
	}
	return true
}

func isCountAndInlineCountSet(query url.Values) bool {

	_, countFound := query[Count]
	_, inlineFound := query[InlineCount]

	if countFound && inlineFound {
		return true
	}

	return false
}
