package main

import (
	"math"
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
	dr := Derivative(spn, x)
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
