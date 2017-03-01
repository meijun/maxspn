package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"strconv"
)

type ACNode interface{}

type AC struct {
	Nodes  []ACNode
	Schema []int
}

type VarNode struct {
	Kth   int
	Value int
}
type NumNode float64
type MulNode []int
type AddNode []int

func LoadAC(filename string) AC {
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
	nodes := make([]ACNode, len(bss))
	for i, ln := range bss {
		bs := bytes.Split(ln, []byte(" "))
		switch string(bs[0]) {
		case "v":
			nodes[i] = VarNode{
				Kth:   parseInt(string(bs[1])),
				Value: parseInt(string(bs[2])),
			}
		case "n":
			nodes[i] = NumNode(parseFloat(string(bs[1])))
		case "*":
			mn := make(MulNode, len(bs)-1)
			for j, v := range bs[1:] {
				mn[j] = parseInt(string(v))
			}
			nodes[i] = mn
		case "+":
			an := make(AddNode, len(bs)-1)
			for j, v := range bs[1:] {
				an[j] = parseInt(string(v))
			}
			nodes[i] = an
		}
	}
	return AC{nodes, schema}
}

// Return log value
func (ac AC) EvalX(xs []int) float64 {
	val := ac.Eval(X2Ass(xs, ac.Schema))
	return val[len(val)-1]
}

func (ac AC) Eval(ass [][]float64) []float64 {
	nn := len(ac.Nodes)
	val := make([]float64, nn)
	for i, n := range ac.Nodes {
		switch n := n.(type) {
		case VarNode:
			val[i] = math.Log(ass[n.Kth][n.Value])
		case NumNode:
			val[i] = math.Log(float64(n))
		case MulNode:
			mul := 0.0
			for _, ai := range n {
				mul += val[ai]
			}
			val[i] = mul
		case AddNode:
			val[i] = logSumExpF(len(n), func(k int) float64 {
				return val[n[k]]
			})
		}
	}
	return val
}

func (ac AC) Info() (varNode, numNode, mulNode, addNode int, mulEdge, addEdge Stat) {
	me := []float64{}
	ae := []float64{}
	sc := make([]*big.Int, len(ac.Nodes))
	for i, n := range ac.Nodes {
		switch n := n.(type) {
		case VarNode:
			varNode++
			sc[i] = big.NewInt(0)
			sc[i].Lsh(big.NewInt(1), uint(n.Kth))
			//sc[i] = 1 << uint(n.Kth)
		case NumNode:
			numNode++
			sc[i] = big.NewInt(0)
		case MulNode:
			mulNode++
			me = append(me, float64(len(n)))
			s := big.NewInt(0)
			//s := 0
			for _, ci := range n {
				if big.NewInt(0).And(s, sc[ci]).Cmp(big.NewInt(0)) != 0 {
					//if s & sc[ci] != 0 {
					log.Fatal("mul", i, s, ci, sc[ci])
				}
				s.Or(s, sc[ci])
				//s |= sc[ci]
			}
			sc[i] = s
		case AddNode:
			addNode++
			ae = append(ae, float64(len(n)))
			s := big.NewInt(-1)
			//s := -1
			for _, ci := range n {
				if s.Cmp(big.NewInt(-1)) == 0 {
					//if s == -1 {
					s = sc[ci]
				} else {
					if s.Cmp(sc[ci]) != 0 {
						//if s != sc[ci] {
						log.Fatal("add", i, s, ci, sc[ci])
					}
				}
			}
			if s.Cmp(big.NewInt(-1)) == 0 || s.Cmp(big.NewInt(0)) == 0 {
				//if s == -1 || s == 0 {
				log.Fatal("add", s)
			}
			sc[i] = s
		}
	}
	mulEdge = analyse(me)
	addEdge = analyse(ae)
	return
}

func (ac AC) MaxMax() []int {
	val := make([]float64, len(ac.Nodes))
	win := make([]int, len(ac.Nodes))
	for i, n := range ac.Nodes {
		switch n := n.(type) {
		case VarNode:
			val[i] = 0
		case NumNode:
			val[i] = math.Log(float64(n))
		case MulNode:
			mul := 0.0
			for _, c := range n {
				mul += val[c]
			}
			val[i] = mul
		case AddNode:
			max := math.Inf(-1)
			winC := -1
			for _, c := range n {
				if max <= val[c] {
					max = val[c]
					winC = c
				}
			}
			val[i] = max
			win[i] = winC
		}
	}
	xs := make([]int, len(ac.Schema))
	reach := make([]bool, len(ac.Nodes))
	reach[len(ac.Nodes)-1] = true
	for i := len(ac.Nodes) - 1; i >= 0; i-- {
		if reach[i] {
			switch n := ac.Nodes[i].(type) {
			case VarNode:
				xs[n.Kth] = n.Value
			case NumNode:
			case MulNode:
				for _, c := range n {
					reach[c] = true
				}
			case AddNode:
				reach[win[i]] = true
			}
		}
	}
	return xs
}

func (ac AC) Derivative(xs []int) []float64 {
	pr := ac.Eval(X2Ass(xs, ac.Schema))
	dr := make([]float64, len(ac.Nodes))
	for i := range dr {
		dr[i] = math.Inf(-1)
	}
	dr[len(dr)-1] = 0.0
	for i := len(ac.Nodes) - 1; i >= 0; i-- {
		switch n := ac.Nodes[i].(type) {
		case MulNode:
			for j, ej := range n {
				other := 0.0
				for k, ek := range n {
					if j != k {
						other += pr[ek]
					}
				}
				dr[ej] = logSumExp(dr[ej], dr[i]+other)
			}
		case AddNode:
			for _, e := range n {
				dr[e] = logSumExp(dr[e], dr[i])
			}
		}
	}
	return dr
}

func X2Ass(xs []int, schema []int) [][]float64 {
	ass := make([][]float64, len(xs))
	for i, x := range xs {
		ass[i] = make([]float64, schema[i])
		ass[i][x] = 1.0
	}
	return ass
}

func logSumExp(as ...float64) float64 {
	return logSumExpF(len(as), func(i int) float64 {
		return as[i]
	})
}

func logSumExpF(n int, f func(i int) float64) float64 {
	max := math.Inf(-1)
	for i := 0; i < n; i++ {
		max = math.Max(max, f(i))
	}
	if math.IsInf(max, 0) {
		return max
	}
	sum := 0.0
	for i := 0; i < n; i++ {
		sum += math.Exp(f(i) - max)
	}
	return math.Log(sum) + max
}

func parseInt(s string) int {
	r, e := strconv.ParseInt(s, 0, 0)
	if e != nil {
		log.Fatal(e)
	}
	return int(r)
}

func parseFloat(s string) float64 {
	r, e := strconv.ParseFloat(s, 64)
	if e != nil {
		log.Fatal(e)
	}
	return r
}
