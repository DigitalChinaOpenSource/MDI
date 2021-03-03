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
	"dataapi/cmd/agent/router/handler/dbbase/parser"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// ErrInvalidInput Client errors
var ErrInvalidInput = errors.New("odata syntax error")

var sqlOperators = map[string]string{
	"eq":         "=",
	"ne":         "!=",
	"gt":         ">",
	"ge":         ">=",
	"lt":         "<",
	"le":         "<=",
	"or":         "or",
	"and":        "and",
	"contains":   "%%%s%%",
	"endswith":   "%%%s",
	"startswith": "%s%%",
	"length":     "LENGTH",
	"indexof":    "INSTR",
	"replace":    "REPLACE",
	"substring":  "SUBSTRING",
	"tolower":    "LOWER",
	"toupper":    "UPPER",
	"trim":       "TRIM",
	"concat":     "CONCAT",
	"year":       "YEAR",
	"month":      "MONTH",
	"day":        "DAY",
	"hour":       "HOUR",
	"minute":     "MINUTE",
	"second":     "SECOND",
	"round":      "ROUND",
	"floor":      "FLOOR",
	"ceiling":    "CEILING",
}

func ODataSQLFilter(query url.Values) (string, error) {

	// Parse url values
	queryMap, err := parser.ParseURLValues(query)
	if err != nil {
		return "", errors.Wrap(ErrInvalidInput, err.Error())
	}

	var finalQuery strings.Builder

	finalQuery.WriteString(" FROM ")
	filter := ""
	// WHERE clause
	if queryMap[parser.Filter] != nil {
		finalQuery.WriteString(" WHERE ")
		filterQuery, _ := queryMap[parser.Filter].(*parser.ParseNode)
		filterClause, err := applyFilter(filterQuery)
		if err != nil {
			return "", errors.Wrap(ErrInvalidInput, err.Error())
		}
		filter = filterClause
		finalQuery.WriteString(filterClause)
	}

	return filter, nil
}

func applyFilter(node *parser.ParseNode) (string, error) {

	//if len(node.Children) != 2 {
	//	return "", ErrInvalidInput
	//}

	var filter strings.Builder

	operator := node.Token.Value.(string)
	sqlOp := sqlOperators[strings.Trim(operator, "_")]
	if operator == "" || sqlOp == "" {
		return "", ErrInvalidInput
	}

	//由于之前的设计是有局限性的，只支持满二叉树结构的情况，所以遇到eq这样的比较符就以为到了终点。现在为了满足像length(name) eq 1这样的语句
	//须在此处判断节点值为length的时候，改节点还有没有子节点，有子节点需要再次递归迭代。同理以下还要实现indexOf，replace等等的。
	//不能简单地判断eq,ne就结束，随着关键字地丰富，indexOf(name.'Ti') eq 7, replace(name,'M','u') eq 'mumu'等情况会出现。所以需要在eq,ne...这一步继续使用递归。

	switch operator {

	case "eq", "ne", "gt", "ge", "lt", "le":

		if _, keyOk := node.Children[0].Token.Value.(string); !keyOk {
			return "", ErrInvalidInput
		}
		//取左子树，判断一下是否是否含有length,indexOf等关键字并且还有子树，是的话继续递归
		leftNode := node.Children[0]
		var left, right string
		if leftNode.Children != nil && len(leftNode.Children) > 0 {
			left, _ = applyFilter(node.Children[0])
		} else {
			left = node.Children[0].Token.Value.(string)
		}
		if CheckSubTreeExists(node.Children[1]) {
			right, _ = applyFilter(node.Children[1])
		} else {
			right = fmt.Sprintf("%v", node.Children[1].Token.Value)
		}

		fmt.Fprintf(&filter, "%s %s %s", left, sqlOp, right)
	//操作符与字符以二叉树的结构组织起来，考虑filter中的分组情况，即()会影响条件执行的顺序。
	//比如，age > 32 or salary < 6000 and id > 56 与 (age > 32 or salary < 6000) and id > 56执行结果可能不同。
	//其体现的业务意义也是不一样的，所以需要标记and，or这些条件连接词，我这里选择的方法是在and,or后面加一个_，即代表它所连接的左右子树是同一组
	//这左右子树需要用括号括起来，直接Fprintf("(左子树+连接符+右子树)")就能实现。
	case "or", "or_", "and", "and_":

		leftFilter, err := applyFilter(node.Children[0]) // Left children
		if err != nil {
			return "", err
		}
		rightFilter, err := applyFilter(node.Children[1]) // Right children
		if err != nil {
			return "", err
		}
		if operator == "or_" || operator == "and_" {
			fmt.Fprintf(&filter, "(%s %s %s)", leftFilter, strings.Trim(operator, "_"), rightFilter)
		} else {
			fmt.Fprintf(&filter, "%s %s %s", leftFilter, operator, rightFilter)
		}
	//Functions
	case "contains", "endswith", "startswith":
		//左右子树都有
		err := HandleContains(node, &filter, sqlOp)
		if err != nil {
			return "", err
		}

	case "indexof", "concat":
		//左右子树都有
		err := HandleIndexOf(node, &filter, sqlOp)
		if err != nil {
			return "", err
		}

	//单子树的情况，length(name),upper(name),lower(addr)
	case "length", "tolower", "toupper", "trim":
		//只有左子树
		err := HandleLength(node, &filter, sqlOp)
		if err != nil {
			return "", err
		}
	case "replace":
		err := HandleReplace(node, &filter, sqlOp)
		if err != nil {
			return "", err
		}
	case "substring", "substring_":
		err := HandleSubstring(node, &filter, operator, sqlOp)
		if err != nil {
			return "", err
		}
	case "year", "month", "day", "hour", "minute", "second","round", "floor", "ceiling":
		err := HandDatetime(node, &filter, sqlOp)
		if err != nil {
			return "", err
		}
	}

	return filter.String(), nil
}

