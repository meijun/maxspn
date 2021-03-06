package main

import (
	"math"
	"math/rand"
	"os/exec"
	"testing"
)

func TestAC2SPN(t *testing.T) {
	ac := LoadAC(ID_AC + "nltcs")
	AC2SPN(ac)
}

func TestSPN_Pr(t *testing.T) {
	ac := LoadAC("data/idspac/nltcs.ac")
	spn := AC2SPN(ac)
	p0 := spn.EvalX([]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	if math.Abs(p0-(-1.750023)) > 1e-6 {
		t.Fail() // from `libra acquery -m nltcs.ac -q nltcs0.q`
	}
	p1 := spn.EvalX([]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1})
	if math.Abs(p1-(-3.519511)) > 1e-6 {
		t.Fail() // from `libra acquery -m nltcs.ac -q nltcs1.q`
	}
	p01 := spn.EvalX([]int{0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1})
	if math.Abs(p01-(-15.535905)) > 1e-6 {
		t.Fail() // from nltcs01.q
	}
	p10 := spn.EvalX([]int{1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0})
	if math.Abs(p10-(-27.309816)) > 1e-6 {
		t.Fail() // from nltcs10.q
	}
	rand.Seed(0)
	for times := 0; times < 100; times++ {
		x := make([]int, 16)
		for i := range x {
			x[i] = rand.Intn(2)
		}
		pa := ac.EvalX(x)
		ps := spn.EvalX(x)
		if math.Abs(pa-ps) > 1e-6 {
			t.Fail()
		}
	}
}

func TestSPN_Info(t *testing.T) {
	//for _, name := range DATA_NAMES {
	//	t.Log(AC2SPN(LoadAC("data/"+name+".ac")).Info())
	//}
	t.Log(LoadSPN(ID_AC_SPN + "nltcs").Info())
}

func TestSPN_SaveAsAC(t *testing.T) {
	AC2SPN(LoadAC("data/idspac/nltcs.ac")).SaveAsAC("data/nltcs.ac2")
}

func TestLoadSPN(t *testing.T) {
	spn := LoadSPN("data/spn.spn")
	t.Log(spn.Info())
	spn.Save("data/spn2.spn")
}

func TestSPN_Plot(t *testing.T) {
	pr := func(name string) {
		spn := LoadSPN(name)
		spn.Plot(name+".dot", 0)
		exec.Command("dot", "-Tpng", "-O", name+".dot").Run()
	}
	pr(TY_SPN + "0")
	pr(TY_SPN + "1")
	pr(TY_SPN + "2")
	pr(TY_SPN + "3")
}

func TestSPN_QuerySPN(t *testing.T) {
	spn := LoadSPN(LR_SPN + "nltcs")
	qSPN := spn.QuerySPN([]byte("????????????????"))
	for times := 0; times < 100; times++ {
		as := make([][]float64, len(spn.Schema))
		for i := range as {
			as[i] = make([]float64, 2)
			as[i][0] = 1 // float64(rand.Intn(2))
			as[i][1] = float64(rand.Intn(2))
		}
		se := spn.Eval(as)
		qe := qSPN.Eval(as)
		for i := range se {
			if math.Abs(se[i]-qe[i]) > 1e-6 {
				t.Fail()
			}
		}
	}
}

func TestSPN_QuerySPN2(t *testing.T) {
	spn := LoadSPN(LR_SPN + "nltcs")
	q := []byte("??????????******")
	qSPN := spn.QuerySPN(q)
	for times := 0; times < 100; times++ {
		as := make([][]float64, len(spn.Schema))
		for i := range as {
			as[i] = make([]float64, 2)
			as[i][0] = float64(rand.Intn(2))
			as[i][1] = 1 // float64(rand.Intn(2))
			if q[i] == '*' {
				as[i][0] = 1
				as[i][1] = 1
			}
		}
		se := spn.Eval(as)
		qe := qSPN.Eval(as)
		if math.Abs(se[len(se)-1]-qe[len(qe)-1]) > 1e-6 {
			t.Fail()
		}
	}
}
