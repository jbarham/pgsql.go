package pgsql

import (
	"os"
	"fmt"
	"time"
	"math"
	"bytes"
	"testing"
)

type rec struct {
	tf bool
	i32 int
	i64 int64
	s   string
	b   []byte
}

var rows = []rec{
	{false, math.MinInt32, math.MinInt64, "hello world", []byte{0xDE, 0xAD}},
	{true, math.MaxInt32, math.MaxInt64, "Γεια σας κόσμο", []byte{0xBE, 0xEF}},
}

func chkerr(t *testing.T, err os.Error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestPq(t *testing.T) {
	var err os.Error
	var res *Result

	conn, err := Connect("dbname=testdb")
	chkerr(t, err)
	defer conn.Close()

	// Create test table, and schedule its deletion.
	err = conn.Exec("CREATE TABLE gopq_test (tf bool, i32 int, i64 bigint, s text, b bytea)")
	chkerr(t, err)
	defer conn.Exec("DROP TABLE gopq_test")

	// Insert test rows.
	stmt, err := conn.Prepare("INSERT INTO gopq_test VALUES ($1, $2, $3, $4, $5)")
	chkerr(t, err)
	defer stmt.Clear()
	for _, row := range rows {
		err = stmt.Exec(row.tf, row.i32, row.i64, row.s, row.b)
		chkerr(t, err)
	}

	// Verify that all test rows were inserted.
	res, err = conn.Query("SELECT COUNT(*) FROM gopq_test")
	chkerr(t, err)
	if !res.Next() {
		t.Fatal("Result.Next failed")
	}
	var count int
	err = res.Scan(&count)
	chkerr(t, err)
	if count != len(rows) {
		t.Fatalf("invalid row count %d, expected %d", count, len(rows))
	}
	res.Clear()

	// Retrieve inserted rows and verify inserted values.
	res, err = conn.Query("SELECT * FROM gopq_test")
	chkerr(t, err)
	t.Log("cols:", res.Names())
	for i := 0; res.Next(); i++ {
		t.Log("row:", i)
		var tf bool
		var i32 int
		var i64 int64
		var s string
		var b []byte

		err := res.Scan(&tf, &i32, &i64, &s, &b)
		if err != nil {
			t.Fatal("scan error:", err)
		}
		if tf != rows[i].tf {
			t.Fatal("bad bool")
		}
		if i32 != rows[i].i32 {
			t.Fatal("bad int32")
		}
		if i64 != rows[i].i64 {
			t.Fatal("bad int64")
		}
		if s != rows[i].s {
			t.Fatal("bad string")
		}
		if !bytes.Equal(b, rows[i].b) {
			t.Fatal("bad byte array")
		}
	}
	res.Clear()

	// Test timestamp parsing.
	res, err = conn.Query("SELECT now()")
	var now time.Time
	res.Next()
	if err = res.Scan(&now); err != nil {
		t.Fatal("time.Time parse error:", err)
	}
	t.Log("now:", now)
	res.Clear()

	// Test parsing of float values.
	var f32 float32
	res, err = conn.Query(fmt.Sprintf("SELECT %v", float32(math.MaxFloat32)))
	chkerr(t, err)
	res.Next()
	chkerr(t, res.Scan(&f32))
	if f32 != math.MaxFloat32 {
		t.Fatalf("float32 conversion error: expected %v, got %v", float32(math.MaxFloat32), f32)
	}
	res.Clear()
	var f64 float64
	res, err = conn.Query(fmt.Sprintf("SELECT %v", float64(math.MaxFloat64)))
	chkerr(t, err)
	res.Next()
	chkerr(t, res.Scan(&f64))
	if f64 != math.MaxFloat64 {
		t.Fatalf("float64 conversion error: expected %v, got %v", float64(math.MaxFloat64), f64)
	}
	res.Clear()
}
