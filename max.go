package main

import (
	"log"
	"math"
	"math/rand"
	"reflect"
	"sync"
)

type XP struct {
	X []int
	P float64
}

func PrbKMax(spn SPN, k int) XP {
	return MaxXP(PrbK(spn, k))
}

func MaxXP(xps []XP) XP {
	max := XP{P: math.Inf(-1)}
	for _, xp := range xps {
		if max.P < xp.P {
			max = xp
		}
	}
	return max
}

func PrbK(spn SPN, k int) []XP {
	prt := partition(spn)
	res := make([]XP, k)
	wg := sync.WaitGroup{}
	for times := 0; times < k; times++ {
		wg.Add(1)
		go func(i int) {
			x := prb1(spn, prt)
			p := spn.EvalX(x)
			res[i] = XP{x, p}
			wg.Done()
		}(times)
	}
	wg.Wait()
	return res
}

func partition(spn SPN) []float64 {
	ass := make([][]float64, len(spn.Schema))
	for i := range ass {
		ass[i] = make([]float64, spn.Schema[i])
		for j := range ass[i] {
			ass[i][j] = 1
		}
	}
	return spn.Eval(ass)
}

func prb1(spn SPN, prt []float64) []int {
	x := make([]int, len(spn.Schema))
	reach := make([]bool, len(spn.Nodes))
	reach[len(spn.Nodes)-1] = true
	for i := len(spn.Nodes) - 1; i >= 0; i-- {
		if reach[i] {
			switch n := spn.Nodes[i].(type) {
			case *Trm:
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

func MaxMax(spn SPN) []int {
	prt := make([]float64, len(spn.Nodes))
	branch := make([]int, len(spn.Nodes))
	for i, n := range spn.Nodes {
		switch n := n.(type) {
		case *Trm:
			prt[i] = 0
		case *Sum:
			eBest, pBest := -1, math.Inf(-1)
			for _, e := range n.Edges {
				crt := e.Weight + prt[e.Node.ID()]
				if pBest < crt {
					pBest = crt
					eBest = e.Node.ID()
				}
			}
			branch[i] = eBest
			prt[i] = pBest
		case *Prd:
			val := 0.0
			for _, e := range n.Edges {
				val += prt[e.Node.ID()]
			}
			prt[i] = val
		}
	}

	x := make([]int, len(spn.Schema))
	reach := make([]bool, len(spn.Nodes))
	reach[len(spn.Nodes)-1] = true
	for i := len(spn.Nodes) - 1; i >= 0; i-- {
		if reach[i] {
			switch n := spn.Nodes[i].(type) {
			case *Trm:
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

func SumMax(spn SPN) []int {
	prt := partition(spn)
	x := make([]int, len(spn.Schema))
	reach := make([]bool, len(spn.Nodes))
	reach[len(spn.Nodes)-1] = true
	for i := len(spn.Nodes) - 1; i >= 0; i-- {
		if reach[i] {
			switch n := spn.Nodes[i].(type) {
			case *Trm:
				x[n.Kth] = n.Value
			case *Sum:
				eBest, pBest := -1, math.Inf(-1)
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

func NaiveBayes(spn SPN) []int {
	xs := make([]int, len(spn.Schema))
	for i := range xs {
		sBest, pBest := -1, math.Inf(-1)
		for j := 0; j < spn.Schema[i]; j++ {
			ps := spn.Eval(marginalAss1(spn.Schema, i, j))
			p := ps[len(ps)-1]
			if pBest < p {
				pBest = p
				sBest = j
			}
		}
		xs[i] = sBest
	}
	return xs
}

func marginalAss1(schema []int, kth int, val int) [][]float64 {
	ass := make([][]float64, len(schema))
	for i := range ass {
		ass[i] = make([]float64, schema[i])
		for j := range ass[i] {
			if i == kth {
				if j == val {
					ass[i][j] = 1
				} else {
					ass[i][j] = 0
				}
			} else {
				ass[i][j] = 1
			}
		}
	}
	return ass
}

func BeamSearch(spn SPN, xps []XP, beamSize int) XP {
	best := XP{P: math.Inf(-1)}
	for i := 0; len(xps) > 0; i++ {
		log.Printf("[BeamSearch] Round %d, len: %d, pBest %f\n", i, len(xps), best.P)
		xps = uniqueX(xps)
		xps = topK(xps, beamSize)
		xp1 := topK(xps, 1)
		if best.P < xp1[0].P {
			best = xp1[0]
		}
		xps = nextGens(xps, spn)
	}
	return best
}

func nextGens(xps []XP, spn SPN) []XP {
	res := []XP{}
	resChan := make([]chan []XP, len(xps))
	for i, xp := range xps {
		ch := make(chan []XP)
		go nextGenP(xp, spn, ch)
		resChan[i] = ch
	}
	for _, ch := range resChan {
		res = append(res, <-ch...)
	}
	return res
}

func nextGen(xp XP, spn SPN, ch chan []XP) {
	res := []XP{}
	for i, cnt := range spn.Schema {
		for xi := 0; xi < cnt; xi++ {
			if xp.X[i] != xi {
				nx := make([]int, len(xp.X))
				copy(nx, xp.X)
				nx[i] = xi
				np := spn.EvalX(nx)
				if np > xp.P {
					res = append(res, XP{nx, np})
				}
			}
		}
	}
	ch <- res
}

func nextGenP(xp XP, spn SPN, ch chan []XP) {
	res := []XP{}
	chs := make([]chan []XP, len(spn.Schema))
	for i, cnt := range spn.Schema {
		chi := make(chan []XP)
		chs[i] = chi
		go genKth(cnt, xp, i, spn, chi)
	}
	for i := range chs {
		res = append(res, <-chs[i]...)
	}
	ch <- res
}

func genKth(cnt int, xp XP, i int, spn SPN, chi chan []XP) {
	r := []XP{}
	for xi := 0; xi < cnt; xi++ {
		if xp.X[i] != xi {
			nx := make([]int, len(xp.X))
			copy(nx, xp.X)
			nx[i] = xi
			np := spn.EvalX(nx)
			if np > xp.P {
				//res = append(res, XP{nx, np})
				r = append(r, XP{nx, np})
			}
		}
	}
	chi <- r
}

func uniqueX(xps []XP) []XP {
	is := make([]bool, len(xps))
	res := make([]XP, 0, len(xps))
	for i, xpi := range xps {
		is[i] = true
		for j, xpj := range xps[:i] {
			if is[j] && reflect.DeepEqual(xpi.X, xpj.X) {
				is[i] = false
				break
			}
		}
		if is[i] {
			res = append(res, xpi)
		}
	}
	return res
}

func topK(xps []XP, k int) []XP {
	if k > len(xps) {
		k = len(xps)
	}
	for i := 0; i < k; i++ {
		for j := i + 1; j < len(xps); j++ {
			if xps[i].P < xps[j].P {
				xps[i], xps[j] = xps[j], xps[i]
			}
		}
	}
	return xps[:k]
}

func Max(spn SPN) float64 {
	val := make([]float64, len(spn.Nodes))
	for i, n := range spn.Nodes {
		switch n := n.(type) {
		case *Trm:
			val[i] = 0
		case *Sum:
			max := math.Inf(-1)
			for _, e := range n.Edges {
				max = math.Max(max, val[e.Node.ID()]+e.Weight)
			}
			val[i] = max
		case *Prd:
			prd := 0.0
			for _, e := range n.Edges {
				prd += val[e.Node.ID()]
			}
			val[i] = prd
		}
	}
	return val[len(val)-1]
}

func Derivative(spn SPN, xs []int) []float64 {
	pr := spn.Eval(X2Ass(xs, spn.Schema))
	dr := make([]float64, len(spn.Nodes))
	for i := range dr {
		dr[i] = math.Inf(-1)
	}
	dr[len(dr)-1] = 0.0
	for i := len(spn.Nodes) - 1; i >= 0; i-- {
		switch n := spn.Nodes[i].(type) {
		case *Sum:
			for _, e := range n.Edges {
				dr[e.Node.ID()] = logSumExp(dr[e.Node.ID()], dr[i]+e.Weight)
			}
		case *Prd:
			for j, e := range n.Edges {
				other := 0.0
				for k, e := range n.Edges {
					if j != k {
						other += pr[e.Node.ID()]
					}
				}
				dr[e.Node.ID()] = logSumExp(dr[e.Node.ID()], dr[i]+other)
			}
		}
	}
	return dr
}
