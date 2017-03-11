package main

import (
	"math"
)

func ExactOrder(spn SPN, baseline float64) float64 {
	x := make([]int, len(spn.Schema))
	for i := range x {
		x[i] = -1
	}
	return dfs2(spn, x, baseline)
}
func dfs2(spn SPN, x []int, baseline float64) float64 {
	x2 := make([]int, len(x))
	copy(x2, x)
	x = x2
	for i := range x {
		if x[i] == -1 {
			x[i] = 0
			xi0 := eval2(spn, x)
			x[i] = 1
			xi1 := eval2(spn, x)
			if xi0 < baseline && xi1 < baseline {
				return baseline
			}
			if xi0 < baseline {
				x[i] = 1
			}
			if xi1 < baseline {
				x[i] = 0
			}
		}
	}
	for i := range x {
		if x[i] == -1 {
			x[i] = 0
			baseline = math.Max(dfs2(spn, x, baseline), baseline)
			x[i] = 1
			baseline = math.Max(dfs2(spn, x, baseline), baseline)
			return baseline
		}
	}
	return spn.EvalX(x)
}

func eval2(spn SPN, x []int) float64 {
	a := make([][]float64, len(spn.Schema))
	for i := range a {
		a[i] = make([]float64, 2)
		if x[i] == -1 {
			a[i][0] = 1
			a[i][1] = 1
		} else {
			a[i][x[i]] = 1
		}
	}
	return spn.Eval(a)[len(spn.Nodes)-1]
}

func Exact(spn SPN, baseline float64) float64 {
	x := make([]int, len(spn.Schema))
	return dfs(spn, x, 0, baseline)
}

func dfs(spn SPN, x []int, xi int, baseline float64) float64 {
	if xi == len(spn.Schema) {
		return math.Max(baseline, eval(spn, x, xi))
	}
	x[xi] = 0
	if eval(spn, x, xi+1) > baseline {
		baseline = math.Max(baseline, dfs(spn, x, xi+1, baseline))
	}
	x[xi] = 1
	if eval(spn, x, xi+1) > baseline {
		baseline = math.Max(baseline, dfs(spn, x, xi+1, baseline))
	}
	return baseline
}

func eval(spn SPN, x []int, xi int) float64 {
	a := make([][]float64, len(spn.Schema))
	for i := range a {
		a[i] = make([]float64, 2)
		if i < xi {
			a[i][x[i]] = 1
		} else {
			a[i][0] = 1
			a[i][1] = 1
		}
	}
	return spn.Eval(a)[len(spn.Nodes)-1]
}
