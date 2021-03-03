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

import (
	"dataapi/internal/pkg/webHelper"
	"strings"
)

// CreateWorkload 创建工作负载
func CreateWorkload(body *strings.Reader) (map[string]interface{}, error) {
	url := AgentRancherHost + AgentVersions + ProjectId + "/workloads"
	return webHelper.HttpToRancherWithBody("POST", url, body, AgentRancherBearerToken)
}

// RedeployWorkload 重新部署工作负载
func RedeployWorkload(deploymentName string) (map[string]interface{}, error) {
	url := AgentRancherHost + AgentVersions + ProjectId + "/workloads/deployment:" + deploymentName + "?action=redeploy"
	return webHelper.HttpToRancherWithBody("POST", url, strings.NewReader(""), AgentRancherBearerToken)
}

// DeleteWorkload 删除工作负载
func DeleteWorkload(deploymentName string) (map[string]interface{}, error){
	url := AgentRancherHost + AgentVersions+"c-wsxm5:p-29rmg/workloads/deployment:mdi-agent:agent-" + deploymentName
	return webHelper.HttpToRancher("DELETE",url,AgentRancherBearerToken)
}

// CreateService 创建服务发现
func CreateService(body *strings.Reader) (map[string]interface{}, error) {
	url := AgentRancherHost + AgentVersions + ProjectId + "/services"
	return webHelper.HttpToRancherWithBody("POST", url, body, AgentRancherBearerToken)
}

//DeleteService 删除服务发现
func DeleteService(deploymentName string) (map[string]interface{},error){
	url := AgentRancherHost + AgentVersions+"c-wsxm5:p-29rmg/services/mdi-agent:service-agent-" + deploymentName
	return webHelper.HttpToRancher("DELETE", url, AgentRancherBearerToken)
}

// CreateIngress 创建负载均衡
func CreateIngress(body *strings.Reader) (map[string]interface{}, error) {
	url := AgentRancherHost + AgentVersions + ProjectId + "/ingresses"
	return webHelper.HttpToRancherWithBody("POST", url, body, AgentRancherBearerToken)
}

//DeleteIngress 删除负载均衡
func DeleteIngress(deploymentName string) (map[string]interface{}, error){
	url := AgentRancherHost + AgentVersions+"c-wsxm5:p-29rmg/ingresses/mdi-agent:ingress-agent-" + deploymentName
	return webHelper.HttpToRancher("DELETE", url, AgentRancherBearerToken)
}

// CreateNamespacedSecrets 在命名空间下创建密钥
//{
//	"annotations": { },
//	"created": "2020-10-20T07:11:32Z",  // 不需要
//	"creatorId": "u-n3phsfy4p6",  // 不需要
//	"data": {
//		"aka": "YWth" 	 // 可用多对 Base64加密
//	},
//	"kind": "Opaque",
//	"labels": {
//		"cattle.io/creator": "norman" // 默认
//	},
//	"name": "",
//	"namespaceId": "mdi-agent",     // 必填
//	"ownerReferences": [ ],
//	"projectId": "c-wsxm5:p-29rmg",
//	"uuid": "d2395bce-b0c3-4990-b9be-b7c20cd8f8f0" // 不需要
//}
func CreateNamespacedSecrets(body *strings.Reader) (map[string]interface{}, error) {
	url := AgentRancherHost + AgentVersions + ProjectId + "/namespacedSecrets"
	return webHelper.HttpToRancherWithBody("POST", url, body, AgentRancherBearerToken)
}

// DeleteNamespacedSecrets 删除Secrets
func DeleteNamespacedSecrets(secretName string) (map[string]interface{}, error) {
	url := AgentRancherHost + AgentVersions + ProjectId + "/namespacedSecrets/" + secretName
	return webHelper.HttpToRancher("DELETE", url, AgentRancherBearerToken)
}

// CreateSecrets
//{
//	"annotations": { },
//	"created": "2020-10-20T07:11:32Z",  // 不需要
//	"creatorId": "u-n3phsfy4p6",  // 不需要
//	"data": {
//		"aka": "YWth" 	 // 可用多对 Base64加密
//	},
//	"kind": "Opaque",
//	"labels": {
//		"cattle.io/creator": "norman" // 默认
//	},
//	"name": "",
//	"namespaceId": "",   //
//	"ownerReferences": [ ],
//	"projectId": "c-wsxm5:p-29rmg",
//	"uuid": "d2395bce-b0c3-4990-b9be-b7c20cd8f8f0" // 不需要
//}
func CreateSecrets(body *strings.Reader) (map[string]interface{}, error) {
	url := AgentRancherHost + AgentVersions + ProjectId + "/secrets"
	return webHelper.HttpToRancherWithBody("POST", url, body, AgentRancherBearerToken)
}

// DeleteSecrets 删除Secrets
func DeleteSecrets(secretName string) (map[string]interface{}, error) {
	url := AgentRancherHost + AgentVersions + ProjectId + "/secrets/" + secretName
	return webHelper.HttpToRancher("DELETE", url, AgentRancherBearerToken)
}

// CreateCert 创建证书
//{
//	"annotations": {},
//  "key":"",
//	"certs": "",
//	"labels": {
//		"cattle.io/creator": "norman"
//	},
//	"name": "whdc-cert",
//	"namespaceId": null,
//	"ownerReferences": [ ],
//	"projectId": "c-wsxm5:p-29rmg",
//	"serialNumber": "1",
//	"subjectAlternativeNames": [ ],
//	"uuid": "a1a29882-a3bf-428d-a394-5233ae2301c6",
//	"version": "3"
//}
func CreateCert(body *strings.Reader) (map[string]interface{}, error) {
	url := AgentRancherHost + AgentVersions + ProjectId + "/certificates"
	return webHelper.HttpToRancherWithBody("POST", url, body, AgentRancherBearerToken)
}

// DeleteCert 删除证书
func DeleteCert(certName string) (map[string]interface{}, error) {
	url := AgentRancherHost + AgentVersions + ProjectId + "/secrets/" + certName
	return webHelper.HttpToRancher("DELETE", url, AgentRancherBearerToken)
}

// UpdateConfigMap 更新指定的configmap
func UpdateConfigMap(name string, body *strings.Reader) (map[string]interface{}, error) {
	url := AgentRancherHost + AgentVersions + ProjectId + "/configMaps/" + name
	return webHelper.HttpToRancherWithBody("PUT", url, body, AgentRancherBearerToken)
}