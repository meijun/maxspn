package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
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
func (spn SPN) EvalX(x []int) float64 {
	val := spn.Eval(X2Ass(x, spn.Schema))
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

func (spn SPN) Plot(filename string, treeDepth int) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Open file error: %v\n", err)
	}
	defer file.Close()
	dep := make([]int, len(spn.Nodes))
	for i, n := range spn.Nodes {
		switch n := n.(type) {
		case *Trm:
			dep[i] = 0
		case *Sum:
			dep[i] = len(spn.Nodes)
			for _, e := range n.Edges {
				if dep[i] > dep[e.Node.ID()] {
					dep[i] = dep[e.Node.ID()]
				}
			}
			dep[i]++
		case *Prd:
			dep[i] = len(spn.Nodes)
			for _, e := range n.Edges {
				if dep[i] > dep[e.Node.ID()] {
					dep[i] = dep[e.Node.ID()]
				}
			}
			dep[i]++
		}
	}
	fmt.Fprintln(file, "graph{")
	fmt.Fprintln(file, `rankdir="LR"`)
	plotDFS(spn.Nodes[len(spn.Nodes)-1], "n", map[int]struct{}{}, dep, treeDepth, file)
	fmt.Fprintln(file, "}")
}

func plotDFS(n Node, from string, vis map[int]struct{}, dep []int, treeDepth int, file *os.File) string {
	var nid string
	if dep[n.ID()] > treeDepth { // DAG
		nid = fmt.Sprintf("n%d", n.ID())
		if _, ok := vis[n.ID()]; ok {
			return nid
		}
		vis[n.ID()] = struct{}{}
	} else { // Tree
		nid = fmt.Sprintf("%sn%d", from, n.ID())
	}
	var color string
	switch n := n.(type) {
	case *Trm:
		if n.Value == 0 {
			color = "red"
		} else {
			color = "green"
		}
	case *Sum:
		color = "blue"
		for _, e := range n.Edges {
			cid := plotDFS(e.Node, nid, vis, dep, treeDepth, file)
			fmt.Fprintf(file, "%s--%s[color=grey]\n", nid, cid)
		}
	case *Prd:
		color = "grey"
		for _, e := range n.Edges {
			cid := plotDFS(e.Node, nid, vis, dep, treeDepth, file)
			fmt.Fprintf(file, "%s--%s[color=grey]\n", nid, cid)
		}
	}
	fmt.Fprintf(file, "%s[color=%s][shape=point]\n", nid, color)
	return nid
}

func (spn SPN) QuerySPN(q []byte) SPN {
	idMap := map[int]int{}
	varCnt := 0
	schema := make([]int, 0, len(spn.Schema))
	for i, c := range q {
		if c == '?' {
			idMap[i] = varCnt
			varCnt++
			schema = append(schema, spn.Schema[i])
		}
	}
	if varCnt == 0 {
		log.Println(string(q))
		log.Fatal("QuerySPN no '?'")
	}
	nn := len(spn.Nodes)
	ns := make([]Node, nn+1)
	we := make([]float64, nn)
	for i, n := range spn.Nodes {
		switch n := n.(type) {
		case *Trm:
			if q[n.Kth] == '?' {
				ns[i] = &Trm{Kth: idMap[n.Kth], Value: n.Value}
			} else {
				w := math.Inf(-1)
				if n.Value == 0 && (q[n.Kth] == '0' || q[n.Kth] == '*') {
					w = 0
				}
				if n.Value == 1 && (q[n.Kth] == '1' || q[n.Kth] == '*') {
					w = 0
				}
				we[i] = w
			}
		case *Sum:
			if ns[n.Edges[0].Node.ID()] == nil {
				we[i] = logSumExpF(len(n.Edges), func(k int) float64 {
					return n.Edges[k].Weight + we[n.Edges[k].Node.ID()]
				})
			} else {
				es := make([]SumEdge, len(n.Edges))
				for j, e := range n.Edges {
					es[j] = SumEdge{e.Weight + we[e.Node.ID()], ns[e.Node.ID()]}
				}
				ns[i] = &Sum{Edges: es}
			}
		case *Prd:
			es := make([]PrdEdge, 0, len(n.Edges))
			w := 0.0
			for _, e := range n.Edges {
				w += we[e.Node.ID()]
				if ns[e.Node.ID()] != nil {
					es = append(es, PrdEdge{ns[e.Node.ID()]})
				}
			}
			we[i] = w
			if len(es) > 0 {
				ns[i] = &Prd{Edges: es}
			}
		}
	}
	if _, ok := ns[nn-1].(*Sum); !ok {
		ns[nn] = &Sum{Edges: []SumEdge{{we[nn-1], ns[nn-1]}}}
	}
	nodes := make([]Node, 0, nn+1)
	for _, n := range ns {
		if n != nil {
			n.SetID(len(nodes))
			nodes = append(nodes, n)
		}
	}
	return SPN{nodes, schema}
}

