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

package rancher

const (
	AgentRancherHost        = "https://rancher.wh.digitalchina.com"
	AgentVersions           = "/v3/project/"
	AgentRancherBearerToken = "Bearer token-qw8fn:z6j59xn7nr2gfh6lzk2wxd87w759b2mjb9mp72gdck6hqrc88d5xll"
	ProjectId               = "c-wsxm5:p-29rmg"
	CertificateId           = "mdi-kernel:whdc-cert"
	ImagePullSecrets        = "harbor-yuchonga-brilj"
	KernelNamespaceId       = "mdi-kernel"
	AgentNameSpaceId        = "mdi-agent"
	AgentImageUrl           = "harbor.dev.wh.digitalchina.com/mdi/agent:latest"
	AgentImageName          = "agent"
	AgentContainerPort      = 8081
	AgentWorkloadScale      = 1
	AgentSecretKeyForSysDB  = "dbconnstr"
	AgentDomainFormat       = "mdi-agent-%s.dev.wh.digitalchina.com"
	OathkeeperRulesMapKey   = "oathkeeper-rules"
	WebUIHost               = "https://mdi-webui.dev.wh.digitalchina.com"

)
