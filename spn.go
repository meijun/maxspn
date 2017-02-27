package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"math"
	"strconv"
)

type Node interface {
	ID() int
	SetID(id int)
}

type SPN struct {
	Nodes  []Node
	Schema []int
}

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
	id    int
}
type Sum struct {
	Edges []SumEdge
	id    int
}
type Prd struct {
	Edges []PrdEdge
	id    int
}

func (t *Trm) ID() int      { return t.id }
func (t *Trm) SetID(id int) { t.id = id }
func (s *Sum) ID() int      { return s.id }
func (s *Sum) SetID(id int) { s.id = id }
func (p *Prd) ID() int      { return p.id }
func (p *Prd) SetID(id int) { p.id = id }

// By ID-SPN
func AC2SPN(ac AC) SPN {
	nn := len(ac.Nodes)
	ns := make([]Node, nn)
	we := make([]float64, nn)
	zs := make([]bool, nn) // zero scope
	for i, n := range ac.Nodes {
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
				if _, ok := ac.Nodes[ci].(AddNode); !ok {
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
				if _, ok := ac.Nodes[ci].(NumNode); !ok {
					sum.Edges = append(sum.Edges, SumEdge{we[ci], ns[ci]})
				}
			}
			ns[i] = sum
			zs[i] = false
		}
	}
	if _, ok := ac.Nodes[nn-1].(AddNode); !ok {
		ns = append(ns, &Sum{Edges: []SumEdge{{we[nn-1], ns[nn-1]}}})
	}
	nodes := make([]Node, 0, nn)
	for _, n := range ns {
		if n != nil {
			n.SetID(int(len(nodes)))
			nodes = append(nodes, n)
		}
	}
	return SPN{nodes, ac.Schema}
}

// Return log(Eval(x))
func (spn SPN) EvalX(xs []int) float64 {
	val := spn.Eval(X2Ass(xs, spn.Schema))
	return val[len(val)-1]
}

func (spn SPN) Eval(ass [][]float64) []float64 {
	val := make([]float64, len(spn.Nodes))
	for _, n := range spn.Nodes {
		switch n := n.(type) {
		case *Trm:
			val[n.ID()] = math.Log(ass[n.Kth][n.Value])
		case *Sum:
			val[n.ID()] = logSumExpF(len(n.Edges), func(k int) float64 {
				return n.Edges[k].Weight + val[n.Edges[k].Node.ID()]
			})
		case *Prd:
			prd := 0.0
			for _, e := range n.Edges {
				prd += val[e.Node.ID()]
			}
			val[n.ID()] = prd
		}
	}
	return val
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
	for _, n := range spn.Nodes {
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

func (spn SPN) SaveAsAC(filename string) {
	data := make([]byte, 0, len(spn.Nodes))
	data = formatSchema(data, spn.Schema)
	id := make([]int, len(spn.Nodes))
	crt := 0
	for i, n := range spn.Nodes {
		switch n := n.(type) {
		case *Trm:
			data = append(data, 'v')
			data = append(data, ' ')
			data = strconv.AppendInt(data, int64(n.Kth), 10)
			data = append(data, ' ')
			data = strconv.AppendInt(data, int64(n.Value), 10)
			data = append(data, '\n')
			id[i] = crt
			crt++
		case *Sum:
			sid := make([]int, len(n.Edges))
			for j, e := range n.Edges {
				// new number node
				data = append(data, 'n')
				data = append(data, ' ')
				data = strconv.AppendFloat(data, math.Exp(e.Weight), 'f', -1, 64)
				data = append(data, '\n')
				crt++
				// new mul node
				data = append(data, '*')
				data = append(data, ' ')
				data = strconv.AppendInt(data, int64(crt)-1, 10)
				data = append(data, ' ')
				data = strconv.AppendInt(data, int64(id[e.Node.ID()]), 10)
				data = append(data, '\n')
				sid[j] = crt
				crt++
			}
			data = append(data, '+')
			for j := range n.Edges {
				data = append(data, ' ')
				data = strconv.AppendInt(data, int64(sid[j]), 10)
			}
			data = append(data, '\n')
			id[i] = crt
			crt++
		case *Prd:
			data = append(data, '*')
			for _, e := range n.Edges {
				data = append(data, ' ')
				data = strconv.AppendInt(data, int64(id[e.Node.ID()]), 10)
			}
			data = append(data, '\n')
			id[i] = crt
			crt++
		}
	}
	data = append(data, []byte("EOF\n")...)
	ioutil.WriteFile(filename, data, 0666)
}

func (spn SPN) Save(filename string) {
	data := make([]byte, 0, len(spn.Nodes))
	data = formatSchema(data, spn.Schema)
	for _, n := range spn.Nodes {
		switch n := n.(type) {
		case *Trm:
			data = append(data, 'v')
			data = append(data, ' ')
			data = strconv.AppendInt(data, int64(n.Kth), 10)
			data = append(data, ' ')
			data = strconv.AppendInt(data, int64(n.Value), 10)
			data = append(data, '\n')
		case *Sum:
			data = append(data, '+')
			for _, e := range n.Edges {
				data = append(data, ' ')
				data = strconv.AppendInt(data, int64(e.Node.ID()), 10)
				data = append(data, ' ')
				data = strconv.AppendFloat(data, e.Weight, 'f', -1, 64)
			}
			data = append(data, '\n')
		case *Prd:
			data = append(data, '*')
			for _, e := range n.Edges {
				data = append(data, ' ')
				data = strconv.AppendInt(data, int64(e.Node.ID()), 10)
			}
			data = append(data, '\n')
		}
	}
	data = append(data, []byte("EOF\n")...)
	ioutil.WriteFile(filename, data, 0666)
}

func LoadSPN(filename string) SPN {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	bss := bytes.Split(bs, []byte("\n"))
	end := len(bss) - 1
	for string(bss[end]) != "EOF" {
		end--
	}

	scs := bytes.Split(bss[0][1:len(bss[0])-1], []byte(" "))
	schema := make([]int, len(scs))
	for i, s := range scs {
		schema[i] = parseInt(string(s))
	}

	bss = bss[1:end]
	nodes := make([]Node, len(bss))
	for i, ln := range bss {
		bs := bytes.Split(ln, []byte(" "))
		switch string(bs[0]) {
		case "v":
			nodes[i] = &Trm{
				Kth:   parseInt(string(bs[1])),
				Value: parseInt(string(bs[2])),
				id:    i,
			}
		case "+":
			es := make([]SumEdge, len(bs)/2)
			for j := 1; j < len(bs); j += 2 {
				node := nodes[parseInt(string(bs[j]))]
				weight := parseFloat(string(bs[j+1]))
				es[j/2] = SumEdge{
					Weight: weight,
					Node:   node,
				}
			}
			nodes[i] = &Sum{
				Edges: es,
				id:    i,
			}
		case "*":
			es := make([]PrdEdge, len(bs)-1)
			for j, v := range bs[1:] {
				es[j] = PrdEdge{
					Node: nodes[parseInt(string(v))],
				}
			}
			nodes[i] = &Prd{
				Edges: es,
				id:    i,
			}
		}
	}
	return SPN{nodes, schema}
}

func formatSchema(data []byte, schema []int) []byte {
	data = append(data, '(')
	for i := 0; i < len(schema); i++ {
		if i != 0 {
			data = append(data, ' ')
		}
		data = strconv.AppendInt(data, int64(schema[i]), 10)
	}
	data = append(data, ')', '\n')
	return data
}
