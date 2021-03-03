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

package postgres

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
	for idx, entity := range actions {
		sql.Reset()
		sql.WriteString(`----------------------创建表：` + entity.SchemaNameAttr + `-----------------------`)
		sql.WriteString("\n")
		sql.WriteString(`CREATE TABLE public.` + entity.SchemaNameAttr)
		sql.WriteString("\n(\n")
		//字段清单
		for i := 0; i < len(entity.Fields); i++ {
			if i == 0 {
				sql.WriteString(`	`)
			} else {
				sql.WriteString(`	,`)
			}
			sql.WriteString(e.parseFieldStruct(entity.Fields[i]))
			sql.WriteString("\n")
		}
		//主键和索引
		var ind strings.Builder
		if entity.Indexes != nil {
			for _, i := range *entity.Indexes {
				if i.IsPrimaryAttr == true {
					sql.WriteString(`	,CONSTRAINT ` + i.SchemaNameAttr + ` PRIMARY KEY ` + e.parseIndexFields(i))
					sql.WriteString("\n")
				} else {
					ind.WriteString(`CREATE INDEX ` + i.SchemaNameAttr + ` ON public.` + entity.SchemaNameAttr + ` USING btree ` + e.parseIndexFields(i) + `;COMMENT ON INDEX public.` + i.SchemaNameAttr + `
					IS '` + i.DisplayNameAttr + `';`)
					ind.WriteString("\n")
				}
			}
		}
		//唯一约束
		if entity.UniqueConstraints != nil {
			for _, c := range *entity.UniqueConstraints {
				sql.WriteString(`	,` + e.parseUniqueCostraintStruct(c))
				sql.WriteString("\n")
			}
		}
		sql.WriteString(");\n")
		//表备注
		jstr, _ := json.Marshal(modeling.DescriptionExtra{IsOriginal: entity.IsOriginalAttr, DisplayName: entity.DisplayNameAttr})
		sql.WriteString(`COMMENT ON TABLE public.` + entity.SchemaNameAttr + ` IS '` + string(jstr) + `';`)
		sql.WriteString("\n")
		//字段备注
		for _, f := range entity.Fields {
			jstr, _ := json.Marshal(modeling.DescriptionExtra{IsOriginal: f.IsOriginalAttr, DisplayName: f.DisplayNameAttr})
			sql.WriteString(`COMMENT ON COLUMN public.` + entity.SchemaNameAttr + `.` + f.SchemaNameAttr + ` IS '` + string(jstr) + `';`)
			sql.WriteString("\n")
		}
		//索引
		sql.WriteString(ind.String())
		if entity.UniqueConstraints != nil {
			for _, c := range *entity.UniqueConstraints {
				sql.WriteString(`COMMENT ON CONSTRAINT ` + c.SchemaNameAttr + ` ON public.` + entity.SchemaNameAttr + `
				IS '` + c.DisplayNameAttr + `';`)
				sql.WriteString("\n")
			}
		}
		fmt.Println(sql.String())

		//执行脚本
		actions[idx].Result = e.runSQL(sql.String())
	}
}

func (e *EnvConverter) parseRemoveEntity(actions []migration.RemoveItem) {
	var sql strings.Builder
	for idx, entity := range actions {
		sql.Reset()
		sql.WriteString(`----------------------删除表：` + entity.SchemaNameAttr + `-----------------------`)
		sql.WriteString("\n")
		sql.WriteString(`DROP TABLE ` + entity.SchemaNameAttr + `;`)
		sql.WriteString("\n")
		fmt.Println(sql.String())
		//执行脚本
		actions[idx].Result = e.runSQL(sql.String())
	}
}

