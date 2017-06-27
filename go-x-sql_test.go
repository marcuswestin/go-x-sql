package sql_test

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	sql "github.com/marcuswestin/go-x-sql"
)

type Person struct {
	Id        int64
	FirstName string
	LastName  string
	Age       int
}

var testDb = "goxsqltestdb"

func TestCockroachDB(t *testing.T) {
	var err error
	db := sql.MustConnect("postgres", "postgres://root@localhost/?sslmode=disable&port=5432", sql.DbNameConvention_under_score)

	db.MustExec("DROP DATABASE IF EXISTS " + testDb)
	defer db.MustExec("DROP DATABASE " + testDb)
	db.MustExec("CREATE DATABASE " + testDb)
	db.MustExec("SET database= " + testDb)
	db.MustExec(`CREATE TABLE Person (
		id SERIAL,
		first_name STRING(255),
		last_name STRING(255),
		age INT,
		PRIMARY KEY (id)
	);`)

	expected := Person{0, "Marcus", "Westin", 31}
	expected.Id, err = db.InsertAndGetId(`
		INSERT INTO person (first_name, last_name, age) VALUES (?, ?, ?) RETURNING id`,
		expected.FirstName, expected.LastName, expected.Age)
	if err != nil {
		t.Fatal(err)
	}
	checkValues(t, db, expected)
}

func TestMysql(t *testing.T) {
	var err error
	db := sql.MustConnect("mysql", "root:@/", sql.DbNameConventionCamelCase_Capitalized)

	db.MustExec("DROP DATABASE IF EXISTS " + testDb)
	defer db.MustExec("DROP DATABASE " + testDb)
	db.MustExec("CREATE DATABASE " + testDb)
	db.MustExec("USE " + testDb)
	db.MustExec(`CREATE TABLE Person (
		Id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		FirstName VARCHAR(255),
		LastName VARCHAR(255),
		Age INT,
		PRIMARY KEY (Id)
	);`)

	expected := Person{0, "Marcus", "Westin", 31}
	expected.Id, err = db.InsertAndGetId(`
		INSERT INTO Person SET FirstName=?, LastName=?, Age=?`,
		expected.FirstName, expected.LastName, expected.Age)
	if err != nil {
		t.Fatal(err)
	}
	checkValues(t, db, expected)
}

func checkValues(t *testing.T, db sql.Db, expected Person) {
	if expected.Id == 0 {
		t.Fatal("Expected an ID")
	}
	expected.Age *= 2
	err := db.UpdateOne(`UPDATE Person SET Age=?`, expected.Age)
	if err != nil {
		t.Fatal(err)
	}

	var person Person
	err = db.SelectOne(&person, "SELECT * FROM person WHERE id=?", expected.Id)
	if err != nil {
		t.Fatal(err)
	}

	if person.FirstName != expected.FirstName || person.LastName != expected.LastName || person.Age != expected.Age {
		t.Fatal("Selected row don't match expected values", person, expected)
	}
}
