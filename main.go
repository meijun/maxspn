package main

import (
	"log"
	"os"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Lshortfile | log.Ltime)
}

func main() {
	f := func(dataSet string) {
		Exp(dataSet, MaxMaxMethod, "MM")
		Exp(dataSet, MaxMaxBSMethod, "MM_BS")
		Exp(dataSet, SumMaxMethod, "SM")
		Exp(dataSet, SumMaxBSMethod, "SM_BS")
		Exp(dataSet, Prb1kMethod, "P1k")
		Exp(dataSet, Prb1kBSMethod, "P1kBS")
	}
	f(LR_SPN)
	f(ID_AC_SPN)
}
