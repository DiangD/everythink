package session

import (
	"eggorm/log"
	"reflect"
)

// 钩子的类型
const (
	BeforeInsert = "BeforeInsert"
	AfterInsert  = "AfterInsert"

	BeforeDelete = "BeforeDelete"
	AfterDelete  = "AfterDelete"

	BeforeUpdate = "BeforeUpdate"
	AfterUpdate  = "AfterUpdate"

	BeforeQuery = "BeforeQuery"
	AfterQuery  = "AfterQuery"
)

//CallMethod 通过反射调用函数
func (s *Session) CallMethod(method string, value interface{}) {
	fm := reflect.ValueOf(s.RefTable().Model).MethodByName(method)
	if value != nil {
		fm = reflect.ValueOf(value).MethodByName(method)
	}

	in := []reflect.Value{reflect.ValueOf(s)}
	if fm.IsValid() {
		if v := fm.Call(in); len(v) > 0 {
			if err, ok := v[0].Interface().(error); ok {
				log.Error(err)
			}
		}
	}
	return
}
