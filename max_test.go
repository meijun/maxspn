package main

import (
	"math"
	"math/rand"
	"testing"
)

func TestPrbK(t *testing.T) {
	spn := AC2SPN(LoadAC("data/idspac/nltcs.ac"))
	t.Log(PrbKMax(spn, 1))
	t.Log(PrbKMax(spn, 1))
	t.Log(PrbKMax(spn, 1))
	t.Log(PrbKMax(spn, 1))
	t.Log(PrbKMax(spn, 1))
	t.Log(PrbKMax(spn, 10))
	t.Log(PrbKMax(spn, 100))
}

func TestMaxMax(t *testing.T) {
	spn := AC2SPN(LoadAC("data/idspac/nltcs.ac"))
	x := MaxMax(spn)
	t.Log(x, spn.EvalX(x))
}

func TestSumMax(t *testing.T) {
	spn := AC2SPN(LoadAC("data/idspac/nltcs.ac"))
	x := SumMax(spn)
	t.Log(x, spn.EvalX(x))
}

func TestNaiveBayes(t *testing.T) {
	spn := AC2SPN(LoadAC("data/idspac/nltcs.ac"))
	x := NaiveBayes(spn)
	t.Log(x, spn.EvalX(x))
}

func TestBeamSearch(t *testing.T) {
	spn := AC2SPN(LoadAC("data/idspac/nltcs.ac"))
	t.Log(BeamSearch(spn, PrbK(spn, 100), 16))
}

func TestMax(t *testing.T) {
	spn := AC2SPN(LoadAC("data/idspac/nltcs.ac"))
	t.Log(Max(spn))
	t.Log(partition(spn)[len(spn.Nodes)-1])
}

func TestDerivative(t *testing.T) {
	spn := AC2SPN(LoadAC("data/idspac/nltcs.ac"))
	x := make([]int, len(spn.Schema))
	dr := DerivativeX(spn, x)
	for i := range spn.Nodes {
		if n, ok := spn.Nodes[i].(*Trm); ok {
			nx := make([]int, len(x))
			copy(nx, x)
			nx[n.Kth] = n.Value
			np := spn.EvalX(nx)
			//t.Logf("%d %d: %f %f %f\n", n.Kth, n.Value, dr[i], np, math.Abs(dr[i]-np))
			if math.Abs(dr[i]-np) > 1e-6 {
				t.Errorf("%d %d: %f %f %f\n", n.Kth, n.Value, dr[i], np, math.Abs(dr[i]-np))
			}
		}
	}
}

func TestDerivativeS(t *testing.T) {
	spn := LoadSPN(LR_SPN + "nltcs")
	times := 10
	for kth := 0; kth < times; kth++ {
		as := make([][]float64, len(spn.Schema))
		for i := range as {
			as[i] = make([]float64, 2)
			as[i][0] = float64(rand.Intn(2))
			as[i][1] = float64(rand.Intn(2))
		}
		d := Derivative(spn, as)
		ds := DerivativeS(spn, as)
		//t.Log(d)
		//t.Log(ds)
		for i := range d {
			if math.Abs(d[i]-ds[i]) > 1e-6 {
				t.Fail()
			}
		}
	}
}

func TestDerivativeSS(t *testing.T) {
	spn := LoadSPN(LR_SPN + "nltcs")
	times := 10
	for kth := 0; kth < times; kth++ {
		as := make([][]float64, len(spn.Schema))
		for i := range as {
			as[i] = make([]float64, 2)
			as[i][0] = float64(rand.Intn(2))
			as[i][1] = float64(rand.Intn(2))
		}
		d := Derivative(spn, as)
		ds := DerivativeSS(spn, as)
		//t.Log(d)
		//t.Log(ds)
		for i := range d {
			if math.Abs(d[i]-ds[i]) > 1e-6 {
				t.Fail()
			}
		}
	}
}

func TestTopKMaxMax(t *testing.T) {
	spn := LoadSPN(TY_SPN + "4")
	//spn := LoadSPN(LR_SPN + "dna")
	xs := TopKMaxMax(spn, 100)
	xm := MaxMax(spn)
	t.Log(" ", spn.EvalX(xm), xm)
	for i, x := range xs {
		t.Log(i, spn.EvalX(x), x)
	}
}
