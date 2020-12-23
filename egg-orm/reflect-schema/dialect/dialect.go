package dialect

import "reflect"

var dialectMap = map[string]Dialect{}

//Dialect 兼容多种数据库
type Dialect interface {
	//基本数据类型与数据库类型映射
	DataTypeOf(typ reflect.Value) string
	TableExistSQL(tableName string) (string, []interface{})
}

//RegisterDialect 注册方言
func RegisterDialect(name string, dialect Dialect) {
	dialectMap[name] = dialect
}

//GetDialect 获取方言
func GetDialect(name string) (Dialect, bool) {
	dialect, ok := dialectMap[name]
	return dialect, ok
}
