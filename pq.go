package pq

/*
#include <stdlib.h>
#include <libpq-fe.h>
*/
import "C"

import (
	"os"
	"unsafe"
	"strconv"
)

type Error string

func (e Error) String() string {
	return "libpq error: " + string(e)
}

func connError(db *C.PGconn) os.Error {
	return Error(C.GoString(C.PQerrorMessage(db)))
}

type Conn struct {
	db      *C.PGconn
	stmtNum int
}

func Connect(params string) (conn *Conn, err os.Error) {
	cparams := C.CString(params)
	defer C.free(unsafe.Pointer(cparams))
	db := C.PQconnectdb(cparams)
	if C.PQstatus(db) != C.CONNECTION_OK {
		err = connError(db)
		C.PQfinish(db)
		return nil, err
	}
	return &Conn{db, 0}, nil
}

type Stmt struct {
	name string
	res  *C.PGresult
}

func (s *Stmt) Clear() {
	if s != nil && s.res != nil {
		C.PQclear(s.res)
	}
}

func (s *Stmt) Error() os.Error {
	return resultError(s.res)
}

type Result struct {
	res                   *C.PGresult
	nrows, ncols, currRow int
	cols                  []string
}

func newResult(res *C.PGresult) *Result {
	ncols := int(C.PQnfields(res))
	cols := make([]string, 0, ncols)
	for i := 0; i < ncols; i++ {
		cols = append(cols, C.GoString(C.PQfname(res, C.int(i))))
	}
	return &Result{res, int(C.PQntuples(res)), ncols, -1, cols}
}

func (r *Result) Next() bool {
	if r.currRow+1 < r.nrows {
		r.currRow++
		return true
	}
	return false
}

func (r *Result) Row() (vals []string, err os.Error) {
	vals = make([]string, 0, r.ncols)
	for i := 0; i < r.ncols; i++ {
		vals = append(vals, C.GoString(C.PQgetvalue(r.res, C.int(r.currRow), C.int(i))))
	}
	return
}

func (r *Result) RowMap() (vals map[string]string, err os.Error) {
	row, err := r.Row()
	if err != nil {
		return
	}
	vals = make(map[string]string, len(row))
	for i, col := range r.cols {
		vals[col] = row[i]
	}
	return
}

func (r *Result) Clear() {
	if r.res != nil {
		C.PQclear(r.res)
		r.res = nil
	}
}

func resultError(res *C.PGresult) os.Error {
	serr := C.GoString(C.PQresultErrorMessage(res))
	if serr == "" {
		return nil
	}
	return os.NewError("result error: " + serr)
}

func (c *Conn) Exec(cmd string) os.Error {
	cmdstr := C.CString(cmd)
	defer C.free(unsafe.Pointer(cmdstr))
	cres := C.PQexec(c.db, cmdstr)
	defer C.PQclear(cres)
	return resultError(cres)
}

func (c *Conn) Query(cmd string) (res *Result, err os.Error) {
	cmdstr := C.CString(cmd)
	defer C.free(unsafe.Pointer(cmdstr))
	cres := C.PQexec(c.db, cmdstr)
	if err = resultError(cres); err != nil {
		C.PQclear(cres)
		return
	}
	return newResult(cres), nil
}

func (c *Conn) Prepare(query string) (*Stmt, os.Error) {
	// Generate unique statement name.
	stmtname := strconv.Itoa(c.stmtNum)
	stmtnamestr := C.CString(stmtname)
	c.stmtNum++
	defer C.free(unsafe.Pointer(stmtnamestr))
	querystr := C.CString(query)
	defer C.free(unsafe.Pointer(querystr))
	res := C.PQprepare(c.db, stmtnamestr, querystr, 0, nil)
	err := resultError(res)
	if err != nil {
		C.PQclear(res)
		return nil, err
	}
	return &Stmt{stmtname, res}, nil
}

func (c *Conn) Options() string {
	return C.GoString(C.PQoptions(c.db))
}

func (c *Conn) error() os.Error {
	return connError(c.db)
}

func (c *Conn) Reset() os.Error {
	if c == nil || c.db == nil {
		return os.NewError("nil postgresql connection")
	}
	C.PQreset(c.db)
	if C.PQstatus(c.db) != C.CONNECTION_OK {
		err := connError(c.db)
		c.Close()
		return err
	}
	return nil
}

func (c *Conn) Close() {
	if c != nil && c.db != nil {
		C.PQfinish(c.db)
		c.db = nil
	}
}

func (c *Conn) Finish() { c.Close() }
