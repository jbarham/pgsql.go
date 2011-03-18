// Program to exercise pgsql.Pool.
package main

import (
	"fmt"
	"log"
	"time"
	"strconv"
	"runtime"
	"github.com/jbarham/pgsql.go"
)

func testPool(done chan bool, name string, p *pgsql.Pool) {
	for i := 0; i < 5; i++ {
		conn, err := p.Acquire()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Goroutine %s acquired connection\n", name)
		// Run a trivial query.
		res, err := conn.Query("SELECT 1+1")
		if err != nil {
			log.Printf("Query error: %s\n", err)
		} else {
			if res.Next() {
				var x int
				if err := res.Scan(&x); err != nil {
					log.Printf("Scan error: %s\n", err)
				} else if x != 2 {
					log.Printf("Result is %d, expected 2\n", x)
				}
			} else {
				log.Printf("Next failed")
			}
		}
		time.Sleep(1e9)
		p.Release(conn)
		fmt.Printf("Goroutine %s released connection\n", name)
	}
	fmt.Printf("Goroutine %s done\n", name)
	done <- true
}

func do(n int) {
	pool, err := pgsql.NewPool("dbname=testdb", n-1, 5)
	if err != nil {
		log.Fatal(err)
	}
	pool.Debug = true
	done := make(chan bool)
	for i := 0; i < n; i++ {
		go testPool(done, strconv.Itoa(i), pool)
		time.Sleep(1e9)
	}
	for i := 0; i < n; i++ {
		<-done
	}
	fmt.Println("do done")
	// pool should be garbage collected after function exit.
}

func main() {
	do(3)
	time.Sleep(7 * 1e9) // Idle connections should be closed.
	runtime.GC()        // Force garbage collection.
	time.Sleep(5 * 1e9) // Pool should be garbage collected.
}
