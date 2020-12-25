package eggorm

import (
	"database/sql"
	"eggorm/dialect"
	"eggorm/log"
	"eggorm/session"
)

//Engine 用户操作数据库的入口
type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

func NewEngine(driverName, databaseName string) (e *Engine, err error) {
	//db连接
	db, err := sql.Open(driverName, databaseName)
	if err != nil {
		log.Error(err)
		return
	}
	//测试连接成功
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}
	d, ok := dialect.GetDialect(driverName)
	if !ok {
		log.Errorf("d %s Not Found", driverName)
		return
	}
	e = &Engine{db: db, dialect: d}
	log.Info("Connect database success")
	return
}

func (engine *Engine) Close() {
	if err := engine.db.Close(); err != nil {
		log.Error(err)
	}
	log.Info("Close database success")
}

func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db, engine.dialect)
}

type TxFunc func(s *session.Session) (interface{}, error)

func (engine *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	s := engine.NewSession()
	if err = s.Begin(); err != nil {
		log.Error(err)
		return
	}
	defer func() {
		if p := recover(); p != nil {
			_ = s.Rollback()
			panic(p)
		} else if err != nil {
			_ = s.Rollback()
		} else {
			err = s.Commit()
		}
	}()
	return f(s)
}
