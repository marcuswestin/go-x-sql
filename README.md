# go-x-sql
Improved sql package for go

## TODO

- Abstract out mysql/cockroachdb/pg differences
    + X Setting to replace `?, ?, ?` with `$1, $2, $3` for postgres/cockroachdb
    + Use RETURNING Id to get Id from cockroachdb: https://www.cockroachlabs.com/docs/insert.html#go
    + Improve error handling: see http://go-database-sql.org/errors.html#identifying-specific-database-errors (and use https://github.com/VividCortex/mysqlerr for mysql errors, and pg for pg errors)
    + Allow for `INSERT INTO SET col1=?, col2=? in pg/cockroachdb`
    + Fix go-x-sql_test.go to share more code for pg and mysql tests

## No prepared statements?

No. They cause more problems than they're worth. See http://go-database-sql.org/prepared.html#avoiding-prepared-statements and http://go-database-sql.org/prepared.html#prepared-statements-in-transactions for details.