package sql

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/jmoiron/sqlx"
)

type db struct {
	sqlxDb *sqlx.DB
	api
}
type api struct {
	sqlxAPI sqlxAPI
}

type sqlxAPI interface {
	sqlx.Execer
	sqlx.Queryer
	Select(dest interface{}, query string, arg ...interface{}) error
	Get(dest interface{}, query string, arg ...interface{}) error
}

// Connect
//////////

func mustConnect(driverName, dataSourceString string) Db {
	db, err := Connect(driverName, dataSourceString)
	if err != nil {
		panic(err)
	}
	return db
}
func connect(driverName, dataSourceString string) (Db, error) {
	sqlxAPI, err := sqlx.Open(driverName, dataSourceString)
	if err != nil {
		return nil, err
	}
	if err := sqlxAPI.Ping(); err != nil {
		return nil, errors.New("Connect error: could not ping database (" + err.Error() + ")")
	}
	sqlxAPI.MapperFunc(ColumnNameMapperSame)
	// sqlxAPI = sqlxAPI.Unsafe()
	return &db{sqlxAPI, api{sqlxAPI}}, nil
}

// Selects
//////////

func (api *api) Select(dest interface{}, query string, args ...interface{}) error {
	return api.sqlxAPI.Select(dest, query, args...)
}

func (api *api) SelectOne(dest interface{}, query string, args ...interface{}) error {
	return api.sqlxAPI.Get(dest, query, args...)
}

func (api *api) SelectOneMaybe(dest interface{}, query string, args ...interface{}) (bool, error) {
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

func (api *api) Insert(query string, args ...interface{}) (id int64, err error) {
	res, err := api.sqlxAPI.Exec(query, args...)
	if err != nil {
		return
	}
	return res.LastInsertId()
}

func (api *api) InsertIgnoreId(query string, args ...interface{}) error {
	_, err := api.sqlxAPI.Exec(query, args...)
	return err
}

func (api *api) InsertIgnoreDuplicate(query string, args ...interface{}) error {
	err := api.InsertIgnoreId(query, args...)
	if err == nil || IsDuplicateEntryError(err) {
		return nil
	} else {
		return err
	}
}

// Updates
//////////

func (api *api) Update(query string, args ...interface{}) (rowsAffected int64, err error) {
	res, err := api.sqlxAPI.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (api *api) UpdateOne(query string, args ...interface{}) error {
	return api.UpdateNum(1, query, args...)
}

func (api *api) UpdateNum(expectedRowsAffected int64, query string, args ...interface{}) error {
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
	_, err := api.sqlxAPI.Exec(query, args...)
	return err
}
func (api *api) MustExec(query string, args ...interface{}) {
	err := api.Exec(query, args...)
	if err != nil {
		panic(err)
	}
}

// Top-level Db API: Column mapper & transactions
/////////////////////////////////////////////////

func (db *db) SetColumnNameMapperFn(mapperFn ColumnNameMapperFn) {
	db.sqlxDb.MapperFunc(mapperFn)
}

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
	txErr := txFunc(&api{txConn})
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

// Util
///////

func isDuplicateEntryError(err error) bool {
	return err != nil && (false ||
		// mysql:
		strings.Contains(err.Error(), "Duplicate entry") ||
		// cockroachdb:
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

func underscoreToCamelCase(under_scored_string string) (camelCasedString string) {
	// from https://github.com/asaskevich/govalidator/blob/master/utils.go#L101
	return strings.Replace(
		strings.Title(
			strings.Replace(
				strings.ToLower(under_scored_string),
				"_", " ", -1)),
		" ", "", -1)
}

func capitalizeString(lowercaseString string) (CapitalizedString string) {
	return strings.Title(lowercaseString)
}
