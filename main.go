package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
)

func init() {
	rand.Seed(0)
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Lshortfile | log.Ltime)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		<-c
		for _, f := range finally {
			f()
		}
		os.Exit(0)
	}()
}

var finally []func()

func main() {

}
