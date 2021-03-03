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

package handler

import (
	"dataapi/cmd/kernel/utils"
	"dataapi/internal/kernel/config"
	"dataapi/internal/kernel/db"
	"dataapi/internal/kernel/metadata/convert"
	"dataapi/internal/kernel/metadata/migration"
	"dataapi/internal/kernel/metadata/modeling"
	"dataapi/internal/kernel/metadata/mxgraph"
	"dataapi/internal/kernel/model"
	"dataapi/internal/kernel/tidb"
	"dataapi/internal/pkg/rancher"
	putils "dataapi/internal/pkg/utils"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// CurrentUser 请求中的用户登录名字段
	CurrentUser = "X-User"
)

// GetCurrentLoginUser 当前登录的用户名
func GetCurrentLoginUser(c *gin.Context) string {
	return c.GetString(CurrentUser)
}

// ResponseResult api返回的数据格式
func ResponseResult(success bool, message string, data interface{}) gin.H {
	return gin.H{
		"success": success,
		"message": message,
		"data":    data,
	}
}

//EnvHandler 环境相关的处理业务
type EnvHandler struct{}

//GetProjectEnvs 查询项目包含的所有环境
func (h *EnvHandler) GetProjectEnvs(id string) []ProjectEnvView {
	sdb := db.GetSysDbContext()
	var r []ProjectEnvView
	var envs []model.Environment
	result := sdb.Where(`project_id = ?`, id).Order(`type desc`).Find(&envs)
	if result.Error == nil {
		for _, i := range envs {
			e := ProjectEnvView{
				ID:     i.EnvironmentID,
				Type:   i.Type,
				Domain: i.AgentDomain,
			}
			if i.Type == 0 {
				var uname string
				sdb.Model(model.UserProfile{}).Select("display_name").Where("login_name = ?", i.Owner).Limit(1).Find(&uname)
				e.Name = uname + "的开发环境"
			} else if i.Type == 1 {
				e.Name = "测试环境"
			} else if i.Type == 2 {
				e.Name = "生产环境"
			}
			e.Users = h.getEnvironmentUsers(e.ID)
			r = append(r, e)
		}
	}
	return r
}
func (h *EnvHandler) getEnvironmentUsers(id string) []model.UserProfile {
	sdb := db.GetSysDbContext()
	var users []model.UserProfile

	sdb.Model(model.EnvironmentUser{}).Select("user_profile.*").
		Joins("join user_profile on user_profile.login_name = environment_user.user_login_name and environment_user.environment_id = ?", id).
		Order("environment_user.environment_user_id").
		Find(&users)
	return users
}

//RemoveEnvironmentUser 移除环境的参与成员
func (h *EnvHandler) RemoveEnvironmentUser(c *gin.Context) {
	id := c.Param("id")
	type RemoveUser struct {
		LoginName string `json:"loginname"`
	}
	var user RemoveUser
	c.BindJSON(&user)
	result := db.GetSysDbContext().Where("environment_id = ? and user_login_name = ?", id, user.LoginName).Delete(&model.EnvironmentUser{})
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, ResponseResult(false, "删除失败", nil))
		return
	}
	c.JSON(http.StatusOK, ResponseResult(true, "删除成功", nil))
}

// GetPublishedModel 获取已发布的model
func (h *EnvHandler) GetPublishedModel(c *gin.Context) {
	id := c.Param("id")
	sdb := db.GetSysDbContext()
	env := model.Environment{}
	result := sdb.Where(`environment_id = ?`, id).First(&env)
	if result.Error == nil {
		xml := getEnvSchemaXML(env)
		sdb.Model(&env).Update(`metadata_published`, xml)
		c.String(http.StatusOK, xml)
		return
	}
	c.JSON(http.StatusBadRequest, ResponseResult(false, "无法识别的环境参数", nil))
}