func (spn SPN) StageSPN(q []int) SPN {
	idMap := map[int]int{}
	varCnt := 0
	schema := make([]int, 0, len(spn.Schema))
	for i, c := range q {
		if c == -1 {
			idMap[i] = varCnt
			varCnt++
			schema = append(schema, spn.Schema[i])
		}
	}
	if varCnt == 0 {
		log.Println(q)
		log.Fatal("q has no -1")
	}
	nn := len(spn.Nodes)
	ns := make([]Node, nn+1)
	we := make([]float64, nn)
	for i, n := range spn.Nodes {
		switch n := n.(type) {
		case *Trm:
			if q[n.Kth] == -1 {
				ns[i] = &Trm{Kth: idMap[n.Kth], Value: n.Value}
			} else {
				w := math.Inf(-1)
				if n.Value == 0 && q[n.Kth] == 0 {
					w = 0
				}
				if n.Value == 1 && q[n.Kth] == 1 {
					w = 0
				}
				we[i] = w
			}
		case *Sum:
			if ns[n.Edges[0].Node.ID()] == nil {
				we[i] = logSumExpF(len(n.Edges), func(k int) float64 {
					return n.Edges[k].Weight + we[n.Edges[k].Node.ID()]
				})
			} else {
				es := make([]SumEdge, len(n.Edges))
				for j, e := range n.Edges {
					es[j] = SumEdge{e.Weight + we[e.Node.ID()], ns[e.Node.ID()]}
				}
				ns[i] = &Sum{Edges: es}
			}
		case *Prd:
			es := make([]PrdEdge, 0, len(n.Edges))
			w := 0.0
			for _, e := range n.Edges {
				w += we[e.Node.ID()]
				if ns[e.Node.ID()] != nil {
					es = append(es, PrdEdge{ns[e.Node.ID()]})
				}
			}
			we[i] = w
			if len(es) > 0 {
				ns[i] = &Prd{Edges: es}
			}
		}
	}
	if _, ok := ns[nn-1].(*Sum); !ok {
		ns[nn] = &Sum{Edges: []SumEdge{{we[nn-1], ns[nn-1]}}}
	}
	nodes := make([]Node, 0, nn+1)
	for _, n := range ns {
		if n != nil {
			n.SetID(len(nodes))
			nodes = append(nodes, n)
		}
	}
	return SPN{nodes, schema}
}

func (spn SPN) FastStageSPN(q []int) SPN {
	varCnt := 0
	for _, c := range q {
		if c == -1 {
			varCnt++
		}
	}
	if varCnt == 0 {
		log.Println(q)
		log.Fatal("q has no -1")
	}
	nn := len(spn.Nodes)
	ns := make([]Node, nn)
	we := make([]float64, nn)
	ch := make([]bool, nn)
	for i, n := range spn.Nodes {
		switch n := n.(type) {
		case *Trm:
			if q[n.Kth] == -1 {
				ns[i] = n
			} else {
				w := math.Inf(-1)
				if n.Value == 0 && q[n.Kth] == 0 {
					w = 0
				}
				if n.Value == 1 && q[n.Kth] == 1 {
					w = 0
				}
				we[i] = w
				ch[i] = true
			}
		case *Sum:
			chi := false
			for _, e := range n.Edges {
				if ch[e.Node.ID()] {
					chi = true
					break
				}
			}
			if chi {
				if ns[n.Edges[0].Node.ID()] == nil {
					we[i] = logSumExpF(len(n.Edges), func(k int) float64 {
						return n.Edges[k].Weight + we[n.Edges[k].Node.ID()]
					})
				} else {
					es := make([]SumEdge, len(n.Edges))
					for j, e := range n.Edges {
						es[j] = SumEdge{e.Weight + we[e.Node.ID()], ns[e.Node.ID()]}
					}
					ns[i] = &Sum{Edges: es, id: n.id}
				}
				ch[i] = true
			} else {
				ns[i] = n
			}
		case *Prd:
			chi := false
			for _, e := range n.Edges {
				if ch[e.Node.ID()] {
					chi = true
					break
				}
			}
			if chi {
				es := make([]PrdEdge, 0, len(n.Edges))
				w := 0.0
				for _, e := range n.Edges {
					w += we[e.Node.ID()]
					if ns[e.Node.ID()] != nil {
						es = append(es, PrdEdge{ns[e.Node.ID()]})
					}
				}
				we[i] = w
				if len(es) > 0 {
					ns[i] = &Prd{Edges: es, id: n.id}
				}
				ch[i] = true
			} else {
				ns[i] = n
			}
		}
	}
	if _, ok := ns[nn-1].(*Sum); !ok {
		//ns[nn] = &Sum{Edges: []SumEdge{{we[nn-1], ns[nn-1]}}}
		log.Println("Last node is not Sum")
	}
	return SPN{ns, spn.Schema}
}
