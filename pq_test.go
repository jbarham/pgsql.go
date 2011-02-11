package pq

import (
	"fmt"
	"testing"
)

var rows = []string{"1, 1.0, 'one'", "2, 2.0, 'two'", "3, 3.0, 'three'"}

func TestPq(t *testing.T) {
	conn, err := Connect("dbname=testdb")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	err = conn.Exec("CREATE TABLE gopq_test (i int, f float, s text)")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Exec("DROP TABLE gopq_test")
	for _, row := range rows {
		err = conn.Exec(fmt.Sprintf("INSERT INTO gopq_test VALUES (%s)", row))
		if err != nil {
			t.Fatal(err)
		}
	}
	res, err := conn.Query("SELECT * FROM gopq_test")
	if err != nil {
		t.Fatal(err)
	}
	for res.Next() {
		fmt.Println(res.Row())
		fmt.Println(res.RowMap())
	}
	res.Clear()
}
