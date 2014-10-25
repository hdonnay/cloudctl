package main

import (
	"fmt"
	"log"
	"os"
)

var l *log.Logger

func init() {
	l = log.New(os.Stderr, "DEBUG ", log.Lmicroseconds|log.Lshortfile)
}
func debug(v ...interface{}) {
	l.Output(2, fmt.Sprint(v...))
}
