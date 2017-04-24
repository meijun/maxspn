package main

import (
	"flag"
	"log"
	"math"
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
	flag.Parse()
	FinalExperiment()
	return

	//spn := LoadSPN(LR_SPN + "nltcs")
	//q := make([]int, len(spn.Schema))
	//q[0] = -1
	//log.Println(spn.EvalX(q))
	//spn = spn.FastStageSPN(q)
	//log.Println(spn.EvalX(q))
	//return
	stage := func(spn SPN) float64 {
		x := make([]int, len(spn.Schema))
		for i := range x {
			x[i] = -1
		}
		//return ExactStage(spn, x, math.Inf(-1))
		return ExactFastStage(spn, x, math.Inf(-1), 0)
	}
	Exp(LR_SPN, stage, "Stage")
}
