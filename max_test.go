package main

import "testing"

func TestPrbK(t *testing.T) {
	spn := AC2SPN(LoadAC("data/nltcs.ac"))
	t.Log(PrbK(spn, 1))
	t.Log(PrbK(spn, 1))
	t.Log(PrbK(spn, 1))
	t.Log(PrbK(spn, 1))
	t.Log(PrbK(spn, 1))
	t.Log(PrbK(spn, 10))
	t.Log(PrbK(spn, 100))
}

func TestMaxMax(t *testing.T) {
	spn := AC2SPN(LoadAC("data/nltcs.ac"))
	x := MaxMax(spn)
	t.Log(x, spn.Pr(X2Assign(x, 2)))
}

func TestSumMax(t *testing.T) {
	spn := AC2SPN(LoadAC("data/nltcs.ac"))
	x := SumMax(spn)
	t.Log(x, spn.Pr(X2Assign(x, 2)))
}

func TestNaiveBayes(t *testing.T) {
	spn := AC2SPN(LoadAC("data/nltcs.ac"))
	x := NaiveBayes(spn, 2)
	t.Log(x, spn.Pr(X2Assign(x, 2)))
}

func TestBeamSearch(t *testing.T) {
	spn := AC2SPN(LoadAC("data/nltcs.ac"))
	varCnt := 16
	schema := make(X, varCnt)
	for i := 0; i < varCnt; i++ {
		schema[i] = 2
	}
	t.Log(BeamSearch(spn, PrbKInit(spn, 100), schema, 16))
}
