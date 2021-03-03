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
	"dataapi/internal/kernel/metadata/migration"
	"dataapi/internal/kernel/metadata/modeling"
	"errors"
)

//target 和 source对比方法：
//   为source生成一个数组tag，数组长度为source中字段数量，值全部为0 ，0表示该字段未在target中找到对应字段
//   在target循环中遍历source的字段，如果找到相对于的字段，则将tag值标为1， 1表示该字段已经在target中找到对应字段
//   如果没找到对应字段，则表示该字段在source中被删除了
//   target中所有字段都循环完之后，如果source中还存在tag为0 则表示 该字段为source中新建的字段

// ModelingToMigration 将TargetModeling 和 SourceModeling 生成 Migration
func ModelingToMigration(modelingTarget *modeling.Model, modelingSource *modeling.Model) (migration.Migration, error) {
	migrationResult := migration.Migration{
		RemoveForeignKey: &[]migration.RemoveForeignKeyAction{},
		CreateForeignKey: &[]migration.CreateForeignKeyAction{},
		ChangeForeignKey: &[]migration.ChangeForeignKeyAction{},
		RemoveEntity:     &[]migration.RemoveItem{},
		CreateEntity:     &[]migration.CreateEntityAction{},
		ChangeEntity:     &[]migration.ChangeEntityAction{},
	}

	// 对比重命名操作的元数据，计算增量
	renameActionMetadata, err := GetRenameActionMetadata(modelingTarget.RenameActions, modelingSource.RenameActions)

	if err != nil {
		return migrationResult, err
	}

	migrationResult.RenameActions = &renameActionMetadata

	// 简化重命名操作
	var renameActions []modeling.RenameAction
	var fieldRenameActions []modeling.RenameAction
	var tableRenameActions []modeling.RenameAction
	if len(renameActionMetadata) > 0 {
		renameActions, err = GetSimplifiedRenameAction(renameActionMetadata)

		// 将命名操作分为字段类型和表类型
		for _, value := range renameActions {
			if value.TargetAttr == "field" {
				fieldRenameActions = append(fieldRenameActions, value)
			} else {
				tableRenameActions = append(tableRenameActions, value)
			}
		}
	}

	// 对比两个Modeling中的实体
	ContrastModelingEntities(modelingTarget.Entities, modelingSource.Entities, tableRenameActions, fieldRenameActions, &migrationResult)

	// 对比两个Modeling的外键
	ContrastModelingForeignKey(modelingTarget.ForeignKeys, modelingSource.ForeignKeys, &migrationResult)

	return migrationResult, nil
}

/*
算法原理:
	原数组 和 简化数组
	第一个时间节点的直接加入简化数组
	后面的每个时间节点内对简化数组进行检索
	如果不存在任意的X,Y 使简化数组[x].After == 原数组[Y].Before && 简化数组[x].Target == 原数组[Y].Target 成立 就直接将其加入简化数组
	如果存在 则简化数组[x].After = 原数组[Y].After, 简化数组[x].CreateOn = 原数组[Y].CreateOn
*/

// GetSimplifiedRenameAction 获取简化重命名操作的数组
func GetSimplifiedRenameAction(renameActionMetadata []modeling.RenameAction) ([]modeling.RenameAction, error) {
	renameActions := &[]modeling.RenameAction{}

	// 获取当前Rename操作的时间节点
	time := renameActionMetadata[0].CreatedOnAttr

	for i := 0; i < len(renameActionMetadata); i++ {
		thisTime := renameActionMetadata[i].CreatedOnAttr
		if time == thisTime {
			FindParentNode(renameActions, renameActionMetadata[i])

		} else {
			time = thisTime
			i--
		}
	}

	return *renameActions, nil
}

// FindParentNode 寻找其父节点是否存在
func FindParentNode(renameActions *[]modeling.RenameAction, renameAction modeling.RenameAction) {
	isExistParent := false
	for index, value := range *renameActions {
		if renameAction.TargetAttr == value.TargetAttr && renameAction.BeforeAttr == value.AfterAttr && renameAction.CreatedOnAttr != value.CreatedOnAttr {
			if value.TargetAttr == "field" && value.TableNameAttr != renameAction.TableNameAttr {
				continue
			} else if value.TargetAttr == "field" && value.TableNameAttr == renameAction.TableNameAttr {
				(*renameActions)[index].TableNameAttr = renameAction.TableNameAttr
			}

			(*renameActions)[index].AfterAttr = renameAction.AfterAttr
			(*renameActions)[index].CreatedOnAttr = renameAction.CreatedOnAttr
			isExistParent = true
			break
		}
	}

	if !isExistParent {
		*renameActions = append(*renameActions, renameAction)
	}
}

