package debug

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

// Run will launch the go debugger on another port
func Run() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}