func getEnvSchemaXML(env model.Environment) string {
	converter := &tidb.EnvConverter{
		Owner: env.Owner,
		Connector: config.DbConnector{
			ID:           env.EnvironmentID,
			Host:         env.SQLHost,
			Port:         env.SQLPort,
			UserName:     env.SQLUser,
			Password:     env.SQLPassword,
			DatabaseName: env.SQLDBName,
		},
	}
	return converter.SchemaToModelXML()
}

// GetCurrentModel 获取编辑中的model
func (h *EnvHandler) GetCurrentModel(c *gin.Context) {
	id := c.Param("id")
	sdb := db.GetSysDbContext()
	env := model.Environment{}
	result := sdb.Where(`environment_id = ?`, id).First(&env)
	if result.Error == nil {
		c.String(http.StatusOK, env.MetadataCurrent)
		return
	}
	c.JSON(http.StatusBadRequest, ResponseResult(false, "无法识别的环境参数", nil))
}

// GetCurrent 获取当前画布元数据和model元数据
func (h *EnvHandler) GetCurrent(c *gin.Context) {
	id := c.Param("id")
	db := tidb.GetSysDbContext()
	env := model.Environment{}
	result := db.Where(`environment_id = ?`, id).First(&env)
	if result.Error == nil {
		type model struct {
			GraphCurrent    string `json:"graph_current"`
			MetadataCurrent string `json:"metadata_current"`
		}
		respData := model{
			GraphCurrent:    env.GraphCurrent,
			MetadataCurrent: env.MetadataCurrent,
		}
		c.JSON(http.StatusOK, ResponseResult(true, "", respData))
		return
	}
	c.JSON(http.StatusNotFound, ResponseResult(true, "环境获取失败", nil))
}

// SaveModel 保存编辑中的model
func (h *EnvHandler) SaveModel(c *gin.Context) {
	id := c.Param("id")
	modelXml := c.PostForm("content")
	graphXml := c.PostForm("graph_content")
	sdb := db.GetSysDbContext()
	env := model.Environment{}
	result := sdb.Where(`environment_id = ?`, id).First(&env)
	if result.Error == nil {
		env.MetadataCurrent = modelXml
		env.GraphCurrent = graphXml
		sdb.Model(&env).Updates(env)
		c.JSON(http.StatusOK, ResponseResult(true, "保存成功", nil))
		return
	}
	c.JSON(http.StatusBadRequest, ResponseResult(false, "无法识别的环境参数", nil))
}

// PrePublish model预发布
func (h *EnvHandler) PrePublish(c *gin.Context) {

	_, m := h.getMigration(c.Param("sourceid"))
	output, _ := xml.MarshalIndent(m, "", "  ")

	c.JSON(http.StatusOK, gin.H{"migration": string(output), "list": h.getMigrationPlan(m)})

}

func (h *EnvHandler) getMigrationPlan(m migration.Migration) interface{} {
	type migrationPlan struct {
		Title string
		Items []string
	}
	var list []migrationPlan
	if ce := *m.CreateEntity; len(ce) > 0 {
		plan := migrationPlan{
			Title: "创建实体",
			Items: []string{},
		}
		for _, i := range ce {
			plan.Items = append(plan.Items, i.SchemaNameAttr)
		}
		list = append(list, plan)
	}
	if re := *m.RemoveEntity; len(re) > 0 {
		plan := migrationPlan{
			Title: "删除实体",
			Items: []string{},
		}
		for _, i := range re {
			plan.Items = append(plan.Items, i.SchemaNameAttr)
		}
		list = append(list, plan)
	}
	if ce := *m.ChangeEntity; len(ce) > 0 {
		plan := migrationPlan{
			Title: "变更实体",
			Items: []string{},
		}
		for _, i := range ce {
			plan.Items = append(plan.Items, i.SchemaNameAttr)
		}
		list = append(list, plan)
	}
	if cf := *m.CreateForeignKey; len(cf) > 0 {
		plan := migrationPlan{
			Title: "创建关系",
			Items: []string{},
		}
		for _, i := range cf {
			plan.Items = append(plan.Items, i.SchemaNameAttr)
		}
		list = append(list, plan)
	}
	if rf := *m.RemoveForeignKey; len(rf) > 0 {
		plan := migrationPlan{
			Title: "删除关系",
			Items: []string{},
		}
		for _, i := range rf {
			plan.Items = append(plan.Items, i.SchemaNameAttr)
		}
		list = append(list, plan)
	}
	if cf := *m.ChangeForeignKey; len(cf) > 0 {
		plan := migrationPlan{
			Title: "变更关系",
			Items: []string{},
		}
		for _, i := range cf {
			plan.Items = append(plan.Items, i.SchemaNameAttr)
		}
		list = append(list, plan)
	}
	return list
}