// GetRenameActionMetadata 获取改名操作元数据增量数组
func GetRenameActionMetadata(targetRenameAction *[]modeling.RenameAction, sourceRenameAction *[]modeling.RenameAction) ([]modeling.RenameAction, error) {
	var result []modeling.RenameAction

	if targetRenameAction == nil && sourceRenameAction == nil {
		return result, nil
	} else if targetRenameAction == nil {
		return *sourceRenameAction, nil
	}

	if len(*sourceRenameAction) < len(*targetRenameAction) || sourceRenameAction == nil {
		return nil, errors.New("Source版本落后于Target版本, 若要执行请进行回滚操作")
	}

	for i := len(*targetRenameAction); i < len(*sourceRenameAction); i++ {
		result = append(result, (*sourceRenameAction)[i])
	}

	return result, nil
}

// ContrastModelingEntities 对比两个Modeling中的实体
func ContrastModelingEntities(targetEntities []modeling.Entity, sourceEntities []modeling.Entity, tableRenameActions []modeling.RenameAction, fieldRenameActions []modeling.RenameAction, migration *migration.Migration) {
	// 实体标记 0为未找到对应实体 1表示已匹配对应实体
	tagEntities := make([]int, 0)
	for i := 0; i < len(sourceEntities); i++ {
		tagEntities = append(tagEntities, 0)
	}

	for _, modelingTargetEntityValue := range targetEntities {
		findEntity := false

		// 寻找相同的实体
		findEntity = FindSameEntity(modelingTargetEntityValue, sourceEntities, tableRenameActions, fieldRenameActions, &tagEntities, migration.ChangeEntity)

		// NewModel中找不到对应的实体时，则表明该实体被删除
		if !findEntity {
			*migration.RemoveEntity = append(*migration.RemoveEntity, GetRemoveEntityAction(modelingTargetEntityValue))
		}
	}

	// 当所有modelingTarget都遍历完成了 modelingSource中仍然有未标记的Entity则表示该Entity是新创建的
	for index, value := range tagEntities {
		if value == 0 {
			*migration.CreateEntity = append(*migration.CreateEntity, GetCreateEntityAction(sourceEntities[index]))
		}
	}
}

// FindSameEntity 寻找相对对应的实体
func FindSameEntity(TargetEntity modeling.Entity, sourceEntities []modeling.Entity, tableRenameActions []modeling.RenameAction, fieldRenameActions []modeling.RenameAction, tagEntities *[]int, changeEntityActions *[]migration.ChangeEntityAction) bool {
	for modelingSourceEntityIndex, modelingSourceEntityValue := range sourceEntities {
		// 当新的modeling中实体被标记时 直接进入下一个
		if (*tagEntities)[modelingSourceEntityIndex] == 1 {
			continue
		}

		// 检查实体是否是相对应的
		findEntity := IsSameEntity(changeEntityActions, TargetEntity, modelingSourceEntityValue, tableRenameActions, fieldRenameActions)

		// 找到时进行标记
		if findEntity {
			(*tagEntities)[modelingSourceEntityIndex] = 1
			return true
		}
	}
	return false
}

// IsSameEntity 判断是否为相对应的实体
func IsSameEntity(changeEntityActions *[]migration.ChangeEntityAction, targetEntity modeling.Entity, sourceEntity modeling.Entity, tableRenameActions []modeling.RenameAction, fieldRenameActions []modeling.RenameAction) bool {
	//判断两者是否是改名的实体
	for _, value := range tableRenameActions {
		if targetEntity.SchemaNameAttr == value.BeforeAttr && sourceEntity.SchemaNameAttr == value.AfterAttr {
			GetChangeEntityAction(targetEntity, sourceEntity, fieldRenameActions, changeEntityActions)
			return true
		}
	}

	// 判断两者名字是否相同, 相同则为对应实体
	if targetEntity.SchemaNameAttr == sourceEntity.SchemaNameAttr {
		GetChangeEntityAction(targetEntity, sourceEntity, fieldRenameActions, changeEntityActions)
		return true
	}

	return false
}

