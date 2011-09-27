include $(GOROOT)/src/Make.inc

TARG=github.com/jbarham/pgsql.go

CGOFILES=\
	pgsql.go\
	pool.go\

CGO_CFLAGS=-I`pg_config --includedir`
CGO_LDFLAGS=-L`pg_config --libdir` -lpq

include $(GOROOT)/src/Make.pkg
