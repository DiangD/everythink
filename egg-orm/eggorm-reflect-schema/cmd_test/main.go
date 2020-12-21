package main

import (
	"eggorm"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	engine, _ := eggorm.NewEngine("sqlite3", "egg.db")
	defer engine.Close()
	s := engine.NewSession()
	_, _ = s.Raw("drop table if exists User;").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text);").Exec()
	res, _ := s.Raw("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam").Exec()
	count, _ := res.RowsAffected()
	fmt.Printf("Exec success, %d affected\n", count)
}