// ContrastModelingForeignKey 对比两个Modeling中的外键
func ContrastModelingForeignKey(targetForeignKeys *[]modeling.ForeignKey, sourceForeignKeys *[]modeling.ForeignKey, migration *migration.Migration) {
	// 外键标记 0为未找到对应外键 1表示已匹配对应外键
	tagForeignKeys := make([]int, 0)

	if sourceForeignKeys != nil {
		for i := 0; i < len(*sourceForeignKeys); i++ {
			tagForeignKeys = append(tagForeignKeys, 0)
		}
	}

	if targetForeignKeys != nil {
		for _, modelingTargetForeignKeyValue := range *targetForeignKeys {
			findForeignKey := false

			findForeignKey = FindSameForeignKey(modelingTargetForeignKeyValue, sourceForeignKeys, &tagForeignKeys, migration.ChangeForeignKey)

			if !findForeignKey {
				*migration.RemoveForeignKey = append(*migration.RemoveForeignKey, GetRemoveForeignKeyAction(modelingTargetForeignKeyValue))
			}
		}
	}

	// 当所有modelingTarget都遍历完成了 modelingSource中仍然有未标记的ForeignKey则表示该ForeignKey是新创建的
	for index, value := range tagForeignKeys {
		if value == 0 && sourceForeignKeys != nil {
			*migration.CreateForeignKey = append(*migration.CreateForeignKey, GetCreateForeignKeyAction((*sourceForeignKeys)[index]))
		}
	}
}

// FindSameForeignKey 想找相同的外键
func FindSameForeignKey(TargetForeignKey modeling.ForeignKey, sourceForeignKeys *[]modeling.ForeignKey, tagForeignKeys *[]int, changeForeignKeys *[]migration.ChangeForeignKeyAction) bool {
	for modelingSourceForeignKeyIndex, modelingSourceForeignKeyValue := range *sourceForeignKeys {
		// 如果被标记则表明已经有相同的字段对应 直接跳过
		if (*tagForeignKeys)[modelingSourceForeignKeyIndex] == 1 {
			continue
		}

		// 名字是否相同
		if TargetForeignKey.SchemaNameAttr == modelingSourceForeignKeyValue.SchemaNameAttr {
			GetChangeForeignKeyAction(TargetForeignKey, modelingSourceForeignKeyValue, changeForeignKeys)
			(*tagForeignKeys)[modelingSourceForeignKeyIndex] = 1
			return true
		}
	}
	return false
}

// GetCreateEntityAction 获取创建实体操作
func GetCreateEntityAction(modelingSourceEntity modeling.Entity) migration.CreateEntityAction {
	createEntityAction := migration.CreateEntityAction{
		DisplayNameAttr:   modelingSourceEntity.DisplayNameAttr,
		IsOriginalAttr:    modelingSourceEntity.IsOriginalAttr,
		SchemaNameAttr:    modelingSourceEntity.SchemaNameAttr,
		ClusteredAttr:     modelingSourceEntity.ClusteredAttr,
		Fields:            []migration.Field{},
		UniqueConstraints: &[]migration.UniqueConstraint{},
		Indexes:           &[]migration.Index{},
	}

	// 若字段存在则创建字段
	if modelingSourceEntity.Fields != nil {
		for _, value := range modelingSourceEntity.Fields {
			createEntityAction.Fields = append(createEntityAction.Fields, migration.Field{
				SchemaNameAttr:    value.SchemaNameAttr,
				IsNullAttr:        value.IsNullAttr,
				DisplayNameAttr:   value.DisplayNameAttr,
				IsOriginalAttr:    value.IsOriginalAttr,
				DataTypeAttr:      value.DataTypeAttr,
				TypeOption:        value.TypeOption,
				DefaultConstraint: value.DefaultConstraint,
			})
		}
	}

	// 若索引存在则创建索引
	if modelingSourceEntity.Indexes != nil {
		for _, value := range *modelingSourceEntity.Indexes {
			*createEntityAction.Indexes = append(*createEntityAction.Indexes, migration.Index{
				SchemaNameAttr:  value.SchemaNameAttr,
				DisplayNameAttr: value.DisplayNameAttr,
				IsPrimaryAttr:   value.IsPrimaryAttr,
				Columns:         value.Columns,
			})
		}
	}

	// 若存在唯一约束则创建唯一约束
	if modelingSourceEntity.UniqueConstraints != nil {
		for _, value := range *modelingSourceEntity.UniqueConstraints {
			*createEntityAction.UniqueConstraints = append(*createEntityAction.UniqueConstraints, migration.UniqueConstraint{
				SchemaNameAttr:  value.SchemaNameAttr,
				DisplayNameAttr: value.DisplayNameAttr,
				Columns:         value.Columns,
			})
		}
	}

	return createEntityAction
}

