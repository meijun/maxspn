package main

import (
	"math"
	"math/rand"
)

type X []int

func X2Assign(x X, stateCnt int) Assign {
	n := len(x)
	as := make(Assign, n)
	for i, xi := range x {
		as[i] = make([]float64, stateCnt)
		as[i][xi] = 1
	}
	return as
}

func PrbK(spn SPN, k int) (X, float64) {
	prt := partition(spn)

	xBest, pBest := X{}, math.Inf(-1)
	for times := 0; times < k; times++ {
		x := prb1(spn, prt)
		p := spn.Pr(X2Assign(x, 2))
		if p > pBest {
			pBest = p
			xBest = x
		}
	}
	return xBest, pBest
}

func partition(spn SPN) []float64 {
	prt := make([]float64, len(spn))
	for i, n := range spn {
		switch n := n.(type) {
		case *Trm:
			prt[i] = 0
		case *Sum:
			prt[i] = logSumExpF(len(n.Edges), func(k int) float64 {
				return n.Edges[k].Weight + prt[n.Edges[k].Node.ID()]
			})
		case *Prd:
			val := 0.0
			for _, e := range n.Edges {
				val += prt[e.Node.ID()]
			}
			prt[i] = val
		}
	}
	return prt
}

func prb1(spn SPN, prt []float64) X {
	x := X{}
	reach := make([]bool, len(spn))
	reach[len(spn)-1] = true
	for i := len(spn) - 1; i >= 0; i-- {
		if reach[i] {
			switch n := spn[i].(type) {
			case *Trm:
				for len(x) <= n.Kth {
					x = append(x, 0)
				}
				x[n.Kth] = n.Value
			case *Sum:
				r := math.Log(rand.Float64()) + prt[i]
				crt := math.Inf(-1)
				for _, e := range n.Edges {
					crt = logSumExp(crt, e.Weight+prt[e.Node.ID()])
					if r < crt {
						reach[e.Node.ID()] = true
						break
					}
				}
			case *Prd:
				for _, e := range n.Edges {
					reach[e.Node.ID()] = true
				}
			}
		}
	}
	return x
}

func logSumExp(as ...float64) float64 {
	return logSumExpF(len(as), func(i int) float64 {
		return as[i]
	})
}

func MaxMax(spn SPN) X {
	prt := make([]float64, len(spn))
	branch := make([]SID, len(spn))
	for i, n := range spn {
		switch n := n.(type) {
		case *Trm:
			prt[i] = 0
		case *Sum:
			eBest, pBest := SID(-1), math.Inf(-1)
			for _, e := range n.Edges {
				crt := e.Weight + prt[e.Node.ID()]
				if pBest < crt {
					pBest = crt
					eBest = e.Node.ID()
				}
			}
			branch[i] = eBest
		case *Prd:
			val := 0.0
			for _, e := range n.Edges {
				val += prt[e.Node.ID()]
			}
			prt[i] = val
		}
	}

	x := X{}
	reach := make([]bool, len(spn))
	reach[len(spn)-1] = true
	for i := len(spn) - 1; i >= 0; i-- {
		if reach[i] {
			switch n := spn[i].(type) {
			case *Trm:
				for len(x) <= n.Kth {
					x = append(x, 0)
				}
				x[n.Kth] = n.Value
			case *Sum:
				reach[branch[i]] = true
			case *Prd:
				for _, e := range n.Edges {
					reach[e.Node.ID()] = true
				}
			}
		}
	}
	return x
}

func SumMax(spn SPN) X {
	prt := partition(spn)
	x := X{}
	reach := make([]bool, len(spn))
	reach[len(spn)-1] = true
	for i := len(spn) - 1; i >= 0; i-- {
		if reach[i] {
			switch n := spn[i].(type) {
			case *Trm:
				for len(x) <= n.Kth {
					x = append(x, 0)
				}
				x[n.Kth] = n.Value
			case *Sum:
				eBest, pBest := SID(-1), math.Inf(-1)
				for _, e := range n.Edges {
					crt := e.Weight + prt[e.Node.ID()]
					if pBest < crt {
						pBest = crt
						eBest = e.Node.ID()
					}
				}
				reach[eBest] = true
			case *Prd:
				for _, e := range n.Edges {
					reach[e.Node.ID()] = true
				}
			}
		}
	}
	return x
}

func NaiveBayes(spn SPN, stateCnt int) X {
	varCnt := 0
	for _, n := range spn {
		if n, ok := n.(*Trm); ok {
			if n.Kth > varCnt {
				varCnt = n.Kth
			}
		}
	}
	varCnt++
	getAssign := func(k int, s int) Assign {
		as := make(Assign, varCnt)
		for i := range as {
			as[i] = make([]float64, stateCnt)
			for j := range as[i] {
				if i == k {
					if j == s {
						as[i][j] = 1
					} else {
						as[i][j] = 0
					}
				} else {
					as[i][j] = 1
				}
			}
		}
		return as
	}
	x := make(X, varCnt)
	for i := 0; i < varCnt; i++ {
		sBest, pBest := -1, math.Inf(-1)
		for j := 0; j < stateCnt; j++ {
			p := spn.Pr(getAssign(i, j))
			if pBest < p {
				pBest = p
				sBest = j
			}
		}
		x[i] = sBest
	}
	return x
}
