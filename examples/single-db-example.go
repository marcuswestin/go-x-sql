package db

// This package would be used like this:
//
// 	db.Driver = "mysql"
// 	db.Driver = "root:@/"
// 	db.Setup()
// 	db.Exec("Create Table Foo ();")

import (
	_ "github.com/go-sql-driver/mysql"
	// _ "github.com/lib/pq"
	sql "github.com/marcuswestin/go-x-sql"
)

var (
	db               sql.Db
	driver           = "mysql"
	dataSourceString = "root:@/"
	// driver           = "postgres"
	// dataSourceString = "postgres://root@localhost/pqgotest?sslmode=disable&port=5432"
)

func Connect() (err error) {
	db, err = sql.Connect(Driver, DataSourceString, sql.DbNameConventionCamelCase_uncapitalized)
	return
}
func MustConnect() {
	db = sql.MustConnect(Driver, DataSourceString)
}

func Select(dest interface{}, query string, args ...interface{}) error {
	return db.Select(dest, query, args...)
}
func SelectOne(dest interface{}, query string, args ...interface{}) error {
	return db.SelectOne(dest, query, args...)
}
func SelectOneMaybe(dest interface{}, query string, args ...interface{}) (bool, error) {
	return db.SelectOneMaybe(dest, query, args...)
}
func Insert(query string, args ...interface{}) (id int64, err error) {
	return db.Insert(query, args...)
}
func InsertIgnoreId(query string, args ...interface{}) error {
	return db.InsertIgnoreId(query, args...)
}
func InsertIgnoreDuplicate(query string, args ...interface{}) error {
	return db.InsertIgnoreDuplicate(query, args...)
}
func Update(query string, args ...interface{}) (rowsAffected int64, err error) {
	return db.Update(query, args...)
}
func UpdateOne(query string, args ...interface{}) error {
	return db.UpdateOne(query, args...)
}
func UpdateNum(expectedRowsAffected int64, query string, args ...interface{}) error {
	return db.UpdateNum(expectedRowsAffected, query, args...)
}
func Exec(query string, args ...interface{}) error {
	return db.Exec(query, args...)
}
func MustExec(query string, args ...interface{}) {
	db.MustExec(query, args...)
	return
}

type TxFn func(tx Tx) error
type Tx sql.Tx

func Transact(txFn TxFn) error {
	return db.Transact(func(tx sql.Tx) error {
		return txFn(tx)
	})
}
