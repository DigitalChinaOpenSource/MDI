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

package tidb

import (
	"dataapi/internal/kernel/config"
	"dataapi/internal/kernel/metadata/migration"
	"dataapi/internal/kernel/metadata/modeling"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

//EnvConverter ..
type EnvConverter struct {
	Connector    config.DbConnector
	Owner        string
	Db           *sql.DB
	SuccessedCnt int16
	FailedCnt    int16
}

// MigrationToSQL ...
func (e *EnvConverter) MigrationToSQL(xd []byte) migration.Migration {
	m := migration.Migration{}
	err := xml.Unmarshal(xd, &m)
	if err != nil {
		log.Fatal(err)
	}

	e.Db = GetEnvDbContext(e.Connector)
	//defer e.Db.Close()

	//createEntity部分
	if m.CreateEntity != nil {
		e.parseCreateEntity(*m.CreateEntity)
	}

	//removeEntity部分
	if m.RemoveEntity != nil {
		e.parseRemoveEntity(*m.RemoveEntity)
	}

	//changeEntity部分
	if m.ChangeEntity != nil {
		e.parseChangeEntity(*m.ChangeEntity)
	}

	//createForeignKey部分
	if m.CreateForeignKey != nil {
		e.parseCreateForeign(*m.CreateForeignKey)
	}

	//removeForeignKey部分
	if m.RemoveForeignKey != nil {
		e.parseRemoveForeign(*m.RemoveForeignKey)
	}

	//changeForeignKey部分
	if m.ChangeForeignKey != nil {
		e.parseChangeForeign(*m.ChangeForeignKey)
	}

	//修改名称历史
	if m.RenameActions != nil {
		e.parseRenameActions(m.RenameActions)
	}

	return m
	// output, _ := xml.MarshalIndent(m, "", "  ")
	// return xml.Header + string(output)
}

func (e *EnvConverter) parseCreateEntity(actions []migration.CreateEntityAction) {
	var sql strings.Builder
	var comment strings.Builder
	for idx, entity := range actions {
		sql.Reset()
		comment.Reset()
		//sql.WriteString(`----------------------创建表：` + entity.SchemaNameAttr + `-----------------------`)
		//sql.WriteString("\n")
		sql.WriteString("CREATE TABLE `" + entity.SchemaNameAttr + "`")
		sql.WriteString("\n(\n")
		//字段清单
		for i := 0; i < len(entity.Fields); i++ {
			f := entity.Fields[i]
			if i == 0 {
				sql.WriteString(`	`)
			} else {
				sql.WriteString(`	,`)
			}
			sql.WriteString(e.parseFieldStruct(f))
			sql.WriteString("\n")
			comment.WriteString(fmt.Sprintf("INSERT INTO sys_schemainfo VALUES (2,'%s',%s,'%s',null,now(),null);\n", entity.SchemaNameAttr+"_"+f.SchemaNameAttr, strconv.FormatBool(f.IsOriginalAttr), f.DisplayNameAttr))
		}
		//主键和索引
		var ind strings.Builder
		if entity.Indexes != nil {
			for _, i := range *entity.Indexes {
				if i.IsPrimaryAttr == true {
					sql.WriteString(`	,PRIMARY KEY ` + e.parseIndexFields(i))
					sql.WriteString("\n")
				} else {
					ind.WriteString(` , INDEX ` + i.SchemaNameAttr + e.parseIndexFields(i) + ` USING BTREE;`)
					ind.WriteString("\n")
					comment.WriteString(fmt.Sprintf("INSERT INTO sys_schemainfo VALUES (3,'%s',0,'%s',null,now(),null);\n", i.SchemaNameAttr, i.DisplayNameAttr))
				}
			}
		}
		//唯一约束
		if entity.UniqueConstraints != nil {
			for _, c := range *entity.UniqueConstraints {
				sql.WriteString(`	,` + e.parseUniqueCostraintStruct(c))
				sql.WriteString("\n")
				comment.WriteString(fmt.Sprintf("INSERT INTO sys_schemainfo VALUES (4,'%s',0,'%s',null,now(),null);\n", c.SchemaNameAttr, c.DisplayNameAttr))
			}
		}
		sql.WriteString(");\n")
		//表备注
		sql.WriteString(fmt.Sprintf("INSERT INTO sys_schemainfo VALUES (1,'%s',%s,'%s',null,now(),null);\n", entity.SchemaNameAttr, strconv.FormatBool(entity.IsOriginalAttr), entity.DisplayNameAttr))
		sql.WriteString(comment.String())
		//执行脚本
		actions[idx].Result = e.runSQL(sql.String())
	}
}

func (e *EnvConverter) parseRemoveEntity(actions []migration.RemoveItem) {
	var sql strings.Builder
	for idx, entity := range actions {
		sql.Reset()
		//sql.WriteString(`----------------------删除表：` + entity.SchemaNameAttr + `-----------------------`)
		//sql.WriteString("\n")
		sql.WriteString("DROP TABLE `" + entity.SchemaNameAttr + "`;")
		sql.WriteString(fmt.Sprintf("DELETE FROM sys_schemainfo WHERE schema_cate=1 and schema_name='%s';", entity.SchemaNameAttr))
		sql.WriteString("\n")
		// fmt.Println(sql.String())
		//执行脚本
		actions[idx].Result = e.runSQL(sql.String())
	}
}

func (e *EnvConverter) parseChangeEntity(actions []migration.ChangeEntityAction) {
	var sql strings.Builder
	for idx, entity := range actions {
		sql.Reset()
		//sql.WriteString(`----------------------修改表：` + entity.SchemaNameAttr + `-----------------------`)
		//sql.WriteString("\n")
		//修改表名
		if entity.SchemaNameAttr != entity.NewSchemaNameAttr {
			sql.WriteString("RENAME TABLE `" + entity.SchemaNameAttr + "`  TO `" + entity.NewSchemaNameAttr + "`;")
			sql.WriteString("\n")
		}
		//修改表显示名
		if entity.DisplayNameAttr != entity.NewDisplayNameAttr {
			sql.WriteString(fmt.Sprintf("UPDATE sys_schemainfo SET display_name='%s',update_on=now() WHERE schema_cate=1 and schema_name='%s';\n", entity.NewDisplayNameAttr, entity.NewSchemaNameAttr))
		}
		// fmt.Println(sql.String())
		actions[idx].Result = e.runSQL(sql.String())

		sql.Reset()
		var s string
		//新增字段
		if entity.NewFields != nil {
			nf := *entity.NewFields
			for i, f := range nf {
				s = "ALTER TABLE `" + entity.NewSchemaNameAttr + "` ADD COLUMN " + e.parseFieldStruct(f) + `;`
				s += fmt.Sprintf("INSERT INTO sys_schemainfo VALUES (2,'%s',%s,'%s',null,now(),null);", entity.NewSchemaNameAttr+"_"+f.SchemaNameAttr, strconv.FormatBool(f.IsOriginalAttr), f.DisplayNameAttr)
				sql.WriteString(s)
				sql.WriteString("\n")
				nf[i].Result = e.runSQL(s)
			}
			actions[idx].NewFields = &nf
		}
		//新增唯一约束
		if entity.NewUniqueConstraints != nil {
			nuc := *entity.NewUniqueConstraints
			for i, c := range nuc {
				s = "ALTER TABLE `" + entity.NewSchemaNameAttr + "` ADD " + e.parseUniqueCostraintStruct(c) + `;`
				s += fmt.Sprintf("INSERT INTO sys_schemainfo VALUES (4,'%s',0,'%s',null,now(),null);", c.SchemaNameAttr, c.DisplayNameAttr)
				sql.WriteString(s)
				sql.WriteString("\n")
				nuc[i].Result = e.runSQL(s)
			}
			actions[idx].NewUniqueConstraints = &nuc
		}
		//新增索引
		if entity.NewIndexes != nil {
			ni := *entity.NewIndexes
			for x, i := range ni {
				if i.IsPrimaryAttr == true {
					//tidb不支持修改和删除主键
					ni[x].Result = &migration.ActionResult{SuccessAttr: true}
					continue
				} else {
					s = `CREATE INDEX ` + i.SchemaNameAttr + " ON `" + entity.NewSchemaNameAttr + "`" + e.parseIndexFields(i) + ` USING BTREE;`
					s += fmt.Sprintf("INSERT INTO sys_schemainfo VALUES (3,'%s',0,'%s',null,now(),null);", i.SchemaNameAttr, i.DisplayNameAttr)
				}
				sql.WriteString(s)
				sql.WriteString("\n")
				ni[x].Result = e.runSQL(s)
			}
			actions[idx].NewIndexes = &ni
		}
		//删除字段、索引、唯一约束
		if entity.RemoveFields != nil {
			rf := *entity.RemoveFields
			for x, f := range rf {
				s = "ALTER TABLE `" + entity.NewSchemaNameAttr + "` DROP COLUMN `" + f.SchemaNameAttr + "`;"
				s += fmt.Sprintf("DELETE FROM sys_schemainfo WHERE schema_cate=2 and schema_name='%s';", entity.NewSchemaNameAttr+"_"+f.SchemaNameAttr)
				sql.WriteString(s)
				sql.WriteString("\n")
				rf[x].Result = e.runSQL(s)
			}
			actions[idx].RemoveFields = &rf
		}
		if entity.RemoveIndexes != nil {
			ri := *entity.RemoveIndexes
			for x, i := range ri {
				if i.SchemaNameAttr == "" {
					continue
				}
				s = "ALTER TABLE `" + entity.NewSchemaNameAttr + "` DROP INDEX " + i.SchemaNameAttr + `;`
				s += fmt.Sprintf("DELETE FROM sys_schemainfo WHERE schema_cate=3 and schema_name='%s';", i.SchemaNameAttr)
				sql.WriteString(s)
				sql.WriteString("\n")
				ri[x].Result = e.runSQL(s)
			}
			actions[idx].RemoveIndexes = &ri
		}
		if entity.RemoveUniqueConstraints != nil {
			ruc := *entity.RemoveUniqueConstraints
			for x, u := range ruc {
				s = "ALTER TABLE `" + entity.NewSchemaNameAttr + "` DROP INDEX " + u.SchemaNameAttr + `;`
				s += fmt.Sprintf("DELETE FROM sys_schemainfo WHERE schema_cate=4 and schema_name='%s';", u.SchemaNameAttr)
				sql.WriteString(s)
				sql.WriteString("\n")
				ruc[x].Result = e.runSQL(s)
			}
			actions[idx].RemoveUniqueConstraints = &ruc
		}
		// fmt.Println(sql.String())
		//修改字段
		if entity.ModifyFields != nil {
			mf := *entity.ModifyFields
			for x, f := range mf {
				sql.Reset()
				sql.WriteString("ALTER TABLE `" + entity.NewSchemaNameAttr + "`")
				nullAttr := ""
				//修改字段名
				if f.SchemaNameAttr != f.NewSchemaNameAttr {
					sql.WriteString(" CHANGE `" + f.SchemaNameAttr + "` `" + f.NewSchemaNameAttr + "`")
				} else {
					sql.WriteString(" MODIFY `" + f.SchemaNameAttr + "`")
					if f.IsNullAttr {
						nullAttr += " NULL"
					} else {
						nullAttr += " NOT NULL"
					}
				}
				//修改字段类型
				sql.WriteString(` ` + e.parseFieldType(f.DataTypeAttr, f.TypeOption) + nullAttr)
				//修改默认值约束
				if f.DefaultConstraint == nil {
					sql.WriteString(";ALTER TABLE `" + entity.NewSchemaNameAttr + "` ALTER COLUMN `" + f.NewSchemaNameAttr + "` DROP DEFAULT;")
				} else {
					sql.WriteString(` DEFAULT '` + f.DefaultConstraint.ValueAttr + `';`)
				}
				//修改字段显示名
				if f.DisplayNameAttr != f.NewDisplayNameAttr {
					sql.WriteString(fmt.Sprintf("UPDATE sys_schemainfo SET display_name='%s',update_on=now() WHERE schema_cate=2 and schema_name='%s';", f.NewDisplayNameAttr, entity.NewSchemaNameAttr+"_"+f.NewSchemaNameAttr))
				}
				sql.WriteString("\n")
				mf[x].Result = e.runSQL(sql.String())
				// fmt.Println(sql.String())
			}
			actions[idx].ModifyFields = &mf
		}
		sql.WriteString("\n")
	}
}

func (e *EnvConverter) parseCreateForeign(actions []migration.CreateForeignKeyAction) {
	var sql strings.Builder
	for idx, f := range actions {
		sql.Reset()
		sql.WriteString(fmt.Sprintf("ALTER TABLE `%s` ADD FOREIGN KEY `%s`(`%s`) REFERENCES `%s` (`%s`)", f.ForeignEntityAttr, f.SchemaNameAttr, f.ForeignFieldAttr, f.FromEntityAttr, f.FromFieldAttr))
		//sql.WriteString(`ALTER TABLE ` + f.ForeignEntityAttr + ` ADD FOREIGN KEY ` + f.SchemaNameAttr + `(` + f.ForeignFieldAttr + `) REFERENCES ` + f.FromEntityAttr + ` (` + f.FromFieldAttr + `)`)
		sql.WriteString(` ON UPDATE ` + f.CascadeOptionAttr)
		sql.WriteString(` ON DELETE ` + f.CascadeOptionAttr)
		sql.WriteString(";\n")

		fk := &modeling.ForeignExtra{
			ForeignEntityRelation: f.ForeignEntityRelation,
			FromEntityRelation:    f.FromEntityRelation,
			CascadeOptionAttr:     f.CascadeOptionAttr,
		}
		var extension string
		jstr, _ := json.Marshal(fk)
		extension = string(jstr)
		sql.WriteString(fmt.Sprintf("INSERT INTO sys_schemainfo VALUES (5,'%s',0,'%s','%s',now(),null);\n", f.SchemaNameAttr, f.DisplayNameAttr, extension))

		actions[idx].Result = e.runSQL(sql.String())
	}
}

func (e *EnvConverter) parseRemoveForeign(actions []migration.RemoveForeignKeyAction) {
	var sql strings.Builder
	for idx, f := range actions {
		sql.Reset()
		sql.WriteString("ALTER TABLE `" + f.ForeignEntityAttr + "` DROP FOREIGN KEY `" + f.SchemaNameAttr + "`;")
		sql.WriteString(`DELETE FROM sys_schemainfo WHERE schema_cate=5 and schema_name='` + f.SchemaNameAttr + `';`)
		sql.WriteString("\n")
		// fmt.Println(sql.String())
		actions[idx].Result = e.runSQL(sql.String())
	}
}

func (e *EnvConverter) parseChangeForeign(actions []migration.ChangeForeignKeyAction) {
	var sql strings.Builder
	for idx, f := range actions {
		sql.Reset()

		//外键只能修改显示名称和级联选项，这两个信息存在系统备注表中，直接改备注就行
		fk := &modeling.ForeignExtra{
			ForeignEntityRelation: f.ForeignEntityRelation,
			FromEntityRelation:    f.FromEntityRelation,
			CascadeOptionAttr:     f.CascadeOptionAttr,
		}
		var extension string
		jstr, _ := json.Marshal(fk)
		extension = string(jstr)
		sql.WriteString(fmt.Sprintf("UPDATE sys_schemainfo SET display_name='%s',extension='%s',update_on=now() WHERE schema_cate=5 and schema_name='%s';\n", f.DisplayNameAttr, extension, f.SchemaNameAttr))

		// fmt.Println(sql.String())
		actions[idx].Result = e.runSQL(sql.String())
	}
}

func (e *EnvConverter) parseRenameActions(actions *[]modeling.RenameAction) {
	var sql strings.Builder
	sql.WriteString("INSERT INTO sys_renamelogs values ")
	for i, a := range *actions {
		sql.WriteString(`('` + a.IDAttr + `','` + a.TargetAttr + `','` + a.CreatedOnAttr + `','` + a.BeforeAttr + `','` + a.AfterAttr + `','` + a.TableNameAttr + `')`)
		if i < len(*actions)-1 {
			sql.WriteString(",")
		}
	}
	sql.WriteString("\n")
	e.runSQL(sql.String())
	// fmt.Println(sql.String())
}

func (e *EnvConverter) parseFieldStruct(f migration.Field) string {
	var str strings.Builder
	//字段名
	str.WriteString("`" + f.SchemaNameAttr + "` ")
	//字段类型
	str.WriteString(e.parseFieldType(f.DataTypeAttr, f.TypeOption))
	//是否为空
	if f.IsNullAttr == false {
		str.WriteString("NOT NULL ")
	}
	//默认值
	if f.DefaultConstraint != nil {
		str.WriteString(`DEFAULT '` + f.DefaultConstraint.ValueAttr + `'  `)
	}
	//是否自增
	if f.TypeOption.AutoIncrementAttr == true {
		str.WriteString("AUTO_INCREMENT ")
	}
	//字段备注
	//jstr, _ := json.Marshal(modeling.DescriptionExtra{IsOriginal: f.IsOriginalAttr, DisplayName: f.DisplayNameAttr})
	//str.WriteString(`COMMENT '` + string(jstr) + `'`)
	//str.WriteString(",\n")
	return str.String()
}

func (e *EnvConverter) parseFieldType(typeStr string, option *modeling.TypeOption) string {
	var str strings.Builder
	switch typeStr {
	case "boolean":
		str.WriteString("bit ")
	case "integer":
		if option.LengthAttr <= 16 {
			str.WriteString("smallint ")
		} else if option.LengthAttr <= 32 {
			str.WriteString("int ")
		} else {
			str.WriteString("bigint ")
		}
	case "decimal":
		str.WriteString(`decimal(` + strconv.Itoa(option.LengthAttr) + `,` + strconv.Itoa(option.PrecisionAttr) + `) `)
	case "money":
		str.WriteString("decimal(15,2) ")
	case "datetime":
		str.WriteString("datetime ")
	case "string":
		if option.LengthAttr <= 50 {
			str.WriteString("varchar(50) ")
		} else if option.LengthAttr <= 2000 {
			str.WriteString("varchar(" + strconv.Itoa(option.LengthAttr) + ") ")
		} else {
			str.WriteString("text ")
		}
	}
	return str.String()
}

func (e *EnvConverter) parseIndexFields(i migration.Index) string {
	var str strings.Builder
	str.WriteString("(")
	for o, f := range i.Columns {
		str.WriteString("`" + f.ColumnAttr + "` ")
		if i.IsPrimaryAttr == false {
			if f.DirectionASCAttr == true {
				str.WriteString("ASC")
			} else {
				str.WriteString("DESC")
			}
		}
		if o < len(i.Columns)-1 {
			str.WriteString(",")
		}
	}
	str.WriteString(")")
	return str.String()
}

func (e *EnvConverter) parseUniqueCostraintStruct(c migration.UniqueConstraint) string {
	var str strings.Builder
	str.WriteString(`CONSTRAINT ` + c.SchemaNameAttr + ` UNIQUE (`)
	for i, f := range c.Columns {
		str.WriteString("`" + f.ColumnAttr + "`")
		if i < len(c.Columns)-1 {
			str.WriteString(",")
		}
	}
	str.WriteString(`)`)
	//str.WriteString(`) COMMENT '` + c.DisplayNameAttr + `'`)
	return str.String()
}

func (e *EnvConverter) runSQL(sql string) *migration.ActionResult {
	fmt.Println(sql)

	result := migration.ActionResult{}
	start := time.Now()

	tx, _ := e.Db.Begin()
	_, err := tx.Exec(sql)

	if err == nil {
		tx.Commit()
		result.SuccessAttr = true
		e.SuccessedCnt++
	} else {
		tx.Rollback()
		result.SuccessAttr = false
		result.Description = err.Error()
		e.FailedCnt++
	}
	result.DurationAttr = uint16(time.Since(start).Milliseconds())
	return &result
}
