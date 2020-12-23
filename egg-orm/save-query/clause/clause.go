package clause

import "strings"

type Clause struct {
	sql     map[Type]string
	sqlVars map[Type][]interface{}
}

type Type int

const (
	INSERT Type = iota
	VALUES
	SELECT
	LIMIT
	WHERE
	ORDER
)

//填充clause
func (c *Clause) Set(name Type, vars ...interface{}) {
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]interface{})
	}
	sql, vars := generators[name](vars...)
	c.sql[name] = sql
	c.sqlVars[name] = vars
}

// 按传入order顺序构造sql
func (c *Clause) Build(orders ...Type) (string, []interface{}) {
	var (
		sqlStr []string
		vars   []interface{}
	)
	for _, order := range orders {
		if sql, ok := c.sql[order]; ok {
			sqlStr = append(sqlStr, sql)
			vars = append(vars, c.sqlVars[order]...)
		}
	}
	return strings.Join(sqlStr, " "), vars
}