// Publish 执行迁移发布
func (h *EnvHandler) Publish(c *gin.Context) {
	result := h.SubmitMigration(c.Param("sourceid"), GetCurrentLoginUser(c))
	c.JSON(http.StatusOK, gin.H{"successed": result.SuccessedCnt, "failed": result.FailedCnt})
}

// CreateEnv 创建环境
func (h *EnvHandler) CreateEnv(c *gin.Context) {
	type body struct {
		ProjectID string   `json:"projectid"`
		Creator   string   `json:"owner"`
		Entities  []string `json:"entities"`
	}
	var p body
	if c.BindJSON(&p) == nil {
		mx, mod := h.getCDMGraphXML(p.Entities, p.Creator)
		env := model.Environment{
			ProjectID:         p.ProjectID,
			Type:              0, //开发环境
			Owner:             p.Creator,
			GraphCurrent:      mx,
			MetadataCurrent:   mod,
			MetadataPublished: "",
		}
		CreateEnv(env)
		c.JSON(http.StatusOK, ResponseResult(true, "创建成功", nil))
		return
	}
	c.JSON(http.StatusBadRequest, ResponseResult(false, "创建失败", nil))
}

func (h *EnvHandler) getCDMGraphXML(cdms []string, creator string) (string, string) {
	cells := []mxgraph.MxCell{
		mxgraph.MxCell{ID: "0"},
		mxgraph.MxCell{ID: "1", Parent: "0"},
		mxgraph.MxCell{ID: "0.0", Style: "画布层", Parent: "1", D: &mxgraph.MxCellD{As: "value", Value: `{"displayName":"Layer1","visible":true,"selected":true}`}},
	}
	entities := []modeling.Entity{}
	if len(cdms) > 0 {
		sdb := db.GetSysDbContext()
		for i, e := range cdms {
			var entity model.CDMEntity
			res := sdb.Model(model.CDMEntity{}).Where("entity_id = ?", e).Limit(1).Find(&entity)
			if res.Error == nil {
				cell := mxgraph.MxCell{
					ID:     strconv.Itoa(i + 2),
					Style:  "实体",
					Vertex: "1",
					Parent: "0.0",
					D:      &mxgraph.MxCellD{As: "value"},
					MxGeometry: &mxgraph.MxGeometry{
						X:           34 + (220 * i),
						Y:           22,
						Width:       200,
						Height:      132,
						As:          "geometry",
						MxRectangle: &mxgraph.MxRectangle{Width: 200, Height: 30, As: "alternateBounds"},
					},
				}
				val, _ := json.Marshal(mxgraph.MxCellValue{SchemaName: entity.SchemaName, DisplayName: entity.DisplayName, IsOriginal: true, Clustered: "clustered", UniqueConstraints: []mxgraph.MxCellIndex{}, Indexes: []mxgraph.MxCellIndex{}})
				cell.D.Value = string(val)
				cells = append(cells, cell)

				enti := modeling.Entity{
					SchemaNameAttr:  entity.SchemaName,
					ClusteredAttr:   "",
					DisplayNameAttr: entity.DisplayName,
					IsOriginalAttr:  true,
					Fields:          []modeling.Field{},
				}

				var fields []model.CDMField
				sdb.Model(model.CDMField{}).Where("entity_name = ?", entity.SchemaName).Order("sort").Find(&fields)
				for j, f := range fields {
					c := mxgraph.MxCell{
						ID:     fmt.Sprintf("0.0.%d.%d", i, j),
						Style:  "字段",
						Vertex: "1",
						Parent: cell.ID,
						D:      &mxgraph.MxCellD{As: "value"},
						MxGeometry: &mxgraph.MxGeometry{
							Y:      28 + (j * 26),
							Width:  200,
							Height: 26,
							As:     "geometry",
						},
					}
					cv := mxgraph.MxCellValue{SchemaName: f.SchemaName, DisplayName: f.DisplayName, IsOriginal: true, IsNull: f.IsNullable, DataType: f.DataType, PrimaryKey: f.IsPrimary, Unique: f.IsPrimary}

					fie := modeling.Field{SchemaNameAttr: f.SchemaName, DisplayNameAttr: f.DisplayName, IsOriginalAttr: true, DataTypeAttr: f.DataType, IsNullAttr: f.IsNullable}
					switch f.DataType {
					case "integer":
						cv.IntegerOption = &mxgraph.MxCellTypeOption{Length: f.Length, AutoIncrement: f.IsAutoIncr}
						fie.TypeOption = &modeling.TypeOption{AutoIncrementAttr: f.IsAutoIncr, LengthAttr: f.Length}
					case "decimal":
						cv.DecimalOption = &mxgraph.MxCellTypeOption{Length: f.Length, Precision: f.Precision}
						fie.TypeOption = &modeling.TypeOption{PrecisionAttr: f.Precision, LengthAttr: f.Length}
					case "string":
						cv.StringOption = &mxgraph.MxCellTypeOption{Length: f.Length}
						fie.TypeOption = &modeling.TypeOption{LengthAttr: f.Length}
					default:
						cv.IntegerOption = &mxgraph.MxCellTypeOption{}
						fie.TypeOption = &modeling.TypeOption{}
					}
					v, _ := json.Marshal(cv)
					c.D.Value = string(v)
					cells = append(cells, c)
					enti.Fields = append(enti.Fields, fie)
				}
				enti.UniqueConstraints = &[]modeling.UniqueConstraint{}
				enti.Indexes = &[]modeling.Index{}
				entities = append(entities, enti)
			}
		}
	}
	graph := mxgraph.MxGraphModel{
		Cells: cells,
	}
	output1, err := xml.MarshalIndent(graph, "", "  ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	m := &modeling.Model{
		CollationAttr:       "Chinese_PRC_CI_AS",
		ModelingVersionAttr: "1.0",
		OwnerAttr:           creator,
		Entities:            entities,
		ForeignKeys:         &[]modeling.ForeignKey{},
	}
	output2, err := xml.MarshalIndent(m, "", "  ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return string(output1), xml.Header + string(output2)
}

// getMigration 获取迁移描述内容
func (h *EnvHandler) getMigration(sourceid string) (model.Environment, migration.Migration) {
	sdb := db.GetSysDbContext()
	env := model.Environment{}
	result := sdb.Where(`environment_id = ?`, sourceid).First(&env)
	if result.Error == nil {
		sourceModeling := modeling.Model{}
		xml.Unmarshal([]byte(env.MetadataPublished), &sourceModeling)

		targetModeling := modeling.Model{}
		xml.Unmarshal([]byte(env.MetadataCurrent), &targetModeling)

		mig, _ := convert.ModelingToMigration(&sourceModeling, &targetModeling)
		//output, _ := xml.MarshalIndent(mig, "", "  ")

		return env, mig
	}
	return env, migration.Migration{}
}

// submitMigration 执行迁移
func (h *EnvHandler) SubmitMigration(id string, publisher string) tidb.EnvConverter {
	e, m := h.getMigration(id)
	converter := tidb.EnvConverter{
		Owner: e.Owner,
		Connector: config.DbConnector{
			ID:           e.EnvironmentID,
			Host:         e.SQLHost,
			Port:         e.SQLPort,
			UserName:     e.SQLUser,
			Password:     e.SQLPassword,
			DatabaseName: e.SQLDBName,
		},
	}
	output, _ := xml.MarshalIndent(m, "", "  ")
	res := converter.MigrationToSQL(output)
	sdb := db.GetSysDbContext()
	log := model.EnvironmentHistory{
		EnvironmentID:  e.EnvironmentID,
		Publisher:      publisher,
		PublishOn:      putils.Time(time.Now()),
		MetadataSource: e.MetadataPublished,
		MetadataTarget: e.MetadataCurrent,
		Actions:        uint32(converter.SuccessedCnt + converter.FailedCnt),
		FailedActions:  uint32(converter.FailedCnt),
	}
	resx, _ := xml.MarshalIndent(res, "", "  ")
	log.ActionResult = string(resx)
	sdb.Create(&log)
	mx := converter.SchemaToModelXML()
	sdb.Model(&e).Updates(map[string]interface{}{"metadata_published": mx, "metadata_current": ""})
	////创建agent转发规则
	//utils.CreateOathkeeperRules(rancher.AgentNameSpaceId + `-` + e.AgentKey)
	////更新oathkeeper的规则
	//rules, err := (&SysHandler{}).BuildOathkeeperRules()
	//if err == nil {
	//	data, _ := json.Marshal(rules)
	//	body := strings.NewReader(`{
	//		"data": {
	//			"rules.json": "` + string(data) + `"
	//		   }
	//		}`)
	//	rancher.UpdateConfigMap(rancher.KernelNamespaceId+`:`+rancher.OathkeeperRulesMapKey, body)
	//}
	//重启agent容器
	_, err := rancher.RedeployWorkload(rancher.AgentNameSpaceId + `:agent-` + e.AgentKey)
	if err != nil {
		fmt.Println(err)
	}
	return converter
}

// randomEnvKey 生成环境的识别码
func randomEnvKey(len int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}

// CreateEnv 创建环境
func CreateEnv(env model.Environment) string {
	rkey := strings.ToLower(randomEnvKey(10))
	//1-持久化数据
	sdb := db.GetSysDbContext()
	env.EnvironmentID = putils.GetUUID()
	env.SQLHost = config.AgentDbHost
	port, _ := strconv.Atoi(config.AgentDbPort)
	env.SQLPort = port
	env.SQLDBName = "env_" + rkey
	env.SQLUser = config.AgentDbUser
	env.SQLPassword = config.AgentDbPassword
	env.SQLSchema = "public"
	env.AgentKey = rkey
	result := sdb.Create(&env)
	if result.Error == nil {
		user := model.EnvironmentUser{
			EnvironmentID: env.EnvironmentID,
			UserLoginName: env.Owner,
			Permission:    1,
		}
		sdb.Create(&user)
		//2-创建数据库和默认表
		ctx, _ := sdb.DB()
		res, err := ctx.Exec(`CREATE DATABASE ` + env.SQLDBName + ` CHARACTER SET 'utf8mb4';`)
		if err != nil {
			fmt.Println(res)
		}
		edb := db.GetEnvDbContext(config.DbConnector{
			ID:           env.EnvironmentID,
			Host:         env.SQLHost,
			Port:         env.SQLPort,
			UserName:     env.SQLUser,
			Password:     env.SQLPassword,
			DatabaseName: env.SQLDBName,
		})

		edb.Exec(`
	CREATE TABLE sys_renamelogs
	(
		id varchar(36),
		target varchar(50) NOT NULL,
		create_on datetime NOT NULL,
		before varchar(500),
		after varchar(500),
		table_name varchar(100),
		PRIMARY KEY (id)
	);
	CREATE TABLE sys_schemainfo
	(
		schema_cate tinyint(255) NOT NULL,
		schema_name varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
		is_original bit(1) NULL DEFAULT NULL,
		display_name varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NULL DEFAULT NULL,
		extension text CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NULL,
		create_on datetime(0) NULL DEFAULT NULL,
		update_on datetime(0) NULL DEFAULT NULL,
		PRIMARY KEY (schema_name, schema_cate) USING BTREE
	  );
	`)
		//3-部署agent
		ehost := deployAgent(rkey, env)
		//4-更新数据
		xml := getEnvSchemaXML(env)
		sdb.Model(&env).Updates(map[string]interface{}{"agent_domain": ehost, "metadata_published": xml})
		//5-创建对应的oathkeeper规则
		agentName := "mdi-agent-"+env.AgentKey
		utils.CreateOathkeeperRules(agentName)
		//6-创建oauth2.0客户端
		client := model.OauthClient{
			ClientID:      env.EnvironmentID,
			ClientName:    agentName,
			ClientSecret:  agentName,
			Owner:         env.Owner,
			Scope:         "",
			CreatedAt:     time.Time{},
			UpdatedAt:     time.Time{},
			AccessToken:   "",
			OauthToken:    "",
			ProjectId:     env.ProjectID,
			EnvironmentId: env.EnvironmentID,
		}
		utils.CreateClient(client)
		return env.EnvironmentID
	}
	return ""
}

// DeleteEnv 删除环境
func (h *EnvHandler) DeleteEnv(c *gin.Context) {
	id := c.Param("id")
	env := model.Environment{}
	oathkeeper := model.OathkeeperRule{}
	envUser := model.EnvironmentUser{
		EnvironmentID: id,
	}

	db := db.GetSysDbContext()
	result := db.Where("environment_id = ?",id).Find(&env)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest,ResponseResult(false,"删除失败",nil))
		return
	}

	agentName := "mdi-agent-"+env.AgentKey
	oathkeeper.ID = agentName
	//删除信息及rancher部署
	db.Delete(&env)
	db.Delete(&envUser)
	db.Delete(&oathkeeper)
	DeleteAgent(env.AgentKey)
	utils.DeleteClient(id)

	ctx, _ := db.DB()
	res, err := ctx.Exec(`DROP DATABASE ` + env.SQLDBName)
	if err != nil {
		c.JSON(http.StatusBadRequest,ResponseResult(false,"删除失败",res))
		return
	}
	c.JSON(http.StatusOK,ResponseResult(true,"删除成功",nil))
}

