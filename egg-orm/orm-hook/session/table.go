package session

import (
	"eggorm/log"
	"eggorm/schema"
	"fmt"
	"reflect"
	"strings"
)

//session DCL操作

func (s *Session) Model(value interface{}) *Session {
	if s.refTable == nil || reflect.TypeOf(value) != reflect.TypeOf(s.refTable.Model) {
		s.refTable = schema.Parse(value, s.dialect)
	}
	return s
}

func (s *Session) RefTable() *schema.Schema {
	if s.refTable == nil {
		log.Error("Model is not set")
	}
	return s.refTable
}

func (s *Session) CreateTable() error {
	table := s.RefTable()
	var columns []string
	for _, field := range table.Fields {
		columns = append(columns, fmt.Sprintf("%s %s %s", field.Name, field.Type, field.Tag))
	}
	desc := strings.Join(columns, ",")
	sql := fmt.Sprintf("CREATE TABLE %s (%s);", table.Name, desc)
	_, err := s.Raw(sql).Exec()
	return err
}

func (s *Session) DropTable() error {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS %s", s.RefTable().Name)
	_, err := s.Raw(sql).Exec()
	return err
}

func (s *Session) HasTable() bool {
	sql, value := s.dialect.TableExistSQL(s.RefTable().Name)
	var tmp string
	_ = s.Raw(sql, value...).QueryRow().Scan(&tmp)
	return tmp == s.RefTable().Name
}
