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

import "strings"

// Token constants
const (
	filterTokenOpenParen int = iota
	filterTokenCloseParen
	filterTokenWhitespace
	filterTokenComma
	filterTokenLogical
	filterTokenFunc
	filterTokenFloat
	filterTokenInteger
	filterTokenString
	filterTokenDate
	filterTokenTime
	filterTokenDateTime
	filterTokenBoolean
	FilterTokenLiteral
)

// GlobalFilterTokenizer the global filter tokenizer
var globalFilterTokenizer = filterTokenizer()

// GlobalFilterParser the global filter parser
var globalFilterParser = filterParser()

// ParseFilterString Converts an input string from the $filter part of the URL into a parse
// tree that can be used by providers to create a response.
func ParseFilterString(filter string) (*ParseNode, error) {
	//判断一下filter条件中substring到底穿了几个参数，默认传3个，如果传2个需要标记一下。标记方式：在后面加个"_"
	CheckSubstringParams(&filter)
	tokens, err := globalFilterTokenizer.tokenize(filter)
	if err != nil {
		return nil, err
	}
	postfix, err := globalFilterParser.infixToPostfix(tokens)
	if err != nil {
		return nil, err
	}
	tree, err := globalFilterParser.postfixToTree(postfix)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

//检查substring带的参数数量并作标记处理，标记在postTree时使用
func CheckSubstringParams(filter *string) {
	arr := strings.Split(*filter, "substring")
	res := ""
	if len(arr) > 1 {
		for i := 1; i < len(arr); i++ {
			start := strings.Index(arr[i], "(")
			end := strings.Index(arr[i], ")")
			//切割()获得例如下列字符串：name,1,4或者name,3
			str := arr[i][start+1 : end]
			splitArr := strings.Split(str, ",")
			if len(splitArr) == 2 {
				res += "substring_" + arr[i]
			} else {
				res += "substring" + arr[i]
			}
		}
		res = arr[0] + res
		*filter = res
	}
}

// FilterTokenizer Creates a tokenizer capable of tokenizing filter statements
func filterTokenizer() *Tokenizer {
	t := Tokenizer{}
	t.add("^\\(", filterTokenOpenParen)
	t.add("^\\)", filterTokenCloseParen)
	t.add("^,", filterTokenComma)
	t.add("^(eq|ne|gt|ge|lt|le|and|or) ", filterTokenLogical)
	t.add("^(contains|endswith|startswith)", filterTokenFunc)
	t.add("^-?[0-9]+\\.[0-9]+", filterTokenFloat)
	t.add("^-?[0-9]+", filterTokenInteger)
	t.add("^(?i:true|false)", filterTokenBoolean)
	t.add("^'(''|[^'])*'", filterTokenString)
	t.add("^-?[0-9]{4,4}-[0-9]{2,2}-[0-9]{2,2}", filterTokenDate)
	t.add("^[0-9]{2,2}:[0-9]{2,2}(:[0-9]{2,2}(.[0-9]+)?)?", filterTokenTime)
	t.add("^[0-9]{4,4}-[0-9]{2,2}-[0-9]{2,2}T[0-9]{2,2}:[0-9]{2,2}(:[0-9]{2,2}(.[0-9]+)?)?(Z|[+-][0-9]{2,2}:[0-9]{2,2})", filterTokenDateTime)
	t.add("^[a-zA-Z][a-zA-Z0-9_.]*", FilterTokenLiteral)
	t.add("^_id", FilterTokenLiteral)
	t.ignore("^ ", filterTokenWhitespace)

	return &t
}

// FilterParser creates the definitions for operators and functions
func filterParser() *Parser {
	parser := emptyParser()
	parser.defineOperator("gt", 2, opAssociationLeft, 4)
	parser.defineOperator("ge", 2, opAssociationLeft, 4)
	parser.defineOperator("lt", 2, opAssociationLeft, 4)
	parser.defineOperator("le", 2, opAssociationLeft, 4)
	parser.defineOperator("eq", 2, opAssociationLeft, 3)
	parser.defineOperator("ne", 2, opAssociationLeft, 3)
	parser.defineOperator("and", 2, opAssociationLeft, 2)
	parser.defineOperator("or", 2, opAssociationLeft, 1)
	parser.defineFunction("contains", 2)
	parser.defineFunction("endswith", 2)
	parser.defineFunction("startswith", 2)
	parser.defineFunction("length", 1)
	parser.defineFunction("indexof", 2)
	parser.defineFunction("replace", 3)
	parser.defineFunction("substring", 3)
	parser.defineFunction("substring_", 2)
	parser.defineFunction("tolower", 1)
	parser.defineFunction("toupper", 1)
	parser.defineFunction("trim", 1)
	parser.defineFunction("concat", 2)
	parser.defineFunction("year", 1)
	parser.defineFunction("month", 1)
	parser.defineFunction("day", 1)
	parser.defineFunction("hour", 1)
	parser.defineFunction("minute", 1)
	parser.defineFunction("second", 1)
	parser.defineFunction("round", 1)
	parser.defineFunction("floor", 1)
	parser.defineFunction("ceiling", 1)
	return parser
}