// GetRemoveEntityAction 获取删除实体操作
func GetRemoveEntityAction(modelingTargetEntity modeling.Entity) migration.RemoveItem {
	return migration.RemoveItem{
		SchemaNameAttr: modelingTargetEntity.SchemaNameAttr,
	}
}

// GetChangeEntityAction 获取修改实体操作
func GetChangeEntityAction(modelingTargetEntity modeling.Entity, modelingSourceEntity modeling.Entity, fieldRenameAction []modeling.RenameAction, changeEntityActions *[]migration.ChangeEntityAction) {
	var changeEntityAction = migration.ChangeEntityAction{
		SchemaNameAttr:          modelingTargetEntity.SchemaNameAttr,
		DisplayNameAttr:         modelingTargetEntity.DisplayNameAttr,
		IsOriginalAttr:          modelingTargetEntity.IsOriginalAttr,
		NewDisplayNameAttr:      modelingSourceEntity.DisplayNameAttr,
		NewSchemaNameAttr:       modelingSourceEntity.SchemaNameAttr,
		NewFields:               &[]migration.Field{},
		NewIndexes:              &[]migration.Index{},
		NewUniqueConstraints:    &[]migration.UniqueConstraint{},
		RemoveFields:            &[]migration.RemoveItem{},
		RemoveIndexes:           &[]migration.RemoveItem{},
		RemoveUniqueConstraints: &[]migration.RemoveItem{},
		ModifyFields:            &[]migration.ModifyField{},
	}

	// 实体是否改变
	var isChange = false

	changeEntityAction.IgnoreExistsDataAttr = true
	if modelingTargetEntity.DisplayNameAttr != modelingSourceEntity.DisplayNameAttr {
		changeEntityAction.NewDisplayNameAttr = modelingSourceEntity.DisplayNameAttr
		isChange = true
	}

	if modelingTargetEntity.SchemaNameAttr != modelingSourceEntity.SchemaNameAttr {
		changeEntityAction.NewSchemaNameAttr = modelingSourceEntity.SchemaNameAttr
		isChange = true
	}

	if modelingTargetEntity.IsOriginalAttr != modelingSourceEntity.IsOriginalAttr {
		changeEntityAction.IsOriginalAttr = modelingSourceEntity.IsOriginalAttr
		isChange = true
	}

	// 检查字段变化
	GetChangeFieldAction(modelingTargetEntity.Fields, modelingSourceEntity.Fields, fieldRenameAction, modelingSourceEntity.SchemaNameAttr, &changeEntityAction, &isChange)

	// 检查索引变化
	GetChangeIndexAction(modelingTargetEntity.Indexes, modelingSourceEntity.Indexes, &changeEntityAction, &isChange)

	// 检查唯一约束变化
	GetChangeUniqueConstraintAction(modelingTargetEntity.UniqueConstraints, modelingSourceEntity.UniqueConstraints, &changeEntityAction, &isChange)

	// 如果发生变化则需要将其加入ChangeEntityAction
	if isChange {
		*changeEntityActions = append(*changeEntityActions, changeEntityAction)
	}
}

