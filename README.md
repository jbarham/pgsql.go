pgsql.go
========

pgsql.go is a high-level wrapper for the [PostgreSQL database](http://www.postgresql.org/)
[libpq](http://www.postgresql.org/docs/9.0/static/libpq.html) client library.

Requirements
------------

The pgsql.go package `Makefile` assumes that `pg_config` is in your `$PATH`
to automatically determine the location of the PostgreSQL include directory and
the `libpq.so` shared library.

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

Connection Pools
----------------

Two goroutines cannot safely use the same database connection at the same time.
However it's often necessary for multiple goroutines to access a database
simultaneously, such as when creating a goroutine per HTTP request in a web server.
It's also relatively inefficient to create a new database connection for transient
goroutines.

For this type of situation, pgsql.go provides a connection pool type `Pool` which
allows for safe sharing of multiple connections between any number of goroutines
in the same process.

To create and use a connection pool, simply do the following (with error handling
omitted):

	// In main goroutine:
	// Create a connection pool with up to 3 connections.
	pool, _ := pgsql.NewPool("dbname=testdb", 3, pgsql.DEFAULT_IDLE_TIMEOUT)
	
	// In worker goroutine:
	conn, _ := pool.Acquire() // Get a connection from the pool.
	// Use the connection normally.
	result, _ := conn.Query("SELECT SUM(balance) FROM account")
	result.Next()
	// ...
	pool.Release(conn) // Release the connection back to the pool.

For a complete example demonstrating connection pool usage, see the file
[pool_example.go](https://github.com/jbarham/pgsql.go/blob/master/pool_example.go).
	
About
-----

pgsql.go was written by John E. Barham (jbarham@gmail.com).

The API is inspired by the [gosqlite](http://code.google.com/p/gosqlite/) package
by Russ Cox.
