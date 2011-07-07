// Copyright 2011 John E. Barham. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package pgsql provides high-level access to the PostgreSQL libpq client access library.
package pgsql

/*
#include <stdlib.h>
#include <libpq-fe.h>

static char**makeCharArray(int size) {
	return calloc(sizeof(char*), size);
}

static void setArrayString(char **a, char *s, int n) {
	a[n] = s;
}

static void freeCharArray(char **a, int size) {
	int i;
	for (i = 0; i < size; i++)
		free(a[i]);
	free(a);
}
*/
import "C"

import (
	"os"
	"fmt"
	"time"
	"runtime"
	"unsafe"
	"reflect"
	"strings"
	"strconv"
	"regexp"
	"encoding/hex"
)

// Errors
var (
	ErrMissingDbName = os.NewError("missing dbname")
)

func connError(db *C.PGconn) os.Error {
	return os.NewError("conn error:" + C.GoString(C.PQerrorMessage(db)))
}

func resultError(res *C.PGresult) os.Error {
	serr := C.GoString(C.PQresultErrorMessage(res))
	if serr == "" {
		return nil
	}
	return os.NewError("result error: " + serr)
}

const timeFormat = "2006-01-02 15:04:05-07:00"

// PostgreSQL ISO timestamp format: YYYY-mm-dd HH:MM:SS[.SSS]±TZ[:TZ]
var tsRe = regexp.MustCompile("^([0-9]+-[0-9]+-[0-9]+ [0-9]+:[0-9]+:[0-9]+)(\\.[0-9]+)([\\+\\-]?[0-9:]+)$")

// ParseTimestamp parses the given PostgreSQL timestamp string and upon success
// returns a time.Time and the fractional seconds in the timestamp, since the
// standard time.Time can only represent whole seconds.
func ParseTimestamp(ts string) (t *time.Time, fraction float64, err os.Error) {
	// Extract seconds fraction.
	m := tsRe.FindStringSubmatch(ts)
	if m == nil {
		return nil, 0, os.NewError("invalid timestamp")
	}
	if m[2] != "" {
		fraction, _ = strconv.Atof64(m[2])
	}
	// Parse timestamp excluding seconds fraction.
	t, err = time.Parse(timeFormat, m[1]+m[3])
	return
}

type Result struct {
	res     *C.PGresult
	nrows   int
	currRow int
	ncols   int
	cols    []string
}

func newResult(res *C.PGresult) *Result {
	ncols := int(C.PQnfields(res))
	nrows := int(C.PQntuples(res))
	result := &Result{res: res, nrows: nrows, currRow: -1, ncols: ncols, cols: nil}
	runtime.SetFinalizer(result, (*Result).Clear)
	return result
}

// Names returns the list of column (field) names, in order, in the result.
func (r *Result) Names() []string {
	if r.cols == nil {
		r.cols = make([]string, r.ncols)
		for i := 0; i < r.ncols; i++ {
			r.cols[i] = C.GoString(C.PQfname(r.res, C.int(i)))
		}
	}
	return r.cols
}

// Next returns true and increments the result row index if there are any
// remaining rows in the result; otherwise it returns false.
// Next should be called, and its value checked, before processing the first
// row of a result using Result.Scan.
func (r *Result) Next() bool {
	if r.currRow+1 < r.nrows {
		r.currRow++
		return true
	}
	return false
}

func argErr(i int, argType string, err string) os.Error {
	return os.NewError(fmt.Sprintf("arg %d as %s: %s", i, argType, err))
}

// Scan parses the values of the current result row (set using Result.Next)
// into the given arguments.
func (r *Result) Scan(args ...interface{}) os.Error {
	if len(args) != r.ncols {
		return os.NewError(fmt.Sprintf("incorrect argument count for Result.Scan: have %d want %d", len(args), r.ncols))
	}

	for i, v := range args {
		if int(C.PQgetisnull(r.res, C.int(r.currRow), C.int(i))) == 1 {
			continue
		}
		val := C.GoString(C.PQgetvalue(r.res, C.int(r.currRow), C.int(i)))
		switch v := v.(type) {
		case *[]byte:
			if !strings.HasPrefix(val, "\\x") {
				return argErr(i, "[]byte", "invalid byte string format")
			}
			buf, err := hex.DecodeString(val[2:])
			if err != nil {
				return argErr(i, "[]byte", err.String())
			}
			*v = buf
		case *string:
			*v = val
		case *bool:
			*v = val == "t"
		case *int:
			x, err := strconv.Atoi(val)
			if err != nil {
				return argErr(i, "int", err.String())
			}
			*v = x
		case *int64:
			x, err := strconv.Atoi64(val)
			if err != nil {
				return argErr(i, "int64", err.String())
			}
			*v = x
		case *float32:
			x, err := strconv.Atof32(val)
			if err != nil {
				return argErr(i, "float32", err.String())
			}
			*v = x
		case *float64:
			x, err := strconv.Atof64(val)
			if err != nil {
				return argErr(i, "float64", err.String())
			}
			*v = x
		case *time.Time:
			x, _, err := ParseTimestamp(val)
			if err != nil {
				return argErr(i, "time.Time", err.String())
			}
			*v = *x
		default:
			return os.NewError("unsupported type in Scan: " + reflect.TypeOf(v).String())
		}
	}
	return nil
}

