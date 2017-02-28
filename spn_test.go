package main

import (
	"math"
	"math/rand"
	"testing"
)

func TestAC2SPN(t *testing.T) {
	ac := LoadAC("data/idspac/nltcs.ac")
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
	//for _, name := range DataNames {
	//	t.Log(AC2SPN(LoadAC("data/"+name+".ac")).Info())
	//}
	t.Log(AC2SPN(LoadAC("data/idspac/nltcs.ac")).Info())
}

func TestSPN_SaveAsAC(t *testing.T) {
	AC2SPN(LoadAC("data/idspac/nltcs.ac")).SaveAsAC("data/nltcs.ac2")
}

func TestLoadSPN(t *testing.T) {
	spn := LoadSPN("data/spn.spn")
	t.Log(spn.Info())
	spn.Save("data/spn2.spn")
}
