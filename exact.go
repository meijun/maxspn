package main

import (
	"math"
	"sort"
)

func ExactOrder(spn SPN, baseline float64) float64 {
	x := make([]int, len(spn.Schema))
	for i := range x {
		x[i] = -1
	}
	return dfs2(spn, x, baseline)
}

func dfs2(spn SPN, x []int, baseline float64) float64 {
	//log.Println(len(spn.Schema), baseline)
	x2 := make([]int, len(x))
	copy(x2, x)
	x = x2
	for i := range x {
		if x[i] == -1 {
			x[i] = 0
			xi0 := eval2(spn, x)
			x[i] = 1
			xi1 := eval2(spn, x)
			x[i] = -1
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
	return math.Max(baseline, spn.EvalX(x))
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

func ExactOrderDer(spn SPN, baseline float64) float64 {
	x := make([]int, len(spn.Schema))
	for i := range x {
		x[i] = -1
	}
	return dfs2Der(spn, x, baseline)
}

func dfs2Der(spn SPN, x []int, baseline float64) float64 {
	x2 := make([]int, len(x))
	copy(x2, x)
	x = x2
	as := make([][]float64, len(x))
	for i := range as {
		as[i] = make([]float64, 2)
		if x[i] == 0 || x[i] == -1 {
			as[i][0] = 1
		}
		if x[i] == 1 || x[i] == -1 {
			as[i][1] = 1
		}
	}
	var d [][]float64
	for {
		updated := false
		d = derivativeOfAssignment(spn, as)
		for i := range x {
			if x[i] == -1 {
				xi0 := d[i][0]
				xi1 := d[i][1]

				if xi0 < baseline && xi1 < baseline {
					return baseline
				}
				if xi0 < baseline {
					x[i] = 1
					as[i][0] = 0
					updated = true
				}
				if xi1 < baseline {
					x[i] = 0
					as[i][1] = 0
					updated = true
				}
			}
		}
		if !updated {
			break
		}
	}
	maxVarID := -1
	maxValID := -1
	maxDer := math.Inf(-1)
	for i := range x {
		if x[i] == -1 {
			crtValID := 0
			crtDer := d[i][0]
			if d[i][0] < d[i][1] {
				crtValID = 1
				crtDer = d[i][1]
			}
			if maxVarID == -1 || maxDer < crtDer {
				maxVarID = i
				maxValID = crtValID
				maxDer = crtDer
			}
		}
	}
	if i := maxVarID; i != -1 {
		x[i] = maxValID
		baseline = math.Max(dfs2Der(spn, x, baseline), baseline)
		x[i] = maxValID ^ 1
		baseline = math.Max(dfs2Der(spn, x, baseline), baseline)
		return baseline
	}
	return math.Max(baseline, math.Max(d[0][0], d[0][1]))
}

func eval2Der(spn SPN, x []int) float64 {
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

func ExactSolver(spn SPN) float64 {
	as := make([][]float64, len(spn.Schema))
	for i := range as {
		as[i] = make([]float64, spn.Schema[i])
		for j := range as[i] {
			as[i][j] = 1
		}
	}
	as, d := forwardChecking(spn, math.Inf(-1), as)
	return searchMax(spn, math.Inf(-1), as, d)
}
func searchMax(spn SPN, best float64, as [][]float64, d [][]float64) float64 {
	if isCompleteAssignment(as) {
		p := maximum(d[0])
		if best < p {
			best = p
		}
		return best
	}
	varID, valIDs := order(as, d)
	for _, valID := range valIDs {
		as[varID] = make([]float64, spn.Schema[varID])
		as[varID][valID] = 1
		asNew, dNew := forwardChecking(spn, best, as)
		if maximum(asNew[0]) != 0 {
			best = searchMax(spn, best, asNew, dNew)
		}
	}
	return best
}

func order(as [][]float64, d [][]float64) (int, []int) {
	varID := 0
	varIDCnt := math.MaxInt64
	varIDD := math.Inf(-1)
	for i := range as {
		oneCnt := 0
		maxD := math.Inf(-1)
		for j := range as[i] {
			if as[i][j] == 1 {
				oneCnt++
				maxD = math.Max(maxD, d[i][j])
			}
		}
		if oneCnt > 1 && (varIDCnt > oneCnt || varIDCnt == oneCnt && varIDD < maxD) {
			varID = i
			varIDCnt = oneCnt
			varIDD = maxD
		}
	}
	ids := make([]int, varIDCnt)
	idc := 0
	for i := range as[varID] {
		if as[varID][i] == 1 {
			ids[idc] = i
			idc++
		}
	}
	sort.Slice(ids, func(i, j int) bool { return d[varID][i] > d[varID][j] })
	return varID, ids
}

func isCompleteAssignment(as [][]float64) bool {
	for i := range as {
		one := false
		for j := range as[i] {
			if as[i][j] == 1 {
				if one {
					return false
				} else {
					one = true
				}
			}
		}
	}
	return true
}

func maximum(fs []float64) float64 {
	r := math.Inf(-1)
	for _, f := range fs {
		if r < f {
			r = f
		}
	}
	return r
}

func forwardChecking(spn SPN, best float64, as [][]float64) ([][]float64, [][]float64) {
	as = cloneAssignment(as)
	for {
		d := derivativeOfAssignment(spn, as)
		changed := false
		for i := range as {
			for j := range as[i] {
				if as[i][j] != 0 {
					if best >= d[i][j] {
						as[i][j] = 0
						changed = true
					}
				}
			}
		}
		if !changed {
			return as, d
		}
	}
}

func cloneAssignment(as [][]float64) [][]float64 {
	bs := make([][]float64, len(as))
	for i := range bs {
		bs[i] = make([]float64, len(as[i]))
		copy(bs[i], as[i])
	}
	return bs
}
func derivativeOfAssignment(spn SPN, as [][]float64) [][]float64 {
	der := DerivativeS(spn, as)
	d := make([][]float64, len(spn.Schema))
	for i := range d {
		d[i] = make([]float64, spn.Schema[i])
		for j := range d[i] {
			d[i][j] = math.Inf(-1)
		}
	}
	for i, n := range spn.Nodes {
		if n, ok := n.(*Trm); ok {
			d[n.Kth][n.Value] = logSumExp(d[n.Kth][n.Value], der[i])
		}
	}
	return d
}

func ExactSolverBin(spn SPN) float64 {
	as := make([][]float64, len(spn.Schema))
	for i := range as {
		as[i] = make([]float64, spn.Schema[i])
		for j := range as[i] {
			as[i][j] = 1
		}
	}
	as, d := forwardCheckingBin(spn, math.Inf(-1), as)
	return searchMaxBin(spn, math.Inf(-1), as, d)
}

func searchMaxBin(spn SPN, best float64, as [][]float64, d [][]float64) float64 {
	if isCompleteAssignmentBin(as) {
		return math.Max(best, math.Max(d[0][0], d[0][1]))
	}
	varID, valIDs := orderBin(as, d)
	for _, valID := range valIDs {
		as[varID] = make([]float64, 2)
		as[varID][valID] = 1
		asNew, dNew := forwardCheckingBin(spn, best, as)
		if asNew[0][0] > 0 || asNew[0][1] > 0 {
			best = searchMaxBin(spn, best, asNew, dNew)
		}
	}
	return best
}

func orderBin(as [][]float64, d [][]float64) (int, []int) {
	varID := -1
	varIDD := math.Inf(-1)
	for i := range as {
		if as[i][0] > 0 && as[i][1] > 0 {
			maxD := math.Max(d[i][0], d[i][1])
			if varIDD < maxD {
				varID = i
				varIDD = maxD
			}
		}
	}
	var ids []int
	if varID != -1 {
		if d[varID][0] > d[varID][1] {
			ids = []int{0, 1}
		} else {
			ids = []int{1, 0}
		}
	}
	return varID, ids
}

func isCompleteAssignmentBin(as [][]float64) bool {
	for i := range as {
		if as[i][0] > 0 && as[i][1] > 0 {
			return false
		}
	}
	return true
}

func forwardCheckingBin(spn SPN, best float64, as [][]float64) ([][]float64, [][]float64) {
	as = cloneAssignment(as)
	for {
		d := derivativeOfAssignment(spn, as)
		changed := false
		for i := range as {
			if as[i][0] != 0 && best >= d[i][0] {
				as[i][0] = 0
				changed = true
			}
			if as[i][1] != 0 && best >= d[i][1] {
				as[i][1] = 0
				changed = true
			}
		}
		if !changed {
			return as, d
		}
	}
}
