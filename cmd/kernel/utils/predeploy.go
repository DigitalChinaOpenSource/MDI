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

package utils

import (
	"dataapi/internal/kernel/db"
	"dataapi/internal/kernel/model"
	"dataapi/internal/pkg/rancher"
	"fmt"
	"os"
)

//ReDeployAgent 重新部署agent
func ReDeployAgent() (int16, int16) {
	mode := os.Getenv("mode")
	if mode != "prod" {
		return 0, 0
	}
	fmt.Println("[Init] Ready to redeploy all agents...")
	sdb := db.GetSysDbContext()
	var envList []model.Environment
	sdb.Model(model.Environment{}).Find(&envList)
	var failCount, successCount int16 = 0, 0
	for _, env := range envList {
		workload := rancher.AgentNameSpaceId + `:agent-` + env.AgentKey
		_, err := rancher.RedeployWorkload(workload)
		if err == nil {
			successCount++
			fmt.Printf("[Init] Agent:%s redeploy success\n", workload)
		} else {
			failCount++
			fmt.Errorf("[Init] Agent:%s redeploy failed\n%s", workload, err)
		}
	}
	fmt.Println("[Init] Redeploy agent complete.")
	return successCount, failCount
}
