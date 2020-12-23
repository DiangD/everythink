package schema

import (
	"eggorm/dialect"
	"go/ast"
	"reflect"
)

//schema可理解为struct与数据库表的映射

type Field struct {
	Name string
	Type string
	Tag  string
}

type Schema struct {
	Model      interface{}
	Name       string
	Fields     []*Field
	FieldNames []string
	fieldMap   map[string]*Field
}

func (schema *Schema) GetField(fieldName string) *Field {
	return schema.fieldMap[fieldName]
}

//Parse 解析结构体生成schema
func Parse(dest interface{}, dialect dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(),
		fieldMap: make(map[string]*Field),
	}
	for i := 0; i < modelType.NumField(); i++ {
		f := modelType.Field(i)
		if !f.Anonymous && ast.IsExported(f.Name) {
			field := &Field{
				Name: f.Name,
				Type: dialect.DataTypeOf(reflect.Indirect(reflect.New(f.Type))),
			}
			if v, ok := f.Tag.Lookup("eggorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, f.Name)
			schema.fieldMap[f.Name] = field
		}
	}
	return schema
}

//RecordValues 平铺结构体属性 User{Name:"test",Age:10} => []interface{"test",10}
func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValue []interface{}
	for _, field := range schema.Fields {
		fieldValue = append(fieldValue, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValue
}
