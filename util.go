package main

import (
	"fmt"
	"os"
	"os/signal"
)

func fatal(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}

func fatalf(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}

func await(a ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, a...)
	<-c
	signal.Stop(c)
}
