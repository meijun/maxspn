package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
)

const (
	DATA_DIR = "data/"
	TMP      = "/tmp/"

	// the top one is source, the followings are generated files
	ID_AC        = DATA_DIR + "id-ac/"
	ID_AC_SPN    = DATA_DIR + "id-ac-spn/"
	ID_AC_SPN_AC = DATA_DIR + "id-ac-spn-ac/"

	LR_SPN    = DATA_DIR + "lr-spn/"
	LR_SPN_AC = DATA_DIR + "lr-spn-ac/"
)

var SPN_AC = map[string]string{
	ID_AC_SPN: ID_AC_SPN_AC,
	LR_SPN:    LR_SPN_AC,
}

var DATA_NAMES = []string{
	"nltcs", "msnbc", "kdd", "plants", "baudio", "bnetflix", "jester", "accidents", "tretail", "pumsb_star",
	"dna", "kosarek", "msweb", "book", "tmovie", "cwebkb", "cr52", "c20ng", "bbc", "ad",
}
var VAR_CNT = []int{
	16, 17, 64, 69, 100, 100, 100, 111, 135, 163,
	180, 190, 294, 500, 500, 839, 889, 910, 1058, 1556,
}
var DATA_VAR_CNT = map[string]int{}

func init() {
	for i, name := range DATA_NAMES {
		DATA_VAR_CNT[name] = VAR_CNT[i]
	}
}

// Given source, generate transformed data.
func PrepareData() {
	for _, name := range DATA_NAMES {
		// rename original ac file
		os.Rename(ID_AC+name+".ac", ID_AC+name)

		// id-ac => id-ac-spn => id-ac-spn-ac
		if _, err := os.Stat(ID_AC_SPN_AC + name); os.IsNotExist(err) {
			ac := LoadAC(ID_AC + name)
			spn := AC2SPN(ac)
			spn.Save(ID_AC_SPN + name)
			spn.SaveAsAC(ID_AC_SPN_AC + name)
		}

		// lr-spn => lr-spn-ac
		if _, err := os.Stat(LR_SPN_AC + name); os.IsNotExist(err) {
			spn := LoadSPN(LR_SPN + name)
			spn.SaveAsAC(LR_SPN_AC + name)
		}
	}
	log.Println("[DONE] PrepareData")
}

func Prb1kMethod(spn SPN) float64 {
	return PrbKMax(spn, 1000).P
}
func MaxMaxMethod(spn SPN) float64 {
	return spn.EvalX(MaxMax(spn))
}
func SumMaxMethod(spn SPN) float64 {
	return spn.EvalX(SumMax(spn))
}
func NaiveBayesMethod(spn SPN) float64 {
	return spn.EvalX(NaiveBayes(spn))
}

func Prb1kBSMethod(spn SPN) float64 {
	return BeamSearch(spn, PrbK(spn, 1000), 31).P
}
func MaxMaxBSMethod(spn SPN) float64 {
	x := MaxMax(spn)
	return BeamSearch(spn, []XP{{x, spn.EvalX(x)}}, 31).P
}
func SumMaxBSMethod(spn SPN) float64 {
	x := SumMax(spn)
	return BeamSearch(spn, []XP{{x, spn.EvalX(x)}}, 31).P
}

func Exp(dataSet string, method func(SPN) float64) {
	res := make([]float64, len(DATA_NAMES))
	for i, name := range DATA_NAMES {
		spn := LoadSPN(dataSet + name)
		res[i] = method(spn)
		log.Printf("[DONE] %s\n", DATA_NAMES[i])
	}
	log.Printf("[DONE] %s: %v\n", dataSet, res)
}

func LibraSPNMPE(dataSet string) {
	res := make([]float64, len(DATA_NAMES))
	for i, name := range DATA_NAMES {
		res[i] = libraSPNMPE1(dataSet, name)
	}
	log.Println("[libra]", res)
}

func printResult1(label string, res []float64) {
	fmt.Print(label)
	for _, r := range res {
		fmt.Print("\t")
		fmt.Print(r)
	}
	fmt.Println()
}

func libraSPNMPE1(dataSet, dataName string) float64 {
	x := libraMPE1(SPN_AC[dataSet], dataName)
	if x != nil {
		spn := LoadSPN(dataSet + dataName)
		return spn.EvalX(x)
	} else {
		return math.Inf(-1)
	}
}

func libraMPE1(dataSet, dataName string) []int {
	varCnt := DATA_VAR_CNT[dataName]
	star := make([]byte, varCnt*2)
	for i := 0; i < varCnt; i++ {
		star[i*2] = '*'
		star[i*2+1] = ','
	}
	star[varCnt*2-1] = '\n'

	evFilename := TMP + strconv.FormatUint(rand.Uint64(), 16)

	acFilename := dataSet + dataName
	err := ioutil.WriteFile(evFilename, star, 0666)
	cmd := exec.Command("libra", "acquery", "-m", acFilename, "-ev", evFilename, "-mpe")
	res, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Libra error when(%s, %s): %v\n", dataSet, dataName, err)
		return nil
	}
	log.Printf("[libra] output: %s\n", string(res))
	x := make([]int, varCnt)
	for i := range x {
		if res[i*2] == '1' {
			x[i] = 1
		} else {
			x[i] = 0
		}
	}
	return x
}

func ACMaxMaxExp(dataSet string) {
	ps := make([]float64, len(DATA_NAMES))
	for i, name := range DATA_NAMES {
		ac := LoadAC(dataSet + name)
		x := ac.MaxMax()
		p := AC2SPN(ac).EvalX(x)
		log.Println(x, p)
		ps[i] = p
	}
	log.Println(ps)
}
