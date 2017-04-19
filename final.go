package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	BT   = flag.Bool("BT", false, "Best Tree method")
	NG   = flag.Bool("NG", false, "Normalized Greedy method")
	AMAP = flag.Bool("AMAP", false, "Argmax-Product method")

	BS   = flag.Bool("BS", false, "Beam Search method")
	BS_B = flag.Int("BS_B", 10, "Beam size in BS method")

	KBT   = flag.Bool("KBT", false, "K-Best Tree method")
	KBT_K = flag.Int("KBT_K", 100, "K in KBT method")

	MP       = flag.Bool("MP", false, "Marginal Pruning approach")
	FC       = flag.Bool("FC", false, "Forward Checking approach")
	ORDERING = flag.Bool("ORDERING", false, "Ordering approach")
	STAGE    = flag.Bool("STAGE", false, "Stage approach")

	QEH = flag.String("QEH", "", "MAP query DIR")

	TIMEOUT     = flag.Int("TIMEOUT", 600, "Timeout (in seconds)")
	GROUP_COUNT = flag.Int("GROUP_COUNT", QUERY_COUNT, "Group count")
)

func FinalExperiment() {
	if *QEH == "" {
		return
	}
	os.Mkdir(RESULT_DIR+*QEH, 0777)
	switch {
	case *BT:
		mapInference("BT", BTMethod)
	case *NG:
		mapInference("NG", NGMethod)
	case *AMAP:
		mapInference("AMAP", AMAPMethod)
	case *BS:
		mapInference("BS", BSMethod)
	case *KBT:
		mapInference("KBT", KBTMethod)
	case *MP:
		mapInference("MP", MPMethod)
	case *FC:
		mapInference("FC", FCMethod)
	case *ORDERING:
		mapInference("ORDERING", ORDERINGMethod)
	case *STAGE:
		mapInference("STAGE", STAGEMethod)
	}
}

type MAXMethod func(spn SPN) float64

func mapInference(methodName string, method MAXMethod) {
	path := fmt.Sprintf("%s%s/%s/", RESULT_DIR, *QEH, methodName)
	if err := os.MkdirAll(path, 0777); err != nil {
		log.Fatalf("Mkdir %s: %v\n", path, err)
	}
	if err := os.MkdirAll(path+"time", 0777); err != nil {
		log.Fatalf("Mkdir %stime: %v\n", path, err)
	}
	if err := os.MkdirAll(path+"result", 0777); err != nil {
		log.Fatalf("Mkdir %sresult: %v\n", path, err)
	}
	for _, dataset := range DATASETS {
		log.Printf("[DOING]%s %s\n", methodName, dataset)
		mapInferenceDataset(path, dataset, methodName, method)
	}
}
func mapInferenceDataset(path string, dataset string, methodName string, method MAXMethod) {
	spn := LoadSPN(SPN_DIR + dataset)
	qehPath := fmt.Sprintf("%s%s/%s", QEH_DIR, *QEH, dataset)
	qeh, err := ioutil.ReadFile(qehPath)
	if err != nil {
		log.Fatalf("ReadFile %s: %v\n", qehPath, err)
	}
	qeh = bytes.TrimSpace(qeh)
	qehs := bytes.Split(qeh, []byte{'\n'})
	if len(qehs) != QUERY_COUNT {
		log.Fatal("Query count doesn't match:", len(qehs), QUERY_COUNT)
	}
	res := make([]float64, len(qehs))
	tim := make([]float64, len(qehs))
	wg := sync.WaitGroup{}
	for i, q := range qehs {
		i, q := i, q
		q = bytes.Replace(q, []byte{','}, []byte{}, -1)
		wg.Add(1)
		go func() {
			querySPN := spn.QuerySPN(q)

			tic := time.Now()
			res[i] = method(querySPN)
			tim[i] = time.Since(tic).Seconds()

			if tim[i] > float64(*TIMEOUT)-1 {
				log.Printf("TIMEOUT: %s %s %s %d", path, dataset, methodName, i)
			}
			wg.Done()
		}()
		if (i+1)%*GROUP_COUNT == 0 {
			wg.Wait()
		}
	}
	wg.Wait()
	resData := []byte{}
	for _, r := range res {
		resData = strconv.AppendFloat(resData, r, 'f', -1, 64)
		resData = append(resData, '\n')
	}
	timData := []byte{}
	for _, t := range tim {
		timData = strconv.AppendFloat(timData, t, 'f', -1, 64)
		timData = append(timData, '\n')
	}
	if err := ioutil.WriteFile(fmt.Sprintf("%stime/%s", path, dataset), timData, 0777); err != nil {
		log.Fatal("Write time", path, dataset, err)
	}
	if err := ioutil.WriteFile(fmt.Sprintf("%sresult/%s", path, dataset), resData, 0777); err != nil {
		log.Fatal("Write result", path, dataset, err)
	}
}

func BTMethod(spn SPN) float64 {
	return spn.EvalX(MaxMax(spn))
}
func NGMethod(spn SPN) float64 {
	return spn.EvalX(SumMax(spn))
}
func AMAPMethod(spn SPN) float64 {
	return amap(spn).P
}
func BSMethod(spn SPN) float64 {
	return BeamSearch(spn, PrbK(spn, *BS_B), *BS_B).P
}
func KBTMethod(spn SPN) float64 {
	xs := TopKMaxMax(spn, *KBT_K)
	return MaxXP(EvalXBatch(spn, xs)).P
}
func MPMethod(spn SPN) float64 {
	return 0.0
}
func FCMethod(spn SPN) float64 {
	return 0.0
}
func ORDERINGMethod(spn SPN) float64 {
	return 0.0
}
func STAGEMethod(spn SPN) float64 {
	return 0.0
}