// deployAgent 部署环境实例
func deployAgent(rkey string, env model.Environment) string {
	workloadName := "agent-" + rkey
	workSelector := "deployment-mdi-agent-" + rkey
	ehost := fmt.Sprintf(rancher.AgentDomainFormat, rkey)
	//1-创建workload
	body := strings.NewReader(`{
		    "baseType": "workload",
		    "containers":[
		        {
					"environment": {
						"Environment_Host": "` + ehost + `",
						"Environment_Id": "` + env.EnvironmentID + `",
						"CORS_ORIGIN": "` + rancher.WebUIHost + `",
						"CORS_ALLOW_HEADER": "Authorization,Content-Length,Accept,Origin,X-Requested-With,Content-Type,Pragma"
					},
					"environmentFrom":[
						{"optional":false,"prefix":"","source":"secret","sourceName":"` + rancher.AgentSecretKeyForSysDB + `","type":"/v3/project/schemas/environmentFrom"}
					],
		            "image":"` + rancher.AgentImageUrl + `",
		            "name":"` + rancher.AgentImageName + `",
					"ports":[
						{
							"containerPort":` + strconv.Itoa(rancher.AgentContainerPort) + `,
							"name":"app",
							"protocol":"TCP"
						}
		            ]
		        }
		    ],
		    "dnsPolicy":"ClusterFirst",
		    "imagePullSecrets":[
		        {
		            "name":"` + rancher.ImagePullSecrets + `"
		        }
		    ],
		    "labels":{
		        "workloadselector":"` + workSelector + `"
		    },
		    "name": "` + workloadName + `",
		    "namespaceId":"` + rancher.AgentNameSpaceId + `",
		    "projectId":"` + rancher.ProjectId + `",
		    "scale":` + strconv.Itoa(rancher.AgentWorkloadScale) + `,
		    "selector":{
		        "matchLabels":{
		            "workloadselector":"` + workSelector + `"
		        }
		    }
		}`)
	rancher.CreateWorkload(body)
	//2-创建service
	serviceName := "service-agent-" + rkey
	body = strings.NewReader(`{
			"name": "` + serviceName + `",
			"namespaceId": "` + rancher.AgentNameSpaceId + `",
			"ownerReferences": [ ],
			"ports": [
				{
					"nodePort": 0,
					"port": ` + strconv.Itoa(rancher.AgentContainerPort) + `,
					"protocol": "TCP",
					"targetPort": ` + strconv.Itoa(rancher.AgentContainerPort) + `,
					"type": "/v3/project/schemas/servicePort"
				}
			],
			"projectId": "` + rancher.ProjectId + `",
			"selector": {
				"workloadselector": "` + workSelector + `"
			}
		}`)
	rancher.CreateService(body)
	//3-创建ingress
	serviceID := "mdi-agent:" + serviceName
	ingressName := "ingress-agent-" + rkey
	body = strings.NewReader(`{
			    "name":"` + ingressName + `",
			    "namespaceId":"` + rancher.AgentNameSpaceId + `",
			    "projectId":"` + rancher.ProjectId + `",
			    "rules":[
			        {
			            "host":"` + ehost + `",
						"paths":[
			                {
			                    "pathType":"ImplementationSpecific",
			                    "serviceId":"` + serviceID + `",
			                    "targetPort":` + strconv.Itoa(rancher.AgentContainerPort) + `
			                }
			            ]
			        }
			    ],
			    "tls":[
			        {
			            "certificateId":"` + rancher.CertificateId + `",
			            "hosts":[
			                "` + ehost + `"
			            ]
			        }
			    ]
			}`)

	rancher.CreateIngress(body)

	return ehost
}