// GetCreateForeignKeyAction 获取创建外键操作
func GetCreateForeignKeyAction(modelingSourceForeignKey modeling.ForeignKey) migration.CreateForeignKeyAction {
	return migration.CreateForeignKeyAction{
		SchemaNameAttr:        modelingSourceForeignKey.SchemaNameAttr,
		DisplayNameAttr:       modelingSourceForeignKey.DisplayNameAttr,
		ForeignEntityAttr:     modelingSourceForeignKey.ForeignEntityAttr,
		ForeignFieldAttr:      modelingSourceForeignKey.ForeignFieldAttr,
		FromEntityAttr:        modelingSourceForeignKey.FromEntityAttr,
		FromFieldAttr:         modelingSourceForeignKey.FromFieldAttr,
		CascadeOptionAttr:     modelingSourceForeignKey.CascadeOptionAttr,
		ForeignEntityRelation: modelingSourceForeignKey.ForeignEntityRelation,
		FromEntityRelation:    modelingSourceForeignKey.FromEntityRelation,
	}
}

// GetRemoveForeignKeyAction 获取删除外键操作
func GetRemoveForeignKeyAction(modelingTargetForeignKey modeling.ForeignKey) migration.RemoveForeignKeyAction {
	return migration.RemoveForeignKeyAction{
		ForeignEntityAttr: modelingTargetForeignKey.ForeignEntityAttr,
		SchemaNameAttr:    modelingTargetForeignKey.SchemaNameAttr,
	}
}

// GetChangeForeignKeyAction 获取修改外键操作
func GetChangeForeignKeyAction(modelingTargetForeignKey modeling.ForeignKey, modelingSourceForeignKey modeling.ForeignKey, changeForeignKeys *[]migration.ChangeForeignKeyAction) {
	if modelingTargetForeignKey.CascadeOptionAttr != modelingSourceForeignKey.CascadeOptionAttr || modelingTargetForeignKey.DisplayNameAttr != modelingSourceForeignKey.DisplayNameAttr ||
		modelingTargetForeignKey.FromEntityRelation != modelingSourceForeignKey.FromEntityRelation || modelingTargetForeignKey.ForeignEntityRelation != modelingSourceForeignKey.ForeignEntityRelation {
		*changeForeignKeys = append(*changeForeignKeys, migration.ChangeForeignKeyAction{
			SchemaNameAttr:        modelingSourceForeignKey.SchemaNameAttr,
			DisplayNameAttr:       modelingSourceForeignKey.DisplayNameAttr,
			CascadeOptionAttr:     modelingSourceForeignKey.CascadeOptionAttr,
			ForeignEntityRelation: modelingSourceForeignKey.ForeignEntityRelation,
			FromEntityRelation:    modelingSourceForeignKey.FromEntityRelation,
		})
	}
}

// GetChangeFieldAction 获取字段的修改操作，包括创建  删除  修改
func GetChangeFieldAction(targetFields []modeling.Field, sourceFields []modeling.Field, renameAction []modeling.RenameAction, tableName string, changeEntityAction *migration.ChangeEntityAction, isChange *bool) {
	tagFields := make([]int, 0)

	if targetFields == nil {
		targetFields = []modeling.Field{}
	}

	if sourceFields == nil {
		sourceFields = []modeling.Field{}
	}

	for i := 0; i < len(sourceFields); i++ {
		tagFields = append(tagFields, 0)
	}

	for _, targetFieldValue := range targetFields {
		// 查找名字相同的字段
		isFind := FindSameField(targetFieldValue, sourceFields, &tagFields, renameAction, changeEntityAction.ModifyFields, isChange, tableName)
		if !isFind {
			*changeEntityAction.RemoveFields = append(*changeEntityAction.RemoveFields, migration.RemoveItem{
				SchemaNameAttr: targetFieldValue.SchemaNameAttr,
			})
			*isChange = true
		}
	}

	// 若未标记 则表明说新建字段
	for index, value := range tagFields {
		if value == 0 {
			*changeEntityAction.NewFields = append(*changeEntityAction.NewFields, migration.Field{
				SchemaNameAttr:    sourceFields[index].SchemaNameAttr,
				IsNullAttr:        sourceFields[index].IsNullAttr,
				DisplayNameAttr:   sourceFields[index].DisplayNameAttr,
				IsOriginalAttr:    sourceFields[index].IsOriginalAttr,
				DataTypeAttr:      sourceFields[index].DataTypeAttr,
				TypeOption:        sourceFields[index].TypeOption,
				DefaultConstraint: sourceFields[index].DefaultConstraint,
			})

			*isChange = true
		}
	}
}

