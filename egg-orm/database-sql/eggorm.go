package eggorm

import (
	"database/sql"
	"eggorm/log"
	"eggorm/session"
)

type Engine struct {
	db *sql.DB
}

func NewEngine(driverName, databaseName string) (e *Engine, err error) {
	db, err := sql.Open(driverName, databaseName)
	if err != nil {
		log.Error(err)
		return
	}
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}
	e = &Engine{db: db}
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
	return session.NewSession(engine.db)
}
