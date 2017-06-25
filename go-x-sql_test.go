package sql_test

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	sql "github.com/marcuswestin/go-x-sql"
)

type Person struct {
	Id        uint64
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

	firstName := "Marcus"
	lastName := "Westin"
	age := 31
	// TODO Use RETURNING Id to get Id: https://www.cockroachlabs.com/docs/insert.html#go
	err = db.InsertIgnoreId("INSERT INTO person (first_name, last_name, age) VALUES (?, ?, ?)", firstName, lastName, age)
	if err != nil {
		t.Fatal(err)
	}

	var person Person = Person{}
	err = db.SelectOne(&person, "SELECT * FROM person WHERE first_name=?", firstName)
	if err != nil {
		t.Fatal(err)
	}

	if person.FirstName != firstName || person.LastName != lastName || person.Age != age {
		t.Fatal("Selected row don't match expected values", person, firstName, lastName, age)
	}
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

	firstName := "Marcus"
	lastName := "Westin"
	age := 31
	id, err := db.Insert("INSERT INTO Person SET FirstName=?, LastName=?, Age=?", firstName, lastName, age)
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Fatal("Expected Id to be 1:", id)
	}

	var person Person = Person{}
	err = db.SelectOne(&person, "SELECT * FROM Person WHERE Id=?", id)
	if err != nil {
		t.Fatal(err)
	}

	if person.FirstName != firstName || person.LastName != lastName || person.Age != age {
		t.Fatal("Selected row don't match expected values", person, firstName, lastName, age)
	}
}