// FindSameField 寻找相同的字段
func FindSameField(targetField modeling.Field, sourceFields []modeling.Field, tagFields *[]int, renameAction []modeling.RenameAction, modifyFieldActions *[]migration.ModifyField, isChange *bool, tableName string) bool {
	isFind := false
	for sourceFieldIndex, sourceFieldValue := range sourceFields {
		if (*tagFields)[sourceFieldIndex] == 1 {
			continue
		}

		// 关于表中字段改名， 由于存在顺序问题，表的名字可能经过多次修改存在重名问题，其中字段的修改也会存在问题
		// 暂时进行粗糙对比，后期有待优化
		for _, value := range renameAction {
			if targetField.SchemaNameAttr == value.BeforeAttr && sourceFieldValue.SchemaNameAttr == value.AfterAttr && value.TableNameAttr == tableName {
				isFind = true
				SameFieldIsChange(targetField, sourceFieldValue, modifyFieldActions, isChange)
				break
			}
		}

		if targetField.SchemaNameAttr == sourceFieldValue.SchemaNameAttr && !isFind {
			isFind = true
			SameFieldIsChange(targetField, sourceFieldValue, modifyFieldActions, isChange)
		}

		if isFind {
			(*tagFields)[sourceFieldIndex] = 1
			break
		}
	}

	return isFind
}

// FindSameField 相同字段是否改变 改变则加入Modify中，未改变则不做事
func SameFieldIsChange(targetField modeling.Field, sourceField modeling.Field, actions *[]migration.ModifyField, isChange *bool) {
	// 字段是否发生改变
	fieldIsChange := false

	modifyFieldAction := migration.ModifyField{
		DisplayNameAttr:    targetField.DisplayNameAttr,
		SchemaNameAttr:     targetField.SchemaNameAttr,
		NewDisplayNameAttr: sourceField.DisplayNameAttr,
		NewSchemaNameAttr:  sourceField.SchemaNameAttr,
		IsOriginalAttr:     targetField.IsOriginalAttr,
		IsNullAttr:         targetField.IsNullAttr,
		DefaultConstraint:  targetField.DefaultConstraint,
		DataTypeAttr:       targetField.DataTypeAttr,
		TypeOption:         targetField.TypeOption,
	}

	if targetField.SchemaNameAttr != sourceField.SchemaNameAttr {
		fieldIsChange = true
	}

	if targetField.DisplayNameAttr != sourceField.DisplayNameAttr {
		fieldIsChange = true
	}

	if targetField.IsNullAttr != sourceField.IsNullAttr {
		modifyFieldAction.IsNullAttr = sourceField.IsNullAttr
		fieldIsChange = true
	}

	if targetField.IsOriginalAttr != sourceField.IsOriginalAttr {
		modifyFieldAction.IsOriginalAttr = sourceField.IsOriginalAttr
		fieldIsChange = true
	}

	if targetField.DataTypeAttr != sourceField.DataTypeAttr {
		modifyFieldAction.DataTypeAttr = sourceField.DataTypeAttr
		modifyFieldAction.TypeOption = sourceField.TypeOption
		fieldIsChange = true
	} else if *targetField.TypeOption != *sourceField.TypeOption {
		modifyFieldAction.TypeOption = sourceField.TypeOption
		fieldIsChange = true
	}

	if targetField.DefaultConstraint != nil && sourceField.DefaultConstraint != nil {
		if *targetField.DefaultConstraint != *sourceField.DefaultConstraint {
			modifyFieldAction.DefaultConstraint = sourceField.DefaultConstraint
			fieldIsChange = true
		}
	} else if targetField.DefaultConstraint != nil || sourceField.DefaultConstraint != nil {
		modifyFieldAction.DefaultConstraint = sourceField.DefaultConstraint
		fieldIsChange = true
	}

	if fieldIsChange {
		// 如果字段发生改变 则实体一定改变
		*isChange = true
		*actions = append(*actions, modifyFieldAction)
	}
}