//DeleteAgent 删除环境实例
func DeleteAgent(agentKey string) {
	rancher.DeleteIngress(agentKey)
	rancher.DeleteService(agentKey)
	rancher.DeleteWorkload(agentKey)
}


// GetAgentDomain
func (h *EnvHandler) GetAgentDomain(c *gin.Context) {
	owner := GetCurrentLoginUser(c)
	var envs []model.Environment
	var r []AgentDomainView
	dbContext := tidb.GetSysDbContext()
	result := dbContext.Where("owner = ?", owner).Find(&envs)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error})
		return
	} else {
		for _, i := range envs{
			if i.AgentDomain != ""{
				e := AgentDomainView{
					AgentDomain: i.AgentDomain,
				}
				var projectName string
				dbContext.Model(model.Project{}).Select("name").Where("project_id=?",i.ProjectID).Limit(1).Find(&projectName)
				if i.Type == 0 {
					var uname string
					dbContext.Model(model.UserProfile{}).Select("display_name").Where("login_name = ?", i.Owner).Limit(1).Find(&uname)
					e.DisplayName = projectName+": "+uname + "的开发环境(" + i.AgentDomain + ")"
				} else if i.Type == 1 {
					e.DisplayName = projectName+": "+"测试环境(" + i.AgentDomain + ")"
				} else if i.Type == 2 {
					e.DisplayName = projectName+": "+"生产环境(" + i.AgentDomain + ")"
				}
				r = append(r, e)
			}
		}
		c.JSON(http.StatusOK, r)
		return
	}
}

