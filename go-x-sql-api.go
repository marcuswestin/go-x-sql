package sql

import (
	"database/sql"
	"errors"
)

// API
//////

type Db interface {
	API
	SetColumnNameMapperFn(ColumnNameMapperFn)
	Transact(TxFn) error
}
type API interface {
	Select(dest interface{}, query string, args ...interface{}) error
	SelectOne(dest interface{}, query string, args ...interface{}) error
	SelectOneMaybe(dest interface{}, query string, args ...interface{}) (bool, error)
	Insert(query string, args ...interface{}) (id int64, err error)
	InsertIgnoreId(query string, args ...interface{}) error
	InsertIgnoreDuplicate(query string, args ...interface{}) error
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

// ColumnNameMapperFn is a function that takes a database column name and
// returns its corresponding golang struct field name. The default column
// mapper function is ColumnNameMapperSame
type ColumnNameMapperFn func(dbColumnName string) (structFieldName string)

// Connect
//////////

func Connect(driverName, dataSourceString string) (Db, error) {
	return connect(driverName, dataSourceString)
}

func MustConnect(driverName, dataSourceString string) Db {
	return mustConnect(driverName, dataSourceString)
}

// Errors
/////////

var (
	ErrNoRows       = sql.ErrNoRows
	ErrRowsAffected = errors.New("Affected unexpected number of rows")
)

// Column name mappers
//////////////////////

var (
	ColumnNameMapperSame                  = func(col string) string { return col }
	ColumnNameMapperUnderscoreToCamelCase = func(col string) string { return underscoreToCamelCase(col) }
	ColumnNameMapperCapitalize            = func(col string) string { return capitalizeString(col) }
)

// Misc
///////

func IsDuplicateEntryError(err error) bool {
	return isDuplicateEntryError(err)
}
