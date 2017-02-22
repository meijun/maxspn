package main

import (
	"math"
	"reflect"
	"testing"
)

func testLoadAC_dfs(cnt []int, ac AC, v int, vis []bool) {
	if vis[v] {
		return
	}
	vis[v] = true
	switch nd := ac.Nodes[v].(type) {
	case VarNode:
		cnt[0]++
	case NumNode:
		cnt[1]++
	case MulNode:
		cnt[2]++
		for _, u := range nd {
			testLoadAC_dfs(cnt, ac, u, vis)
		}
		cnt[4] += len(nd)
	case AddNode:
		cnt[3]++
		for _, u := range nd {
			testLoadAC_dfs(cnt, ac, u, vis)
		}
		cnt[4] += len(nd)
	}
}

func TestLoadAC(t *testing.T) {
	ac := LoadAC("data/nltcs.ac")
	cnt := []int{0, 0, 0, 0, 0}
	testLoadAC_dfs(cnt, ac, len(ac.Nodes)-1, make([]bool, len(ac.Nodes)))
	if !reflect.DeepEqual(cnt, []int{32, 2625, 288155, 100776, 878825}) {
		t.Fail() // [var, const, times, plus, edges] from `libra fstats -i nltcs.ac`
	}
}

func TestAC_EvalX(t *testing.T) {
	ac := LoadAC("data/nltcs.ac")
	p0 := ac.EvalX([]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	if math.Abs(p0-(-1.750023)) > 1e-6 {
		t.Fail() // from `libra acquery -m nltcs.ac -q nltcs0.q`
	}
	p1 := ac.EvalX([]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1})
	if math.Abs(p1-(-3.519511)) > 1e-6 {
		t.Fail() // from `libra acquery -m nltcs.ac -q nltcs1.q`
	}
	p01 := ac.EvalX([]int{0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1})
	if math.Abs(p01-(-15.535905)) > 1e-6 {
		t.Fail() // from nltcs01.q
	}
	p10 := ac.EvalX([]int{1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0})
	if math.Abs(p10-(-27.309816)) > 1e-6 {
		t.Fail() // from nltcs10.q
	}
}

func TestAC_MaxMax(t *testing.T) {
	ac := LoadAC("data/nltcs.ac")
	x := ac.MaxMax()
	p := AC2SPN(ac).EvalX(x)
	t.Log(x, p)
}

func TestAC_Derivative(t *testing.T) {
	ac := LoadAC("data/nltcs.ac")
	x := make([]int, len(ac.Schema))
	dr := ac.Derivative(x)
	for i := range ac.Nodes {
		if n, ok := ac.Nodes[i].(VarNode); ok {
			nx := make([]int, len(x))
			copy(nx, x)
			nx[n.Kth] = n.Value
			np := ac.EvalX(nx)
			//t.Logf("%d %d: %f %f %f\n", n.Kth, n.Value, dr[i], np, math.Abs(dr[i]-np))
			if math.Abs(dr[i]-np) > 1e-6 {
				t.Errorf("%d %d: %f %f %f\n", n.Kth, n.Value, dr[i], np, math.Abs(dr[i]-np))
			}
		}
	}
}

func TestAC_Info(t *testing.T) {
	//for _, name := range DataNames {
	//	t.Log(LoadAC("data/" + name + ".ac").Info())
	//}
	t.Log(LoadAC("data/nltcs.ac").Info())
}