const (
	EXPERIMENT_DIR = "experiment/"
	SPN_DIR        = EXPERIMENT_DIR + "learn.spn/"
	QEH_DIR        = EXPERIMENT_DIR + "map.qeh/"
	RESULT_DIR     = EXPERIMENT_DIR + "result.csv/"

	QUERY_COUNT = 1000
)

var DATASETS = []string{
	"nltcs", "msnbc", "kdd", "plants", "baudio",
	"bnetflix", "jester", "accidents", "tretail", "pumsb_star",
	"dna", "kosarek", "msweb", "book", "tmovie",
	"cwebkb", "cr52", "c20ng", "bbc", "ad",
}

func GenerateQEH() {
	for q := 1; q <= 9; q++ {
		for e := 0; q+e <= 10; e++ {
			h := 10 - e - q
			generateQEH(q, e, h)
		}
	}
}

func generateQEH(q, e, h int) {
	dir := fmt.Sprintf("%s%d%d%d/", QEH_DIR, q, e, h)
	err := os.Mkdir(dir, 0777)
	if err != nil {
		log.Fatalf("Mkdir %s error: %v\n", dir, err)
	}
	for _, dataset := range DATASETS {
		log.Printf("Generating %d%d%d %s\n", q, e, h, dataset)
		spn := LoadSPN(SPN_DIR + dataset)
		generateQEHFile(dir+dataset, spn.Schema, q, e, h)
	}
}

const (
	QEH_QUERY  = -1
	QEH_HIDDEN = -2
)

func generateQEHFile(filename string, schema []int, q, e, h int) {
	data := []byte{}
	qq := 1
	if tmp := len(schema) * q / 10; qq < tmp {
		qq = tmp
	}
	ee := len(schema) * e / 10
	hh := len(schema) - qq - ee
	for cntKth := 0; cntKth < QUERY_COUNT; cntKth++ {
		qeh := generateQEH1(schema, qq, ee, hh)
		for i, v := range qeh {
			switch v {
			case QEH_QUERY:
				data = append(data, '?')
			case QEH_HIDDEN:
				data = append(data, '*')
			default:
				data = strconv.AppendInt(data, int64(v), 10)
			}
			if i == len(qeh)-1 {
				data = append(data, '\n')
			} else {
				data = append(data, ',')
			}
		}
	}
	err := ioutil.WriteFile(filename, data, 0777)
	if err != nil {
		log.Fatalf("WriteFile %s: %v\n", filename, err)
	}
}
func generateQEH1(schema []int, q int, e int, h int) []int {
	qeh := make([]int, len(schema))
	for i := range schema {
		switch {
		case i < q:
			qeh[i] = QEH_QUERY
		case i < q+h:
			qeh[i] = QEH_HIDDEN
		default:
			qeh[i] = rand.Intn(schema[i])
		}
	}
	qeh2 := make([]int, len(qeh))
	for i, v := range rand.Perm(len(qeh)) {
		qeh2[i] = qeh[v]
	}
	return qeh2
}

func amap(spn SPN) XP {
	mc := make([]XP, len(spn.Nodes))
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(*TIMEOUT)*time.Second)
	timeout := XP{X: nil, P: math.NaN()}
	for i, n := range spn.Nodes {
		select {
		case <-ctx.Done():
			return timeout
		default:
		}
		switch n := n.(type) {
		case *Trm:
			x := make([]int, len(spn.Schema))
			for xi := range x {
				x[xi] = -1
			}
			x[n.Kth] = n.Value
			mc[i] = XP{x, 0}
		case *Sum:
			xpBest := XP{nil, math.Inf(-1)}
			for _, e := range n.Edges {
				select {
				case <-ctx.Done():
					return timeout
				default:
				}
				p := amapEvalAt(spn, mc[e.Node.ID()].X, i)
				if xpBest.P < p {
					xpBest = XP{mc[e.Node.ID()].X, p}
				}
			}
			mc[i] = xpBest
		case *Prd:
			x := make([]int, len(spn.Schema))
			for xi := range x {
				x[xi] = -1
			}
			for _, e := range n.Edges {
				xe := mc[e.Node.ID()].X
				for xi := range xe {
					if xe[xi] != -1 {
						x[xi] = xe[xi]
					}
				}
			}
			mc[i] = XP{x, amapEvalAt(spn, x, i)}
		}
	}
	return mc[len(spn.Nodes)-1]
}

func amapEvalAt(spn SPN, x []int, at int) float64 {
	val := make([]float64, at+1)
	for i := 0; i <= at; i++ {
		n := spn.Nodes[i]
		switch n := n.(type) {
		case *Trm:
			var v float64
			if x[n.Kth] == n.Value {
				v = 0
			} else {
				v = math.Inf(-1)
			}
			val[i] = v
		case *Sum:
			val[i] = logSumExpF(len(n.Edges), func(k int) float64 {
				return n.Edges[k].Weight + val[n.Edges[k].Node.ID()]
			})
		case *Prd:
			prd := 0.0
			for _, e := range n.Edges {
				prd += val[e.Node.ID()]
			}
			val[i] = prd
		}
	}
	return val[at]
}
