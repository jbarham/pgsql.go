pgsql.go
========

pgsql.go is a high-level wrapper for the [PostgreSQL database](http://www.postgresql.org/)
[libpq](http://www.postgresql.org/docs/9.0/static/libpq.html) client library.

Requirements
------------

The pgsql.go package `Makefile` assumes that the `libpq` library is installed at
`/usr/local/pgsql/lib/libpq.so` and that the `libpq` headers are under
`/usr/local/pgsql/include`.  Modify the `Makefile` to suit your setup.

To use the compiled pgsql.go package, the `libpq.so` library directory should be
in your library path (e.g., `$LD_LIBRARY_PATH` on Linux).

Installation
------------

	cd $GOROOT/src/pkg
	git clone git://github.com/jbarham/pgsql.go.git github.com/jbarham/pgsql.go
	cd github.com/jbarham/pgsql.go
	make install
	
The package self-test `pgsql_test.go` assumes that the user running the test
can connect to a running PostgreSQL server on localhost with an existing `testdb`
database.  See the [PostgreSQL installation documentation](http://www.postgresql.org/docs/9.0/static/admin.html)
([short version](http://www.postgresql.org/docs/9.0/static/install-short.html))
for how to create a test database, or adjust the test connection parameters
to suit your setup.

Usage
-----

	import "github.com/jbarham/pgsql.go"

See the package test file [pgsql_test.go](https://github.com/jbarham/pgsql.go/blob/master/pgsql_test.go)
for example usage.

About
-----

pgsql.go was written by John E. Barham (jbarham@gmail.com).

The API is inspired by the [gosqlite](http://code.google.com/p/gosqlite/) package
by Russ Cox.
