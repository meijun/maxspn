package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"math"
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
			sum := 0.0
			max := math.Inf(-1)
			for _, ai := range n {
				max = math.Max(max, val[ai])
			}
			for _, ai := range n {
				sum += math.Exp(val[ai] - max)
			}
			val[i] = math.Log(sum) + max
		}
	}
	return val[len(val)-1]
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
