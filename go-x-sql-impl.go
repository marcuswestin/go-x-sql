package sql

import (
	"errors"
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"
)

type db struct {
	sqlxDb *sqlx.DB
	api
}
type api struct {
	sqlxAPI  sqlxAPI
	bindType int
	insertAndGetIdFn
}

type sqlxAPI interface {
	sqlx.Execer
	sqlx.Queryer
	Select(dest interface{}, query string, arg ...interface{}) error
	Get(dest interface{}, query string, arg ...interface{}) error
}

// Connect
//////////

func mustConnect(driverName, dataSourceString string, dbNamesMapper DbNameConventionMapper) Db {
	db, err := connect(driverName, dataSourceString, dbNamesMapper)
	if err != nil {
		panic(err)
	}
	return db
}
func connect(driverName, dataSourceString string, dbNamesMapper DbNameConventionMapper) (Db, error) {
	sqlxAPI, err := sqlx.Open(driverName, dataSourceString)
	if err != nil {
		return nil, err
	}
	if err := sqlxAPI.Ping(); err != nil {
		return nil, errors.New("Connect error: could not ping database (" + err.Error() + ")")
	}
	sqlxAPI.MapperFunc(dbNamesMapper)
	// sqlxAPI = sqlxAPI.Unsafe()
	return &db{sqlxAPI, api{sqlxAPI, sqlx.BindType(driverName), getInsertAndGetIdFnForDriver(driverName)}}, nil
}

// Selects
//////////

func (api *api) Select(dest interface{}, query string, args ...interface{}) error {
	query = fixQuery(api, query)
	return api.sqlxAPI.Select(dest, query, args...)
}

func (api *api) SelectOne(dest interface{}, query string, args ...interface{}) error {
	query = fixQuery(api, query)
	return api.sqlxAPI.Get(dest, query, args...)
}

func (api *api) SelectOneMaybe(dest interface{}, query string, args ...interface{}) (bool, error) {
	query = fixQuery(api, query)
	err := api.sqlxAPI.Get(dest, query, args...)
	if err == ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

// Inserts
//////////

func (api *api) InsertIgnoreId(query string, args ...interface{}) error {
	query = fixQuery(api, query)
	_, err := api.sqlxAPI.Exec(query, args...)
	return err
}

func (api *api) InsertIgnoreDuplicate(query string, args ...interface{}) (bool, error) {
	query = fixQuery(api, query)
	err := api.InsertIgnoreId(query, args...)
	if err == nil {
		return false, nil
	} else if IsDuplicateEntryError(err) {
		return true, nil
	} else {
		return false, err
	}
}

func (api *api) InsertAndGetId(query string, args ...interface{}) (id int64, err error) {
	query = fixQuery(api, query)
	return api.insertAndGetIdFn(api, query, args...)
}

type insertAndGetIdFn func(api *api, query string, args ...interface{}) (id int64, err error)

func getInsertAndGetIdFnForDriver(driverName string) insertAndGetIdFn {
	switch driverName {
	case "postgres", "pgx":
		return postgresInsertAndGetIdFn
	case "mysql", "sqlite3":
		return mysqlInsertAndGetIdFn
	default:
		panic("Unkown insert id function for driver " + driverName)
	}
}

func mysqlInsertAndGetIdFn(api *api, query string, args ...interface{}) (id int64, err error) {
	res, err := api.sqlxAPI.Exec(query, args...)
	if err != nil {
		return
	}
	return res.LastInsertId()
}

func postgresInsertAndGetIdFn(api *api, query string, args ...interface{}) (id int64, err error) {
	err = api.sqlxAPI.Get(&id, query, args...)
	return
}

// Updates
//////////

func (api *api) Update(query string, args ...interface{}) (rowsAffected int64, err error) {
	query = fixQuery(api, query)
	res, err := api.sqlxAPI.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (api *api) UpdateOne(query string, args ...interface{}) error {
	query = fixQuery(api, query)
	return api.UpdateNum(1, query, args...)
}

func (api *api) UpdateNum(expectedRowsAffected int64, query string, args ...interface{}) error {
	query = fixQuery(api, query)
	rowsAffected, err := api.Update(query, args...)
	if err != nil {
		return err
	} else if rowsAffected != expectedRowsAffected {
		return ErrRowsAffected
	} else {
		return nil
	}
}

// Exec
///////

func (api *api) Exec(query string, args ...interface{}) error {
	query = fixQuery(api, query)
	_, err := api.sqlxAPI.Exec(query, args...)
	return err
}
func (api *api) MustExec(query string, args ...interface{}) {
	query = fixQuery(api, query)
	Must(api.Exec(query, args...))
}

// Top-level Db API: Column mapper & transactions
/////////////////////////////////////////////////

func (db *db) Transact(txFunc TxFn) error {
	txConn, err := db.sqlxDb.Beginx()
	if err != nil {
		return err
	}

	// Attempt rollback during transaction panics
	defer func() {
		panicVal := recover()
		if panicVal != nil {
			var txErr = errors.New(fmt.Sprintf("%v", panicVal))
			rbErr := txConn.Rollback()
			if rbErr != nil {
				panic(txError("Transaction panic, rollback failed:", txErr, nil, rbErr))
			} else {
				panic(txError("Transaction panic, rollback succeeded:", txErr, nil, rbErr))
			}
		}
	}()

	// Execution transaction function.
	// Rollback on error, Commit otherwise.
	txErr := txFunc(&api{txConn, db.api.bindType, db.api.insertAndGetIdFn})
	if txErr != nil {
		rbErr := txConn.Rollback()
		if rbErr != nil {
			return txError("Transaction error, rollback failed:", txErr, nil, rbErr)
		} else {
			return txError("Transaction error, rollback succeeded:", txErr, nil, rbErr)
		}
	} else {
		commitErr := txConn.Commit()
		if commitErr != nil {
			return txError("Transaction commit failed:", nil, commitErr, nil)
		}
	}

	// All done!
	return nil
}

// Checks
/////////

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func mustInt(i int64, err error) int64 {
	if err != nil {
		panic(err)
	}
	return i
}
func mustBool(b bool, err error) bool {
	if err != nil {
		panic(err)
	}
	return b
}

func checkDest(dest interface{}) error {
	if _, isString := dest.(string); isString {
		return errors.New("Destination is string")
	}
	// TODO: Check that it's a pointer type?
	return nil
}

// Util
///////

func fixQuery(api *api, query string) string {
	return sqlx.Rebind(api.bindType, query)
}

func isDuplicateEntryError(err error) bool {
	return err != nil && (false ||
		// mysql:
		strings.Contains(err.Error(), "Duplicate entry") ||
		// cockroachdb:
		strings.Contains(err.Error(), "duplicate key value") ||
		// ???
		false)
}

func txError(title string, txErr error, commitErr error, rbErr error) error {
	var commitErrMsg = ""
	var rbErrMsg = ""
	if commitErr != nil {
		commitErrMsg = "\n	Commit error: " + commitErr.Error()
	}
	if rbErr != nil {
		rbErrMsg = "\n	Rollback error: " + rbErr.Error()
	}
	return errors.New(title +
		"\n	Transaction error:" + txErr.Error() +
		commitErrMsg +
		rbErrMsg +
		"\n	Stack trace: " + string(debug.Stack()))
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func camelCaseTo_under_score(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func toUPPERCASE(str string) string {
	return strings.ToUpper(str)
}

func uncapitalize(str string) string {
	if str == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(str)
	return string(unicode.ToLower(r)) + str[n:]
}