// ExportModel 导出环境的模型
func (h *EnvHandler) ExportModel(c *gin.Context) {
	id := c.Param("id")
	sdb := db.GetSysDbContext()
	env := model.Environment{}
	result := sdb.Where(`environment_id = ?`, id).First(&env)
	if result.Error == nil {
		c.Writer.Header().Add("Content-Disposition", "attachment; filename=metadata.xml")
		c.Data(http.StatusOK, "application/octet-stream", []byte(env.MetadataPublished))
		return
	}
	c.JSON(http.StatusBadRequest, ResponseResult(false, "无法识别的环境参数", nil))
	return
}

// ImportModel 导入外部模型
func (h *EnvHandler) ImportModel(c *gin.Context) {
	file, _, err := c.Request.FormFile("metadata")
	if err != nil {
		c.String(http.StatusBadRequest, "无效文件")
		return
	}
	data, _ := ioutil.ReadAll(file)
	id := c.Param("id")
	sdb := db.GetSysDbContext()
	env := model.Environment{
		EnvironmentID: id,
	}
	result := sdb.Model(&env).Update("metadata_current", string(data))
	if result.Error == nil {
		c.JSON(http.StatusOK, ResponseResult(true, "导入成功", nil))
		return
	}
	c.JSON(http.StatusOK, ResponseResult(false, "导入失败", nil))
	return
}

