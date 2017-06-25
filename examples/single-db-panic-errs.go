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

func Setup() {
	db = sql.MustConnect(driver, dataSourceString, db.DbNameConvention_under_score)
}

func Select(dest interface{}, query string, args ...interface{}) {
	sql.Must(sql.CheckDest(dest))
	sql.Must(db.Select(dest, query, args...))
}
func SelectOne(dest interface{}, query string, args ...interface{}) {
	sql.Must(sql.CheckDest(dest))
	sql.Must(db.SelectOne(dest, query, args...))
}
func SelectOneMaybe(dest interface{}, query string, args ...interface{}) bool {
	sql.Must(sql.CheckDest(dest))
	return sql.MustBool(db.SelectOneMaybe(dest, query, args...))
}
func Insert(query string, args ...interface{}) (id int64) {
	return sql.MustInt(db.Insert(query, args...))
}
func InsertIgnoreId(query string, args ...interface{}) {
	sql.Must(db.InsertIgnoreId(query, args...))
}
func InsertIgnoreDuplicate(query string, args ...interface{}) bool {
	sql.MustBool(db.InsertIgnoreDuplicate(query, args...))
}
func Update(query string, args ...interface{}) (rowsAffected int64) {
	return sql.MustInt(db.Update(query, args...))
}
func UpdateOne(query string, args ...interface{}) {
	sql.Must(db.UpdateOne(query, args...))
}
func UpdateNum(expectedRowsAffected int64, query string, args ...interface{}) {
	sql.Must(db.UpdateNum(expectedRowsAffected, query, args...))
}
func Exec(query string, args ...interface{}) {
	sql.Must(db.Exec(query, args...))
}

type TxFn func(tx Tx) error
type Tx sql.Tx

func Transact(txFn TxFn) {
	sql.Must(db.Transact(func(tx sql.Tx) error {
		return txFn(tx)
	}))
}
