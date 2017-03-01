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
	Exp(LR_SPN, Prb1kBSMethod)
}
