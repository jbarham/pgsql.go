include $(GOROOT)/src/Make.inc

TARG=github.com/jbarham/pgsql.go

CGOFILES=\
	pgsql.go\
	pool.go\

CGO_CFLAGS=-I/usr/local/pgsql/include
CGO_LDFLAGS=/usr/local/pgsql/lib/libpq.so

include $(GOROOT)/src/Make.pkg
