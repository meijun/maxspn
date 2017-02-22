package main

import (
	"io/ioutil"
	"log"
	"math"
	"os/exec"
)

var DataNames = []string{
	"nltcs", "msnbc", "kdd", "plants", "baudio", "bnetflix", "jester", "accidents", "tretail", "pumsb_star",
	"dna", "kosarek", "msweb", "book", "tmovie", "cwebkb", "cr52", "c20ng", "bbc", "ad",
}
var VarCnt = []int{
	16, 17, 64, 69, 100, 100, 100, 111, 135, 163,
	180, 190, 294, 500, 500, 839, 889, 910, 1058, 1556,
}

func SumMaxBS() {
	res := make([]float64, len(DataNames))
	for i, name := range DataNames {
		spn := AC2SPN(LoadAC("data/" + name + ".ac"))
		x := SumMax(spn)
		xp := BeamSearch(spn, []XP{{x, spn.EvalX(x)}}, 31)
		log.Println(i, name, xp)
		res[i] = xp.P
	}
	log.Println(res)
}

func BSearch() {
	res := make([]float64, len(DataNames))
	for i, name := range DataNames {
		log.Println(i, name)
		spn := AC2SPN(LoadAC("data/" + name + ".ac"))
		xp := BeamSearch(spn, PrbK(spn, 1000), 31)
		log.Println(xp)
		res[i] = xp.P
	}
	log.Println(res)
}

func LibraMPE() {
	res := make([]float64, len(DataNames))
	for i, name := range DataNames {
		res[i] = libraMPE1(name, VarCnt[i])
		log.Println(i, res[i])
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
	AC2SPN(LoadAC(name + ".ac")).SaveAsAC(name + ".ac2")
	cmd := exec.Command("libra", "acquery", "-m", name+".ac", "-ev", name+".ev", "-mpe")
	res, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
		return math.Inf(-1)
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
	return AC2SPN(LoadAC(name + ".ac")).EvalX(x)
}

func Exp() {
	pk := func(spn SPN) float64 {
		return PrbKMax(spn, 100).P
	}
	mm := func(spn SPN) float64 {
		return spn.EvalX(MaxMax(spn))
	}
	sm := func(spn SPN) float64 {
		return spn.EvalX(SumMax(spn))
	}
	nb := func(spn SPN) float64 {
		return spn.EvalX(NaiveBayes(spn))
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