// Clear frees the memory associated with the result.  Cleared results should
// not be subsequently used.
func (r *Result) Clear() {
	if r.res != nil {
		C.PQclear(r.res)
		r.res = nil
		runtime.SetFinalizer(r, nil)
	}
}

func buildCArgs(params ...interface{}) **C.char {
	sparams := make([]string, len(params))
	for i, v := range params {
		var str string
		switch v := v.(type) {
		case []byte:
			str = "\\x" + hex.EncodeToString(v)
		case bool:
			if v {
				str = "t"
			} else {
				str = "f"
			}
		case *time.Time:
			str = v.Format(timeFormat)
		default:
			str = fmt.Sprint(v)
		}

		sparams[i] = str
	}
	cparams := C.makeCharArray(C.int(len(sparams)))
	for i, s := range sparams {
		C.setArrayString(cparams, C.CString(s), C.int(i))
	}
	return cparams
}

type Conn struct {
	db      *C.PGconn
	stmtNum int
}

// Connect creates a new database connection using the given connection string.
// Each parameter setting is in the form 'keyword=value'.
// See http://www.postgresql.org/docs/9.0/static/libpq-connect.html#LIBPQ-PQCONNECTDBPARAMS
// for a list of recognized parameters.
func Connect(params string) (conn *Conn, err os.Error) {
	cparams := C.CString(params)
	defer C.free(unsafe.Pointer(cparams))
	db := C.PQconnectdb(cparams)
	if C.PQstatus(db) != C.CONNECTION_OK {
		err = connError(db)
		C.PQfinish(db)
		return nil, err
	}
	conn = &Conn{db, 0}
	runtime.SetFinalizer(conn, (*Conn).Close)
	return
}

func (c *Conn) exec(stmt string, params ...interface{}) (cres *C.PGresult) {
	stmtstr := C.CString(stmt)
	defer C.free(unsafe.Pointer(stmtstr))
	if len(params) == 0 {
		cres = C.PQexec(c.db, stmtstr)
	} else {
		cparams := buildCArgs(params...)
		defer C.freeCharArray(cparams, C.int(len(params)))
		cres = C.PQexecParams(c.db, stmtstr, C.int(len(params)), nil, cparams, nil, nil, 0)
	}
	return cres
}

// Exec executes the given SQL query with the given parameters.
func (c *Conn) Exec(cmd string, params ...interface{}) os.Error {
	cres := c.exec(cmd, params...)
	defer C.PQclear(cres)
	return resultError(cres)
}

// Query executes the given SQL query with the given parameters, returning a
// Result on successful execution.
func (c *Conn) Query(query string, params ...interface{}) (res *Result, err os.Error) {
	cres := c.exec(query, params...)
	if err = resultError(cres); err != nil {
		C.PQclear(cres)
		return
	}
	return newResult(cres), nil
}

// Prepare creates and returns a prepared statement with the given SQL statement.
func (c *Conn) Prepare(stmt string) (*Statement, os.Error) {
	// Generate unique statement name.
	stmtname := strconv.Itoa(c.stmtNum)
	stmtnamestr := C.CString(stmtname)
	c.stmtNum++
	defer C.free(unsafe.Pointer(stmtnamestr))
	stmtstr := C.CString(stmt)
	defer C.free(unsafe.Pointer(stmtstr))
	res := C.PQprepare(c.db, stmtnamestr, stmtstr, 0, nil)
	err := resultError(res)
	if err != nil {
		C.PQclear(res)
		return nil, err
	}
	statement := &Statement{stmtname, c.db, res}
	runtime.SetFinalizer(statement, (*Statement).Clear)
	return statement, nil
}

// Reset closes the connection to the server and attempts to re-establish a new
// connection using the parameters passed in the original Connect call.
func (c *Conn) Reset() os.Error {
	if c == nil || c.db == nil {
		return os.NewError("nil postgresql connection")
	}
	C.PQreset(c.db)
	if C.PQstatus(c.db) != C.CONNECTION_OK {
		return connError(c.db)
	}
	return nil
}

// Close closes the database connection and frees its associated memory.
func (c *Conn) Close() {
	if c != nil && c.db != nil {
		C.PQfinish(c.db)
		c.db = nil
		runtime.SetFinalizer(c, nil)
	}
}

type Statement struct {
	name string
	db   *C.PGconn
	res  *C.PGresult
}

func (s *Statement) exec(params ...interface{}) *C.PGresult {
	stmtName := C.CString(s.name)
	defer C.free(unsafe.Pointer(stmtName))
	cparams := buildCArgs(params...)
	defer C.freeCharArray(cparams, C.int(len(params)))
	return C.PQexecPrepared(s.db, stmtName, C.int(len(params)), cparams, nil, nil, 0)
}

// Exec executes the prepared statement with the given parameters.
func (s *Statement) Exec(params ...interface{}) os.Error {
	cres := s.exec(params...)
	defer C.PQclear(cres)
	return resultError(cres)
}

// Query executes the prepared statement with the given parameters, returning a
// Result on successful execution.
func (s *Statement) Query(params ...interface{}) (res *Result, err os.Error) {
	cres := s.exec(params...)
	if err = resultError(cres); err != nil {
		C.PQclear(cres)
		return
	}
	return newResult(cres), nil
}

// Clear frees the memory associated with the statement.  Cleared statements should
// not be subsequently used.
func (s *Statement) Clear() {
	if s != nil && s.res != nil {
		C.PQclear(s.res)
		runtime.SetFinalizer(s, nil)
	}
}
