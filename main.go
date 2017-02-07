package main

import (
	"log"
)

var DataNames = []string{
	"accidents", "ad", "baudio", "bbc", "bnetflix",
	"book", "c20ng", "cr52", "cwebkb", "dna",
	"jester", "kdd", "kosarek", "msnbc", "msweb",
	"nltcs", "plants", "pumsb_star", "tmovie", "tretail",
}

func main() {
	pk := func(spn SPN) float64 {
		_, p := PrbK(spn, 100)
		return p
	}
	mm := func(spn SPN) float64 {
		return spn.Pr(X2Assign(MaxMax(spn), 2))
	}
	sm := func(spn SPN) float64 {
		return spn.Pr(X2Assign(SumMax(spn), 2))
	}
	nb := func(spn SPN) float64 {
		return spn.Pr(X2Assign(NaiveBayes(spn, 2), 2))
	}
	fs := []func(SPN) float64{pk, mm, sm, nb}
	ss := make([]SPN, len(DataNames))
	for i, n := range DataNames {
		ss[i] = AC2SPN(LoadAC("data/" + n + ".ac"))
	}
	for _, f := range fs {
		res := make([]float64, len(ss))
		for i, s := range ss {
			res[i] = f(s)
		}
		log.Println(res)
	}
}
