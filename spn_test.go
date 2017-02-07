package main

import (
	"math"
	"math/rand"
	"testing"
)

func TestAC2SPN(t *testing.T) {
	ac := LoadAC("data/nltcs.ac")
	AC2SPN(ac)
}

func TestSPN_Pr(t *testing.T) {
	ac := LoadAC("data/nltcs.ac")
	spn := AC2SPN(ac)
	getAssign := func(a []int) Assign {
		as := make(Assign, 16)
		for i := range as {
			as[i] = []float64{0, 0}
			as[i][a[i]] = 1
		}
		return as
	}
	p0 := spn.Pr(getAssign([]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	if math.Abs(p0-(-1.750023)) > 1e-6 {
		t.Fail() // from `libra acquery -m nltcs.ac -q nltcs0.q`
	}
	p1 := spn.Pr(getAssign([]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}))
	if math.Abs(p1-(-3.519511)) > 1e-6 {
		t.Fail() // from `libra acquery -m nltcs.ac -q nltcs1.q`
	}
	p01 := spn.Pr(getAssign([]int{0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1}))
	if math.Abs(p01-(-15.535905)) > 1e-6 {
		t.Fail() // from nltcs01.q
	}
	p10 := spn.Pr(getAssign([]int{1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0}))
	if math.Abs(p10-(-27.309816)) > 1e-6 {
		t.Fail() // from nltcs10.q
	}
	rand.Seed(0)
	for times := 0; times < 100; times++ {
		x := make([]int, 16)
		for i := range x {
			x[i] = rand.Intn(2)
		}
		pa := ac.Pr(getAssign(x))
		ps := spn.Pr(getAssign(x))
		if math.Abs(pa-ps) > 1e-6 {
			t.Fail()
		}
	}
}
