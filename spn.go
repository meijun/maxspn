package main

import (
	"math"
)

type Node interface {
	ID() SID
	SetID(id SID)
}

type SPN []Node

type SID int

type SumEdge struct {
	Weight float64
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
	for i, n := range ac {
		switch n := n.(type) {
		case VarNode:
			ns[i] = &Trm{
				Kth:   n.Kth,
				Value: n.Value,
			}
			we[i] = 1
		case NumNode:
			we[i] = float64(n)
		case MulNode:
			prd := &Prd{}
			w := 1.0
			for _, ci := range n {
				if _, ok := ac[ci].(NumNode); !ok {
					prd.Edges = append(prd.Edges, PrdEdge{ns[ci]})
				}
				if _, ok := ac[ci].(AddNode); !ok {
					w *= we[ci]
				}
			}
			ns[i] = prd
			we[i] = w
		case AddNode:
			sum := &Sum{}
			for _, ci := range n {
				if _, ok := ac[ci].(NumNode); !ok {
					sum.Edges = append(sum.Edges, SumEdge{we[ci], ns[ci]})
				}
			}
			ns[i] = sum
		}
	}
	if _, ok := ac[nn-1].(AddNode); !ok {
		ns = append(ns, &Sum{Edges: []SumEdge{{we[nn-1], ns[nn-1]}}})
	}
	spn := SPN{}
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
			val := 0.0
			for _, e := range n.Edges {
				val += e.Weight * math.Exp(pr[e.Node.ID()])
			}
			pr[n.ID()] = math.Log(val)
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