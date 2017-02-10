package main

import (
	"io/ioutil"
	"math"
	"strconv"
)

type Node interface {
	ID() SID
	SetID(id SID)
}

type SPN []Node

type SID int

type SumEdge struct {
	Weight float64 // in log
	Node   Node
}
type PrdEdge struct {
	Node Node
}

type Trm struct {
	Kth   int // k-th variable
	Value int // variable state
	id    SID
}
type Sum struct {
	Edges []SumEdge
	id    SID
}
type Prd struct {
	Edges []PrdEdge
	id    SID
}

func (t *Trm) ID() SID      { return t.id }
func (t *Trm) SetID(id SID) { t.id = id }
func (s *Sum) ID() SID      { return s.id }
func (s *Sum) SetID(id SID) { s.id = id }
func (p *Prd) ID() SID      { return p.id }
func (p *Prd) SetID(id SID) { p.id = id }

// By ID-SPN
func AC2SPN(ac AC) SPN {
	nn := len(ac)
	ns := make([]Node, nn)
	we := make([]float64, nn)
	zs := make([]bool, nn) // zero scope
	for i, n := range ac {
		switch n := n.(type) {
		case VarNode:
			ns[i] = &Trm{
				Kth:   n.Kth,
				Value: n.Value,
			}
			we[i] = 0
			zs[i] = false
		case NumNode:
			we[i] = math.Log(float64(n))
			zs[i] = true
		case MulNode:
			prd := &Prd{Edges: make([]PrdEdge, 0, len(n))}
			w := 0.0
			for _, ci := range n {
				if !zs[ci] {
					// if _, ok := ac[ci].(NumNode); !ok {
					prd.Edges = append(prd.Edges, PrdEdge{ns[ci]})
				}
				if _, ok := ac[ci].(AddNode); !ok {
					w += we[ci]
				}
			}
			we[i] = w
			zs[i] = len(prd.Edges) == 0
			if !zs[i] {
				ns[i] = prd
			}
		case AddNode:
			sum := &Sum{Edges: make([]SumEdge, 0, len(n))}
			for _, ci := range n {
				if _, ok := ac[ci].(NumNode); !ok {
					sum.Edges = append(sum.Edges, SumEdge{we[ci], ns[ci]})
				}
			}
			ns[i] = sum
			zs[i] = false
		}
	}
	if _, ok := ac[nn-1].(AddNode); !ok {
		ns = append(ns, &Sum{Edges: []SumEdge{{we[nn-1], ns[nn-1]}}})
	}
	spn := make(SPN, 0, nn)
	for _, n := range ns {
		if n != nil {
			n.SetID(SID(len(spn)))
			spn = append(spn, n)
		}
	}
	return spn
}

// Return log(Pr(x))
func (spn SPN) Pr(assign Assign) float64 {
	pr := make([]float64, len(spn))
	for _, n := range spn {
		switch n := n.(type) {
		case *Trm:
			pr[n.ID()] = math.Log(assign[n.Kth][n.Value])
		case *Sum:
			pr[n.ID()] = logSumExpF(len(n.Edges), func(k int) float64 {
				return n.Edges[k].Weight + pr[n.Edges[k].Node.ID()]
			})
		case *Prd:
			val := 0.0
			for _, e := range n.Edges {
				val += pr[e.Node.ID()]
			}
			pr[n.ID()] = val
		}
	}
	return pr[spn[len(spn)-1].ID()]
}

type Stat struct {
	Sum float64
	Avg float64
	Std float64
	Min float64
	Max float64
}

func (spn SPN) Info() (trmNode, sumNode, prdNode int, sumEdge, prdEdge Stat) {
	se := []float64{}
	pe := []float64{}
	for _, n := range spn {
		switch n := n.(type) {
		case *Trm:
			trmNode++
		case *Sum:
			sumNode++
			se = append(se, float64(len(n.Edges)))
		case *Prd:
			prdNode++
			pe = append(pe, float64(len(n.Edges)))
		}
	}
	sumEdge = analyse(se)
	prdEdge = analyse(pe)
	return
}
func analyse(xs []float64) Stat {
	SumX := 0.0
	SumX2 := 0.0
	min := math.Inf(1)
	max := math.Inf(-1)
	for _, x := range xs {
		SumX += x
		SumX2 += x * x
		min = math.Min(min, x)
		max = math.Max(max, x)
	}
	EX := SumX / float64(len(xs))
	EX2 := SumX2 / float64(len(xs))
	Std := math.Sqrt(EX2 - EX*EX)
	return Stat{
		Sum: SumX,
		Avg: EX,
		Std: Std,
		Min: min,
		Max: max,
	}
}

func (spn SPN) SaveAsAC(name string) {
	ac := []byte{}
	id := make([]int, len(spn))
	crt := 0
	sc := []int{}
	for i, n := range spn {
		switch n := n.(type) {
		case *Trm:
			for len(sc) <= n.Kth {
				sc = append(sc, 0)
			}
			if sc[n.Kth] < n.Value {
				sc[n.Kth] = n.Value
			}
			ac = append(ac, 'v')
			ac = append(ac, ' ')
			ac = strconv.AppendInt(ac, int64(n.Kth), 10)
			ac = append(ac, ' ')
			ac = strconv.AppendInt(ac, int64(n.Value), 10)
			ac = append(ac, '\n')
			id[i] = crt
			crt++
		case *Sum:
			sid := make([]int, len(n.Edges))
			for j, e := range n.Edges {
				// new number node
				ac = append(ac, 'n')
				ac = append(ac, ' ')
				ac = strconv.AppendFloat(ac, math.Exp(e.Weight), 'f', -1, 64)
				ac = append(ac, '\n')
				crt++
				// new mul node
				ac = append(ac, '*')
				ac = append(ac, ' ')
				ac = strconv.AppendInt(ac, int64(crt)-1, 10)
				ac = append(ac, ' ')
				ac = strconv.AppendInt(ac, int64(id[e.Node.ID()]), 10)
				ac = append(ac, '\n')
				sid[j] = crt
				crt++
			}
			ac = append(ac, '+')
			for j := range n.Edges {
				ac = append(ac, ' ')
				ac = strconv.AppendInt(ac, int64(sid[j]), 10)
			}
			ac = append(ac, '\n')
			id[i] = crt
			crt++
		case *Prd:
			ac = append(ac, '*')
			for _, e := range n.Edges {
				ac = append(ac, ' ')
				ac = strconv.AppendInt(ac, int64(id[e.Node.ID()]), 10)
			}
			ac = append(ac, '\n')
			id[i] = crt
			crt++
		}
	}
	bs := make([]byte, 0, len(ac))
	for _, n := range sc {
		bs = append(bs, ' ')
		bs = strconv.AppendInt(bs, int64(n)+1, 10)
	}
	bs[0] = '('
	bs = append(bs, ')', '\n')
	bs = append(bs, ac...)
	bs = append(bs, []byte("EOF\n")...)
	ioutil.WriteFile(name, bs, 0666)
}