// GetChangeIndexAction 获取索引的修改操作，包括创建  删除
func GetChangeIndexAction(targetIndexes *[]modeling.Index, sourceIndexes *[]modeling.Index, changeEntityAction *migration.ChangeEntityAction, isChange *bool) {
	tagIndexes := make([]int, 0)

	if targetIndexes == nil {
		targetIndexes = &[]modeling.Index{}
	}
	if sourceIndexes == nil {
		sourceIndexes = &[]modeling.Index{}
	}

	for i := 0; i < len(*sourceIndexes); i++ {
		tagIndexes = append(tagIndexes, 0)
	}

	// 寻找同名索引
	for _, targetIndexValue := range *targetIndexes {
		isFind := FindSameIndex(targetIndexValue, sourceIndexes, &tagIndexes)

		// 没有寻找到则代表为删除
		if !isFind {
			*changeEntityAction.RemoveIndexes = append(*changeEntityAction.RemoveIndexes, migration.RemoveItem{
				SchemaNameAttr: targetIndexValue.SchemaNameAttr,
			})
			*isChange = true
		}
	}

	// 未被标记代表为新建
	for index, value := range tagIndexes {
		if value == 0 {
			*changeEntityAction.NewIndexes = append(*changeEntityAction.NewIndexes, migration.Index{
				SchemaNameAttr:  (*sourceIndexes)[index].SchemaNameAttr,
				DisplayNameAttr: (*sourceIndexes)[index].DisplayNameAttr,
				IsPrimaryAttr:   (*sourceIndexes)[index].IsPrimaryAttr,
				Columns:         (*sourceIndexes)[index].Columns,
			})

			*isChange = true
		}
	}
}

// FindSameIndex 寻找相同的索引
func FindSameIndex(targetIndex modeling.Index, sourceIndexes *[]modeling.Index, tagIndexes *[]int) bool {
	for sourceIndexIndex, sourceIndexValue := range *sourceIndexes {
		if (*tagIndexes)[sourceIndexIndex] == 1 {
			continue
		} else if targetIndex.SchemaNameAttr == sourceIndexValue.SchemaNameAttr {
			(*tagIndexes)[sourceIndexIndex] = 1
			return true
		}
	}

	return false
}

// GetChangeUniqueConstraintAction 获取唯一约束的修改操作，包括创建  删除
func GetChangeUniqueConstraintAction(targetUniqueConstraint *[]modeling.UniqueConstraint, sourceUniqueConstraint *[]modeling.UniqueConstraint, changeEntityAction *migration.ChangeEntityAction, isChange *bool) {
	tagUniqueConstraints := make([]int, 0)

	if sourceUniqueConstraint == nil {
		sourceUniqueConstraint = &[]modeling.UniqueConstraint{}
	}

	if targetUniqueConstraint == nil {
		targetUniqueConstraint = &[]modeling.UniqueConstraint{}
	}

	for i := 0; i < len(*sourceUniqueConstraint); i++ {
		tagUniqueConstraints = append(tagUniqueConstraints, 0)
	}

	for _, targetIndexValue := range *targetUniqueConstraint {
		isFind := FindSameUniqueConstraintAction(targetIndexValue, sourceUniqueConstraint, &tagUniqueConstraints)

		if !isFind {
			*changeEntityAction.RemoveUniqueConstraints = append(*changeEntityAction.RemoveUniqueConstraints, migration.RemoveItem{
				SchemaNameAttr: targetIndexValue.SchemaNameAttr,
			})
			*isChange = true
		}
	}

	for index, value := range tagUniqueConstraints {
		if value == 0 {
			*changeEntityAction.NewUniqueConstraints = append(*changeEntityAction.NewUniqueConstraints, migration.UniqueConstraint{
				SchemaNameAttr:  (*sourceUniqueConstraint)[index].SchemaNameAttr,
				DisplayNameAttr: (*sourceUniqueConstraint)[index].DisplayNameAttr,
				Columns:         (*sourceUniqueConstraint)[index].Columns,
			})

			*isChange = true
		}
	}
}

// FindSameUniqueConstraintAction 寻找相同的唯一约束
func FindSameUniqueConstraintAction(targetIndex modeling.UniqueConstraint, sourceUniqueConstraint *[]modeling.UniqueConstraint, tagUniqueConstraints *[]int) bool {
	for sourceIndexIndex, sourceIndexValue := range *sourceUniqueConstraint {
		if (*tagUniqueConstraints)[sourceIndexIndex] == 1 {
			continue
		} else if targetIndex.SchemaNameAttr == sourceIndexValue.SchemaNameAttr {
			(*tagUniqueConstraints)[sourceIndexIndex] = 1
			return true
		}
	}

	return false
}
