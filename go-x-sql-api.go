package sql

import (
	"database/sql"
	"errors"
)

// API
//////

type Db interface {
	API
	Transact(TxFn) error
}
type API interface {
	Select(dest interface{}, query string, args ...interface{}) error
	SelectOne(dest interface{}, query string, args ...interface{}) error
	SelectOneMaybe(dest interface{}, query string, args ...interface{}) (bool, error)
	Insert(query string, args ...interface{}) (id int64, err error)
	InsertIgnoreId(query string, args ...interface{}) error
	InsertIgnoreDuplicate(query string, args ...interface{}) (bool, error)
	Update(query string, args ...interface{}) (rowsAffected int64, err error)
	UpdateOne(query string, args ...interface{}) error
	UpdateNum(expectedRowsAffected int64, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) error
	MustExec(query string, args ...interface{})
}
type TxFn func(tx Tx) error
type Tx interface {
	API
}

// DbNameConventionMapper is a function that takes a go struct field name and
// returns its corresponding database column name.
type DbNameConventionMapper func(goCamelCaseName string) (dbColumnName string)

// Connect
//////////

func Connect(driverName, dataSourceString string, dbNamesMapper DbNameConventionMapper) (Db, error) {
	return connect(driverName, dataSourceString, dbNamesMapper)
}

func MustConnect(driverName, dataSourceString string, dbNamesMapper DbNameConventionMapper) Db {
	return mustConnect(driverName, dataSourceString, dbNamesMapper)
}

// Errors
/////////

var (
	ErrNoRows       = sql.ErrNoRows
	ErrRowsAffected = errors.New("Affected unexpected number of rows")
)

var (
	Must     = must
	MustInt  = mustInt
	MustBool = mustBool

	CheckDest = checkDest
)

// Column name mappers
//////////////////////

var (
	DbNameConventionCamelCase_Capitalized   = func(str string) string { return str }
	DbNameConventionCamelCase_uncapitalized = func(str string) string { return uncapitalize(str) }
	DbNameConvention_under_score            = func(str string) string { return camelCaseTo_under_score(str) }
	DbNameConventionUPPERCASE               = func(str string) string { return toUPPERCASE(str) }
	DbNameConventionUPPER_CASE_UNDER_SCORE  = func(str string) string { return toUPPERCASE(camelCaseTo_under_score(str)) }
)

// Misc
///////

func IsDuplicateEntryError(err error) bool {
	return isDuplicateEntryError(err)
}