// CopyEnv 复制环境
func (h *EnvHandler) CopyEnv(c *gin.Context) {
	source := c.Param("sourceid")
	sdb := db.GetSysDbContext()
	env := model.Environment{}
	result := sdb.Where(`environment_id = ?`, source).First(&env)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, ResponseResult(false, "无法识别的环境参数", nil))
		return
	}
	target := c.Param("targetid")
	if target == "" {
		e := model.Environment{
			ProjectID:         env.ProjectID,
			Type:              0, //开发环境
			Owner:             GetCurrentLoginUser(c),
			MetadataCurrent:   env.MetadataPublished,
			MetadataPublished: "",
		}
		target = CreateEnv(e)
	} else {
		e := model.Environment{}
		result = sdb.Where(`environment_id = ?`, target).First(&e)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, ResponseResult(false, "无法识别的目标环境", nil))
			return
		}
		h.clearEnvDatabase(db.GetEnvDbContext(config.DbConnector{
			ID:           e.EnvironmentID,
			Host:         e.SQLHost,
			Port:         e.SQLPort,
			UserName:     e.SQLUser,
			Password:     e.SQLPassword,
			DatabaseName: e.SQLDBName,
		}))
		sdb.Model(&e).Updates(model.Environment{MetadataPublished: "", MetadataCurrent: env.MetadataPublished})
	}
	//执行迁移
	h.SubmitMigration(target, GetCurrentLoginUser(c))
	c.JSON(http.StatusOK, ResponseResult(true, "复制成功", nil))
}

