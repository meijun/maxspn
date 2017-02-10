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

type AC []ACNode

type ACID int

type VarNode struct {
	Kth   int
	Value int
}
type NumNode float64
type MulNode []ACID
type AddNode []ACID

func LoadAC(acFile string) AC {
	bs, err := ioutil.ReadFile(acFile)
	if err != nil {
		log.Fatal(err)
	}
	bss := bytes.Split(bs, []byte("\n"))
	end := len(bss) - 1
	for string(bss[end]) != "EOF" {
		end--
	}
	bss = bss[1:end]
	ac := make(AC, len(bss))
	for i, ln := range bss {
		bs := bytes.Split(ln, []byte(" "))
		switch string(bs[0]) {
		case "v":
			ac[i] = VarNode{
				Kth:   parseInt(string(bs[1])),
				Value: parseInt(string(bs[2])),
			}
		case "n":
			ac[i] = NumNode(parseFloat(string(bs[1])))
		case "*":
			mn := make(MulNode, len(bs)-1)
			for j, v := range bs[1:] {
				mn[j] = ACID(parseInt(string(v)))
			}
			ac[i] = mn
		case "+":
			an := make(AddNode, len(bs)-1)
			for j, v := range bs[1:] {
				an[j] = ACID(parseInt(string(v)))
			}
			ac[i] = an
		}
	}
	return ac
}

type Assign [][]float64

// Return log(Pr(x))
func (ac AC) Pr(assign Assign) float64 {
	nn := len(ac)
	val := make([]float64, nn)
	for i, n := range ac {
		switch n := n.(type) {
		case VarNode:
			val[i] = math.Log(assign[n.Kth][n.Value])
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
	return val[len(val)-1]
}

func (ac AC) Info() (varNode, numNode, mulNode, addNode int, mulEdge, addEdge Stat) {
	me := []float64{}
	ae := []float64{}
	sc := make([]*big.Int, len(ac))
	for i, n := range ac {
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

func logSumExpF(n int, f func(i int) float64) float64 {
	max := math.Inf(-1)
	for i := 0; i < n; i++ {
		max = math.Max(max, f(i))
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
