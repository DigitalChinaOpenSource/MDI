package dbbase

import (
	"errors"
	"fmt"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCountSQL(t *testing.T) {

	dbContent := new(DbContent)
	tableName := "Users"
	// 测试查询语句的生成

	dbSQL, err := dbContent.GenerateCountSQL(tableName)
	fmt.Println(dbSQL.SqlStr)
	//断言
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(dbSQL.SqlStr)
		expectedSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		fmt.Println(assert.Equal(t, tableName, dbSQL.TableName))
		fmt.Println(assert.Equal(t, expectedSQL, dbSQL.SqlStr))
	}

}

func TestValid(t *testing.T) {
	dbContent := new(DbContent)
	var tableName *string
	dbContent.AllTablesOfSchema = append(dbContent.AllTablesOfSchema, "User", "Person")
	dbContent.AllFieldsOfTable = map[string][]Column{
		"User":   {Column{FieldName: "user_id", DataType: "int4", DataSize: 32}, Column{FieldName: "user_name", DataType: "string", DataSize: 200}},
		"Person": {Column{FieldName: "person_id", DataType: "int4", DataSize: 32}, Column{FieldName: "person_name", DataType: "string", DataSize: 200}},
	}

	name := "User"
	tableName = &name
	queryFields := []string{"user_id", "user_name"}

	// 测试查询语句的生成
	isValid, err := dbContent.Valid(tableName, queryFields)
	fmt.Println(isValid)
	if err != nil {
		t.Fatal(err)
	} else {
		//断言
		t.Log(isValid)
		var fields []string
		fields = append(fields, dbContent.AllFieldsOfTable["User"][0].FieldName, dbContent.AllFieldsOfTable["User"][1].FieldName)

		boolDate := assert.Equal(t, fields, queryFields)

		fmt.Println(assert.Equal(t, boolDate, isValid))

	}

}

func TestToFilterStr(t *testing.T) {

	col1 := Column{FieldName: "str", DataType: "string", DataSize: 200}
	var filterField1 *FilterField
	filterField1 = &FilterField{
		Field:         FieldCell{OriginData: "test", BelongColumn: &col1},
		CompareOption: "=",
	}
	col2 := Column{FieldName: "strSQL", DataType: "int4", DataSize: 200}
	//var filterField2 *FilterField  等价于 ↓
	filterField2 := new(FilterField)
	filterField2 = &FilterField{
		//Field: FieldCell{OriginData: "' or 1=1 # ", BelongColumn: &col2},
		Field:         FieldCell{OriginData: 1004, BelongColumn: &col2},
		CompareOption: "=",
	}

	// 测试过滤语句
	//过滤方法没有处理sql注入
	strFilter1, err := filterField1.ToFilterStr()
	//对于注入测试
	strFilter2, err := filterField2.ToFilterStr()

	if err != nil {
		t.Fatal(err)
	} else {

		//断言
		strExpected1 := "str = 'test'"
		fmt.Println(assert.Equal(t, strExpected1, strFilter1))
		strExpected2 := "`strSQL` = 1004"
		fmt.Println(assert.Equal(t, strExpected2, strFilter2))

	}

}

