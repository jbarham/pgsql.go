include $(GOROOT)/src/Make.inc

TARG=github.com/jbarham/pgsql.go

CGOFILES=\
	pgsql.go\
	pool.go\

CGO_CFLAGS=-I$(shell pg_config --includedir)
CGO_LDFLAGS=$(shell pg_config --libdir)/libpq.so

include $(GOROOT)/src/Make.pkg