func escapeQuote(value string) string {
	if len(value) <= 1 {
		return ""
	}

	if value[0] == '\'' {
		value = value[1:]
	}
	if value[len(value)-1] == '\'' {
		value = value[:len(value)-1]
	}

	return value
}

func CheckSubTreeExists(node *parser.ParseNode) bool {
	if len(node.Children) > 0 {
		return true
	} else {
		return false
	}
}

func HandleSubstring(node *parser.ParseNode, filter *strings.Builder, operator string, sqlOp string) error {
	if _, ok := node.Children[0].Token.Value.(string); !ok {
		return ErrInvalidInput
	}
	left := node.Children[0].Token.Value.(string)
	if operator == "substring_" {
		//substring只传两个参数的情况
		right := strconv.Itoa(node.Children[1].Token.Value.(int))
		fmt.Fprintf(filter, "%s(%s,%s)", strings.Trim(sqlOp, "_"), left, right)
	} else {
		mid := strconv.Itoa(node.Children[1].Token.Value.(int))
		right := strconv.Itoa(node.Children[2].Token.Value.(int))
		fmt.Fprintf(filter, "%s(%s,%s,%s)", sqlOp, left, mid, right)
	}
	return nil
}

func HandleReplace(node *parser.ParseNode, filter *strings.Builder, sqlOp string) error {
	//replace会传三个参数，分别是字段名，被替换字段，替换字段,replace也要考虑传的参数也可能存在嵌套情况。比如
	// replace(name,replace(name,'live','life'),substring(addr,3,8))
	if _, ok := node.Children[0].Token.Value.(string); !ok {
		return ErrInvalidInput
	}
	//操作的字段名确定，不递归
	left := node.Children[0].Token.Value.(string)
	var mid, right string
	//后两个参数都有可能存在子树，存在子树则递归生成
	if CheckSubTreeExists(node.Children[1]) {
		mid, _ = applyFilter(node.Children[1])
	} else {
		mid = node.Children[1].Token.Value.(string)
	}
	if CheckSubTreeExists(node.Children[2]) {
		right, _ = applyFilter(node.Children[2])
	} else {
		right = node.Children[2].Token.Value.(string)
	}
	fmt.Fprintf(filter, "%s(%s,%s,%s)", sqlOp, left, mid, right)
	return nil
}

func HandleIndexOf(node *parser.ParseNode, filter *strings.Builder, sqlOp string) error {
	if _, ok := node.Children[1].Token.Value.(string); !ok {
		return ErrInvalidInput
	}
	var left, right string
	if CheckSubTreeExists(node.Children[0]) {
		left, _ = applyFilter(node.Children[0])
	} else {
		left = node.Children[0].Token.Value.(string)
	}
	if CheckSubTreeExists(node.Children[1]) {
		right, _ = applyFilter(node.Children[1])
	} else {
		right = node.Children[1].Token.Value.(string)
	}
	fmt.Fprintf(filter, "%s(%s,%s)", sqlOp, left, right)
	return nil
}

func HandleLength(node *parser.ParseNode, filter *strings.Builder, sqlOp string) error {
	left, leftOk := node.Children[0].Token.Value.(string)
	//可能出现length(trim(toupper(name)))的情况，所以还需判断是否存在子树，存在则进行递归。
	if leftOk {
		if CheckSubTreeExists(node.Children[0]) {
			left, _ = applyFilter(node.Children[0])
		}
		fmt.Fprintf(filter, "%s(%s)", sqlOp, left)
	}
	return nil
}

func HandleContains(node *parser.ParseNode, filter *strings.Builder, sqlOp string) error {
	if _, ok := node.Children[1].Token.Value.(string); !ok {
		return ErrInvalidInput
	}
	//模糊查询
	left := node.Children[0].Token.Value.(string)
	right := fmt.Sprintf("'"+sqlOp+"'", escapeQuote(node.Children[1].Token.Value.(string)))
	fmt.Fprintf(filter, "%s LIKE %s", left, right)
	return nil
}

//处理日期类的操作
func HandDatetime(node *parser.ParseNode, filter *strings.Builder, sqlOp string) error {
	left, leftOk := node.Children[0].Token.Value.(string)
	if leftOk {
		fmt.Fprintf(filter, "%s(%s)", sqlOp, left)
	}
	return nil
}