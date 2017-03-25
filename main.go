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
	//defer profile.Start().Stop()
	f := func(dataSet string) {
		//Exp(dataSet, MaxMaxMethod, "MM")
		//Exp(dataSet, MaxMaxBSMethod, "MM_BS")
		//Exp(dataSet, SumMaxMethod, "SM")
		//Exp(dataSet, SumMaxBSMethod, "SM_BS")
		//Exp(dataSet, Prb1kMethod, "P1k")
		//Exp(dataSet, Prb1kBSMethod, "P1kBS")
		//Exp(dataSet, Max, "Max")
		//Exp(dataSet, ExactSolver, "Exact")
		Exp(dataSet, MCMethod, "MC")
	}
	f(LR_SPN)
	//f(ID_AC_SPN)
}
