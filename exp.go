package main

import (
	"io/ioutil"
	"log"
	"os/exec"
)

var DataNames = []string{
	"accidents", "ad", "baudio", "bbc", "bnetflix",
	"book", "c20ng", "cr52", "cwebkb", "dna",
	"jester", "kdd", "kosarek", "msnbc", "msweb",
	"nltcs", "plants", "pumsb_star", "tmovie", "tretail",
}
var VarCnt = []int{
	111, 1556, 100, 1058, 100,
	500, 910, 889, 839, 180,
	100, 64, 190, 17, 294,
	16, 69, 163, 500, 135,
}

func LibraMPE() {
	res := make([]float64, len(DataNames))
	for i, name := range DataNames {
		log.Println(i)
		res[i] = libraMPE1(name, VarCnt[i])
	}
	log.Println(res)
}
func libraMPE1(name string, varCnt int) float64 {
	star := make([]byte, varCnt*2)
	for i := 0; i < varCnt; i++ {
		star[i*2] = '*'
		star[i*2+1] = ','
	}
	star[varCnt*2-1] = '\n'

	name = "data/" + name
	ioutil.WriteFile(name+".ev", star, 0666)
	//cmd := exec.Command("pwd")
	//cmd := exec.Command("libra", "fstats", "-i", name+".ac")
	cmd := exec.Command("libra", "acquery", "-m", name+".ac", "-ev", name+".ev", "-mpe")
	res, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(res))
	x := make([]int, varCnt)
	for i := range x {
		if res[i*2] == '1' {
			x[i] = 1
		} else {
			x[i] = 0
		}
	}
	return AC2SPN(LoadAC(name + ".ac")).Pr(X2Assign(x, 2))
}

func Exp() {
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
			log.Println(i)
			res[i] = f(s)
		}
		log.Println(res)
	}
}