func (e *EnvConverter) parseChangeEntity(actions []migration.ChangeEntityAction) {
	var sql strings.Builder
	for idx, entity := range actions {
		sql.Reset()
		sql.WriteString(`----------------------修改表：` + entity.SchemaNameAttr + `-----------------------`)
		sql.WriteString("\n")
		//修改表名
		if entity.SchemaNameAttr != entity.NewSchemaNameAttr {
			sql.WriteString(`ALTER TABLE public.` + entity.SchemaNameAttr + ` RENAME TO ` + entity.NewSchemaNameAttr + `;`)
			sql.WriteString("\n")
		}
		//修改表显示名
		if entity.DisplayNameAttr != entity.NewDisplayNameAttr {
			jstr, _ := json.Marshal(modeling.DescriptionExtra{IsOriginal: entity.IsOriginalAttr, DisplayName: entity.NewDisplayNameAttr})
			sql.WriteString(`COMMENT ON TABLE public.` + entity.NewSchemaNameAttr + ` IS '` + string(jstr) + `';`)
			sql.WriteString("\n")
		}
		fmt.Println(sql.String())
		actions[idx].Result = e.runSQL(sql.String())

		sql.Reset()
		var s string
		//新增字段
		if entity.NewFields != nil {
			nf := *entity.NewFields
			for i, f := range nf {
				s = `ALTER TABLE public.` + entity.NewSchemaNameAttr + ` ADD COLUMN ` + e.parseFieldStruct(f) + `;`
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
				s = `ALTER TABLE public.` + entity.NewSchemaNameAttr + ` ADD ` + e.parseUniqueCostraintStruct(c) + `;COMMENT ON CONSTRAINT ` + c.SchemaNameAttr + ` ON public.` + entity.NewSchemaNameAttr + `
				IS '` + c.DisplayNameAttr + `';`
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
					s = `ALTER TABLE public.` + entity.NewSchemaNameAttr + ` ADD CONSTRAINT ` + i.SchemaNameAttr + ` PRIMARY KEY ` + e.parseIndexFields(i) + `;`
				} else {
					s = `CREATE INDEX ` + i.SchemaNameAttr + ` ON public.` + entity.NewSchemaNameAttr + ` USING btree ` + e.parseIndexFields(i) + `;COMMENT ON INDEX public.` + i.SchemaNameAttr + `
					IS '` + i.DisplayNameAttr + `';`
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
				s = `ALTER TABLE public.` + entity.NewSchemaNameAttr + ` DROP COLUMN ` + f.SchemaNameAttr + `;`
				sql.WriteString(s)
				sql.WriteString("\n")
				rf[x].Result = e.runSQL(s)
			}
			actions[idx].RemoveFields = &rf
		}
		if entity.RemoveIndexes != nil {
			ri := *entity.RemoveIndexes
			for x, i := range ri {
				s = `DROP INDEX ` + i.SchemaNameAttr + `;`
				sql.WriteString(s)
				sql.WriteString("\n")
				ri[x].Result = e.runSQL(s)
			}
			actions[idx].RemoveIndexes = &ri
		}
		if entity.RemoveUniqueConstraints != nil {
			ruc := *entity.RemoveUniqueConstraints
			for x, u := range ruc {
				s = `ALTER TABLE public.` + entity.NewSchemaNameAttr + ` DROP CONSTRAINT ` + u.SchemaNameAttr + `;`
				sql.WriteString(s)
				sql.WriteString("\n")
				ruc[x].Result = e.runSQL(s)
			}
			actions[idx].RemoveUniqueConstraints = &ruc
		}
		fmt.Println(sql.String())
		//修改字段
		if entity.ModifyFields != nil {
			mf := *entity.ModifyFields
			for x, f := range mf {
				sql.Reset()
				//修改字段名
				if f.SchemaNameAttr != f.NewSchemaNameAttr {
					sql.WriteString(`ALTER TABLE public.` + entity.NewSchemaNameAttr + ` RENAME ` + f.SchemaNameAttr + ` TO ` + f.NewSchemaNameAttr + `;`)
					sql.WriteString("\n")
				}
				//修改字段显示名
				if f.DisplayNameAttr != f.NewDisplayNameAttr {
					jstr, _ := json.Marshal(modeling.DescriptionExtra{IsOriginal: f.IsOriginalAttr, DisplayName: f.NewDisplayNameAttr})
					sql.WriteString(`COMMENT ON COLUMN public.` + entity.NewSchemaNameAttr + `.` + f.NewSchemaNameAttr + ` IS '` + string(jstr) + `';`)
					sql.WriteString("\n")
				}
				//修改字段类型
				sql.WriteString(`ALTER TABLE public.` + entity.NewSchemaNameAttr + ` ALTER COLUMN ` + f.NewSchemaNameAttr + ` TYPE ` + e.parseFieldType(f.DataTypeAttr, f.TypeOption) + `;`)
				sql.WriteString("\n")
				//修改默认值约束
				if f.DefaultConstraint == nil {
					sql.WriteString(`ALTER TABLE public.` + entity.NewSchemaNameAttr + ` ALTER COLUMN ` + f.NewSchemaNameAttr + ` DROP DEFAULT;`)
				} else {
					sql.WriteString(`ALTER TABLE public.` + entity.NewSchemaNameAttr + ` ALTER COLUMN ` + f.NewSchemaNameAttr + ` SET DEFAULT '` + f.DefaultConstraint.ValueAttr + `';`)
				}
				sql.WriteString("\n")
				mf[x].Result = e.runSQL(sql.String())
				fmt.Println(sql.String())
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
		sql.WriteString(`ALTER TABLE public.` + f.ForeignEntityAttr + ` ADD CONSTRAINT ` + f.SchemaNameAttr + ` FOREIGN KEY (` + f.ForeignFieldAttr + `) REFERENCES public.` + f.FromEntityAttr + ` (` + f.FromFieldAttr + `)`)
		sql.WriteString(` ON UPDATE ` + f.CascadeOptionAttr)
		sql.WriteString(` ON DELETE ` + f.CascadeOptionAttr)
		sql.WriteString(";\n")
		jstr, _ := json.Marshal(modeling.ForeignExtra{ForeignEntityRelation: f.ForeignEntityRelation, FromEntityRelation: f.FromEntityRelation})
		sql.WriteString(`COMMENT ON CONSTRAINT ` + f.SchemaNameAttr + ` ON public.` + f.ForeignEntityAttr + ` IS  '` + string(jstr) + `';`)
		sql.WriteString("\n")
		fmt.Println(sql.String())
		actions[idx].Result = e.runSQL(sql.String())
	}
}

func (e *EnvConverter) parseRemoveForeign(actions []migration.RemoveForeignKeyAction) {
	var sql strings.Builder
	for idx, f := range actions {
		sql.Reset()
		sql.WriteString(`ALTER TABLE public.` + f.ForeignEntityAttr + ` DROP CONSTRAINT ` + f.SchemaNameAttr + `;`)
		sql.WriteString("\n")
		fmt.Println(sql.String())
		actions[idx].Result = e.runSQL(sql.String())
	}
}

func (e *EnvConverter) parseChangeForeign(actions []migration.ChangeForeignKeyAction) {
	var sql strings.Builder
	for idx, f := range actions {
		sql.Reset()
		fq := `select c.conkey,c.confkey,c.conrelid,c1.relname,c.confrelid,c2.relname from pg_constraint c
			left join pg_class c1 on c.conrelid=c1.oid
			left join pg_class c2 on c.confrelid=c2.oid
			where c.contype='f' and c.conname='` + f.SchemaNameAttr + `'`
		var conkey, confkey, conrelid, relname, confrelid, relfname string
		row := e.Db.QueryRow(fq)
		if row != nil {
			row.Scan(&conkey, &confkey, &conrelid, &relname, &confrelid, &relfname)
			//修改外键方法是删除重建
			sql.WriteString(`ALTER TABLE public.` + relname + ` DROP CONSTRAINT ` + f.SchemaNameAttr + `;`)
			sql.WriteString("\n")
			var fd, td string
			e.Db.QueryRow(`select attname from  pg_attribute where attnum=` + strings.TrimSuffix(strings.TrimPrefix(conkey, "{"), "}") + ` and attrelid=` + conrelid).Scan(&fd)
			e.Db.QueryRow(`select attname from  pg_attribute where attnum=` + strings.TrimSuffix(strings.TrimPrefix(confkey, "{"), "}") + ` and attrelid=` + confrelid).Scan(&td)
			sql.WriteString(`ALTER TABLE public.` + relname + ` ADD CONSTRAINT ` + f.SchemaNameAttr + ` FOREIGN KEY (` + fd + `) REFERENCES public.` + relfname + ` (` + td + `)`)
			sql.WriteString(` ON UPDATE ` + f.CascadeOptionAttr)
			sql.WriteString(` ON DELETE ` + f.CascadeOptionAttr)
			sql.WriteString(";\n")
			jstr, _ := json.Marshal(modeling.ForeignExtra{ForeignEntityRelation: f.ForeignEntityRelation, FromEntityRelation: f.FromEntityRelation})
			sql.WriteString(`COMMENT ON CONSTRAINT ` + f.SchemaNameAttr + ` ON public.` + relname + ` IS  '` + string(jstr) + `';`)
			sql.WriteString("\n")
		}
		fmt.Println(sql.String())
		actions[idx].Result = e.runSQL(sql.String())
	}
}

func (e *EnvConverter) parseRenameActions(actions *[]modeling.RenameAction) {
	var sql strings.Builder
	for _, a := range *actions {
		sql.WriteString(`INSERT INTO internal.renamelogs values('` + a.IDAttr + `','` + a.TargetAttr + `','` + a.CreatedOnAttr + `','` + a.BeforeAttr + `','` + a.AfterAttr + `','` + a.TableNameAttr + `');`)
		sql.WriteString("\n")
	}
	e.runSQL(sql.String())
	fmt.Println(sql.String())
}

func (e *EnvConverter) parseFieldStruct(f migration.Field) string {
	var str strings.Builder
	//字段名
	str.WriteString(f.SchemaNameAttr + ` `)
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
		str.WriteString("GENERATED ALWAYS AS IDENTITY ")
	}
	//str.WriteString(",\n")
	return str.String()
}

func (e *EnvConverter) parseFieldType(typeStr string, option *modeling.TypeOption) string {
	var str strings.Builder
	switch typeStr {
	case "boolean":
		str.WriteString("bool ")
	case "integer":
		if option.LengthAttr <= 16 {
			str.WriteString("smallint ")
		} else if option.LengthAttr <= 32 {
			str.WriteString("integer ")
		} else {
			str.WriteString("bigint ")
		}
	case "decimal":
		str.WriteString(`numeric(` + strconv.Itoa(option.LengthAttr) + `,` + strconv.Itoa(option.PrecisionAttr) + `) `)
	case "money":
		str.WriteString("money ")
	case "datetime":
		str.WriteString("timestamp without time zone ")
	case "string":
		if option.LengthAttr <= 50 {
			str.WriteString("character varying(50) ")
		} else if option.LengthAttr <= 2000 {
			str.WriteString("character varying(2000) ")
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
		str.WriteString(f.ColumnAttr + ` `)
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
		str.WriteString(f.ColumnAttr)
		if i < len(c.Columns)-1 {
			str.WriteString(",")
		}
	}
	str.WriteString(")")
	return str.String()
}

func (e *EnvConverter) runSQL(sql string) *migration.ActionResult {
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
