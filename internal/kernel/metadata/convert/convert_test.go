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

package convert

import (
	"dataapi/internal/kernel/metadata/modeling"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestGetSimplifiedRenameAction(t *testing.T) {
	targetModeling := &modeling.Model{}

	targetOneModelingPath, _ := filepath.Abs("TargetOneModeling.xml")
	targetOneFile, err := os.Open(targetOneModelingPath)
	if err != nil {
		log.Fatal(err)
	}
	targetOneData, err := ioutil.ReadAll(targetOneFile)
	targetOneModeling := modeling.Model{}

	err = xml.Unmarshal(targetOneData, &targetOneModeling)
	if err != nil {
		fmt.Println("targetOneModeling 读取失败")
	}

	targetTwoModelingPath, _ := filepath.Abs("TargetTwoModeling.xml")
	targetTwoFile, err := os.Open(targetTwoModelingPath)
	if err != nil {
		log.Fatal(err)
	}
	targetTwoData, err := ioutil.ReadAll(targetTwoFile)
	targetTwoModeling := modeling.Model{}
	if err != nil {
		fmt.Println("TargetTwoModeling 读取失败")
	}

	err = xml.Unmarshal(targetTwoData, &targetTwoModeling)
	if err != nil {
		fmt.Println("TargetTwoModeling 解析失败")
	}

	sourceModelingPath, _ := filepath.Abs("SourceModeling.xml")
	sourceFile, err := os.Open(sourceModelingPath)
	if err != nil {
		log.Fatal(err)
	}
	sourceData, err := ioutil.ReadAll(sourceFile)
	sourceModeling := modeling.Model{}

	err = xml.Unmarshal(sourceData, &sourceModeling)
	if err != nil {
		fmt.Println("SourceModeling 读取失败")
	}

	migrationZero, err0 := ModelingToMigration(targetModeling, &sourceModeling)

	outputZero, err1 := xml.MarshalIndent(migrationZero, "", "  ")
	if err0 == nil && err1 == nil {
		os.Stdout.Write([]byte(xml.Header))
		os.Stdout.Write(outputZero)
		fmt.Println("target0-->source 成功")
	}

	migrationOne, err0 := ModelingToMigration(&targetOneModeling, &sourceModeling)

	outputOne, err1 := xml.MarshalIndent(migrationOne, "", "  ")
	if err0 == nil && err1 == nil {
		os.Stdout.Write([]byte(xml.Header))
		os.Stdout.Write(outputOne)
		fmt.Println("target1-->source 成功")
	}

	migrationTwo, err0 := ModelingToMigration(&targetTwoModeling, &sourceModeling)

	outputTwo, err1 := xml.MarshalIndent(migrationTwo, "", "  ")
	if err0 == nil && err1 == nil {
		os.Stdout.Write([]byte(xml.Header))
		os.Stdout.Write(outputTwo)
		fmt.Println("target2-->source 成功")
	}
}
