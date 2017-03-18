package main

import (
	"log"
	"math/rand"
	"os"
)

func init() {
	rand.Seed(0)
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Lshortfile | log.Ltime)
}

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
		Exp(dataSet, ExactOrderDerMethod, "Exact")
	}
	f(LR_SPN)
	//f(ID_AC_SPN)
}