// ResetEnv 重置环境
func (h *EnvHandler) ResetEnv(c *gin.Context) {
	id := c.Param("id")
	sdb := db.GetSysDbContext()
	env := model.Environment{}
	result := sdb.Where(`environment_id = ?`, id).First(&env)
	if result.Error == nil {
		edb := db.GetEnvDbContext(config.DbConnector{
			ID:           env.EnvironmentID,
			Host:         env.SQLHost,
			Port:         env.SQLPort,
			UserName:     env.SQLUser,
			Password:     env.SQLPassword,
			DatabaseName: env.SQLDBName,
		})
		clear := h.clearEnvDatabase(edb)
		if clear {
			sdb.Model(&env).Update(`metadata_published`, "")
			rancher.RedeployWorkload(rancher.AgentNameSpaceId + `:agent-` + strings.Split(env.SQLDBName, "_")[1])
			c.JSON(http.StatusOK, ResponseResult(true, "重置成功", nil))
		} else {
			c.JSON(http.StatusOK, ResponseResult(false, "重置失败", nil))
		}
		return
	}
	c.JSON(http.StatusBadRequest, ResponseResult(false, "无法识别的环境参数", nil))
}

func (h *EnvHandler) clearEnvDatabase(db *sql.DB) bool {
	tx, _ := db.Begin()
	_, err := tx.Exec("DROP SCHEMA public CASCADE;CREATE SCHEMA public AUTHORIZATION postgres;")
	if err == nil {
		tx.Commit()
		return true
	}
	tx.Rollback()
	return false
}