func TestToFileldStr(t *testing.T) {

	col1 := Column{FieldName: "str", DataType: "string", DataSize: 200}
	var filterField1 *FieldCell
	filterField1 = &FieldCell{OriginData: "test", BelongColumn: &col1}

	col2 := Column{FieldName: "isOk", DataType: "bool", DataSize: 200}
	var filterField2 *FieldCell
	filterField2 = &FieldCell{OriginData: true, BelongColumn: &col2}

	// 测试过滤语句
	//过滤方法没有处理sql注入
	strFilter1 := filterField1.ToFieldStr()

	strFilter2 := filterField2.ToFieldStr()

	//断言
	strExpected1 := "'test'"
	fmt.Println(assert.Equal(t, strExpected1, strFilter1))
	strExpected2 := "true"
	fmt.Println(assert.Equal(t, strExpected2, strFilter2))

}
func TestDbRespondToMap(t *testing.T) {

	col := Column{FieldName: "user_id", DataType: "int4", DataSize: 32}
	col1 := Column{FieldName: "user_name", DataType: "string", DataSize: 200}
	errStr := "im error"
	err := errors.New(errStr)
	FieldCells1 := []FieldCell{}
	FieldCells2 := []FieldCell{}
	rows := &[]Row{}
	columns := &[]Column{}
	filed1 := FieldCell{
		OriginData:   1001,
		BelongColumn: &col,
	}
	filed2 := FieldCell{
		OriginData:   "dog",
		BelongColumn: &col1,
	}

	filed3 := FieldCell{
		OriginData:   1002,
		BelongColumn: &col,
	}
	filed4 := FieldCell{
		OriginData:   "cat",
		BelongColumn: &col1,
	}

	FieldCells1 = append(FieldCells1, filed1, filed2)
	FieldCells2 = append(FieldCells2, filed3, filed4)
	row1 := Row{
		FieldCells1,
	}
	row2 := Row{
		FieldCells2,
	}
	*rows = append(*rows, row1, row2)
	*columns = append(*columns, col, col1)

	dbtable := DbTable{
		Rows: *rows,
		Cols: *columns,
	}
	respond := DbRespond{
		RespondStatus: true,
		RespondData:   dbtable,
		Err:           err,
	}

	//断言
	jsonStr, err := DbRespondToMap(respond)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(jsonStr)
	fmt.Println(jsonStr)
	er := jsonStr["Err"]
	fmt.Println(assert.Equal(t, errStr, er))

	respondStatus := jsonStr["RespondStatus"]
	fmt.Println(assert.Equal(t, true, respondStatus))

	respondData := jsonStr["RespondData"]
	rMap := respondData.([]map[string]interface{})

	fmt.Println(assert.Equal(t, filed1.OriginData, rMap[0]["user_id"]))
	fmt.Println(assert.Equal(t, filed2.OriginData, rMap[0]["user_name"]))
	fmt.Println(assert.Equal(t, filed3.OriginData, rMap[1]["user_id"]))
	fmt.Println(assert.Equal(t, filed4.OriginData, rMap[1]["user_name"]))

}

func TestGetDataArr(t *testing.T) {
	col := Column{FieldName: "user_id", DataType: "int4", DataSize: 32}
	col1 := Column{FieldName: "user_name", DataType: "string", DataSize: 200}
	err := errors.New("im error")
	FieldCells1 := []FieldCell{}
	FieldCells2 := []FieldCell{}
	rows := &[]Row{}
	columns := &[]Column{}
	filed1 := FieldCell{
		OriginData:   1001,
		BelongColumn: &col,
	}
	filed2 := FieldCell{
		OriginData:   "apple",
		BelongColumn: &col1,
	}

	filed3 := FieldCell{
		OriginData:   1002,
		BelongColumn: &col,
	}
	filed4 := FieldCell{
		OriginData:   "banana",
		BelongColumn: &col1,
	}

	FieldCells1 = append(FieldCells1, filed1, filed2)
	FieldCells2 = append(FieldCells2, filed3, filed4)
	row1 := Row{
		FieldCells1,
	}
	row2 := Row{
		FieldCells2,
	}
	*rows = append(*rows, row1, row2)
	*columns = append(*columns, col, col1)

	dbtable := DbTable{
		Rows: *rows,
		Cols: *columns,
	}
	respond := DbRespond{
		RespondStatus: true,
		RespondData:   dbtable,
		Err:           err,
	}

	//断言
	rMap, err := GetDataArr(respond)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(rMap)
	fmt.Println(assert.Equal(t, filed1.OriginData, rMap[0]["user_id"]))
	fmt.Println(assert.Equal(t, filed2.OriginData, rMap[0]["user_name"]))
	fmt.Println(assert.Equal(t, filed3.OriginData, rMap[1]["user_id"]))
	fmt.Println(assert.Equal(t, filed4.OriginData, rMap[1]["user_name"]))

}
