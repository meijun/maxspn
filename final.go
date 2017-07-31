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
	"runtime"
	"strconv"
	"strings"
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

	QEH  = flag.String("QEH", "", "MAP query DIR")
	DATA = flag.String("DATA", "", "MAP query DATASETS (delimited by ',')")

	TIMEOUT     = flag.Int("TIMEOUT", 600, "Timeout (in seconds)")
	GROUP_COUNT = flag.Int("GROUP_COUNT", 25, "Group count")

	COLUMN = flag.String("COLUMN", "BT", "Columns (delimited by ',')")
	TITLE  = flag.String("TITLE", "", "Title")

	WINCNT  = flag.Bool("WINCNT", false, "Summary table")
	FINISH  = flag.Bool("FINISH", false, "Finish before time limit exceed")
	TIMEAVG = flag.Bool("TIMEAVG", false, "Average time")
	RESAVG  = flag.Bool("RESAVG", false, "Average result")
	RESLSE  = flag.Bool("RESLSE", false, "Log Sum Exp result")
	BATTLE  = flag.Bool("BATTLE", false, "Battle")
)

func FinalExperiment() {
	if *QEH == "" {
		return
	}
	log.Printf("QEH: %s\n", *QEH)
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

	case *WINCNT:
		summary("WINCNT", summaryWINCNT)
	case *FINISH:
		summary("FINISH", summaryFINISH)
	case *TIMEAVG:
		summary("TIMEAVG", summaryTIMEAVG)
	case *RESAVG:
		summary("RESAVG", summaryRESAVG)
	case *RESLSE:
		summary("RESLSE", summaryRESLSE)
	case *BATTLE:
		summaryBATTLE("BATTLE")
	}
}

type MAXMethod func(spn SPN) float64

func mapInference(methodName string, method MAXMethod) {
	suffix := ""
	if *KBT {
		suffix = fmt.Sprintf("%d", *KBT_K)
	} else if *BS {
		suffix = fmt.Sprintf("%d", *BS_B)
	}
	path := fmt.Sprintf("%s%s/%s%s/", RESULT_DIR, *QEH, methodName, suffix)
	if err := os.MkdirAll(path, 0777); err != nil {
		log.Fatalf("Mkdir %s: %v\n", path, err)
	}
	if err := os.MkdirAll(path+"time", 0777); err != nil {
		log.Fatalf("Mkdir %stime: %v\n", path, err)
	}
	if err := os.MkdirAll(path+"result", 0777); err != nil {
		log.Fatalf("Mkdir %sresult: %v\n", path, err)
	}
	datasets := DATASETS
	if *DATA != "" {
		datasets = strings.Split(*DATA, ",")
	}
	for _, dataset := range datasets {
		mapInferenceDataset(path, dataset, methodName, method)
		log.Printf("[DONE]%s %s\n", methodName, dataset)
		runtime.GC()
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

			//if tim[i] > float64(*TIMEOUT)-1 {
			//	log.Printf("TIMEOUT: %s %s %s %d", path, dataset, methodName, i)
			//}
			wg.Done()
		}()
		if (i+1)%*GROUP_COUNT == 0 {
			wg.Wait()
			runtime.GC()
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
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(*TIMEOUT)*time.Second)
	return BeamSearchSerial(ctx, spn, PrbKSerial(spn, *BS_B), *BS_B).P
}
func KBTMethod(spn SPN) float64 {
	xs := TopKMaxMaxTimeout(spn, *KBT_K)
	if len(xs) == 0 {
		return math.NaN()
	}
	return MaxXP(EvalXBatchSerial(spn, xs)).P
}
func MPMethod(spn SPN) float64 {
	return ExactMP(spn, math.Inf(-1))
}
func FCMethod(spn SPN) float64 {
	return ExactFC(spn, math.Inf(-1))
}
func ORDERINGMethod(spn SPN) float64 {
	return ExactORDERING(spn, math.Inf(-1))
}
func STAGEMethod(spn SPN) float64 {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(*TIMEOUT)*time.Second)
	x := make([]int, len(spn.Schema))
	for i := range x {
		x[i] = -1
	}
	return ExactSTAGE(ctx, spn, x, math.Inf(-1))
}

const (
	EXPERIMENT_DIR = "experiment/"
	SPN_DIR        = EXPERIMENT_DIR + "learn.spn/"
	QEH_DIR        = EXPERIMENT_DIR + "map.qeh/"
	RESULT_DIR     = EXPERIMENT_DIR + "result.csv/"
	SUMMARY_DIR    = EXPERIMENT_DIR + "summary.csv/"

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
				p := evalAt(spn, mc[e.Node.ID()].X, i)
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
			mc[i] = XP{x, evalAt(spn, x, i)}
		}
	}
	return mc[len(spn.Nodes)-1]
}

func BeamSearchSerial(ctx context.Context, spn SPN, xps []XP, beamSize int) XP {
	best := XP{P: math.Inf(-1)}
	for i := 0; len(xps) > 0; i++ {
		xps = uniqueX(xps)
		xps = topK(xps, beamSize)
		xp1 := topK(xps, 1)
		if best.P < xp1[0].P {
			best = xp1[0]
		}
		select {
		case <-ctx.Done():
			return best
		default:
		}
		xps = nextGensSerial(ctx, xps, spn)
	}
	return best
}

func nextGensSerial(ctx context.Context, xps []XP, spn SPN) []XP {
	res := []XP{}
	resChan := make([]chan []XP, len(xps))
	for i, xp := range xps {
		ch := make(chan []XP, 1)
		select {
		case <-ctx.Done():
		default:
			nextGenD(xp, spn, ch)
		}
		resChan[i] = ch
	}
	for _, ch := range resChan {
		res = append(res, <-ch...)
	}
	return res
}
func PrbKSerial(spn SPN, k int) []XP {
	prt := partition(spn)
	res := make([]XP, k)
	wg := sync.WaitGroup{}
	for times := 0; times < k; times++ {
		wg.Add(1)
		func(i int) {
			x := prb1(spn, prt)
			p := spn.EvalX(x)
			res[i] = XP{x, p}
			wg.Done()
		}(times)
	}
	wg.Wait()
	return res
}
func EvalXBatchSerial(spn SPN, xs [][]int) []XP {
	xps := make([]XP, len(xs))
	for i, x := range xs {
		xps[i] = XP{x, spn.EvalX(x)}
	}
	return xps
}

func TopKMaxMaxTimeout(spn SPN, k int) [][]int {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(*TIMEOUT)*time.Second)
	ls := make([][]*Link, len(spn.Nodes))
	for i, n := range spn.Nodes {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		switch n := n.(type) {
		case *Trm:
			ls[i] = []*Link{{P: 0, Trm: n}}
		case *Sum:
			for _, e := range n.Edges {
				ls[i] = mergeSumLink(ls[i], ls[e.Node.ID()], 0, e.Weight, k)
			}
		case *Prd:
			for _, e := range n.Edges {
				ls[i] = mergePrdLink(ls[i], ls[e.Node.ID()], k)
			}
		}
	}
	if k > len(ls[len(spn.Nodes)-1]) {
		k = len(ls[len(spn.Nodes)-1])
	}
	xs := make([][]int, k)
	for i := range xs {
		xs[i] = make([]int, len(spn.Schema))
		topKMaxMaxDFS(ls[len(spn.Nodes)-1][i], xs[i])
	}
	return xs
}
func ExactMP(spn SPN, baseline float64) float64 {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(*TIMEOUT)*time.Second)
	x := make([]int, len(spn.Schema))
	return dfsMP(ctx, spn, x, 0, baseline)
}

func dfsMP(ctx context.Context, spn SPN, x []int, xi int, baseline float64) float64 {
	select {
	case <-ctx.Done():
		return baseline
	default:
	}

	if xi == len(spn.Schema) {
		return math.Max(baseline, eval(spn, x, xi))
	}
	x[xi] = 0
	if eval(spn, x, xi+1) > baseline {
		baseline = math.Max(baseline, dfsMP(ctx, spn, x, xi+1, baseline))
	}
	x[xi] = 1
	if eval(spn, x, xi+1) > baseline {
		baseline = math.Max(baseline, dfsMP(ctx, spn, x, xi+1, baseline))
	}
	return baseline
}

func ExactFC(spn SPN, baseline float64) float64 {
	x := make([]int, len(spn.Schema))
	for i := range x {
		x[i] = -1
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(*TIMEOUT)*time.Second)
	return dfsFC(ctx, spn, x, baseline)
}

func dfsFC(ctx context.Context, spn SPN, x []int, baseline float64) float64 {
	select {
	case <-ctx.Done():
		return baseline
	default:
	}

	x2 := make([]int, len(x))
	copy(x2, x)
	x = x2
	as := make([][]float64, len(x))
	for i := range as {
		as[i] = make([]float64, 2)
		if x[i] == 0 || x[i] == -1 {
			as[i][0] = 1
		}
		if x[i] == 1 || x[i] == -1 {
			as[i][1] = 1
		}
	}
	var d [][]float64
	for {
		updated := false
		d = derivativeOfAssignment(spn, as)
		for i := range x {
			if x[i] == -1 {
				xi0 := d[i][0]
				xi1 := d[i][1]

				if xi0 < baseline && xi1 < baseline {
					return baseline
				}
				if xi0 < baseline {
					x[i] = 1
					as[i][0] = 0
					updated = true
				}
				if xi1 < baseline {
					x[i] = 0
					as[i][1] = 0
					updated = true
				}
			}
		}
		if !updated {
			break
		}
	}
	maxVarID := -1
	maxValID := 0
	for i := range x {
		if x[i] == -1 {
			maxVarID = i
			break
		}
	}
	if i := maxVarID; i != -1 {
		x[i] = maxValID
		baseline = math.Max(dfsFC(ctx, spn, x, baseline), baseline)
		x[i] = maxValID ^ 1
		baseline = math.Max(dfsFC(ctx, spn, x, baseline), baseline)
		return baseline
	}
	return math.Max(baseline, math.Max(d[0][0], d[0][1]))
}

func ExactORDERING(spn SPN, baseline float64) float64 {
	x := make([]int, len(spn.Schema))
	for i := range x {
		x[i] = -1
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(*TIMEOUT)*time.Second)
	return dfsORDERING(ctx, spn, x, baseline)
}

func dfsORDERING(ctx context.Context, spn SPN, x []int, baseline float64) float64 {
	select {
	case <-ctx.Done():
		return baseline
	default:
	}

	x2 := make([]int, len(x))
	copy(x2, x)
	x = x2
	as := make([][]float64, len(x))
	for i := range as {
		as[i] = make([]float64, 2)
		if x[i] == 0 || x[i] == -1 {
			as[i][0] = 1
		}
		if x[i] == 1 || x[i] == -1 {
			as[i][1] = 1
		}
	}
	var d [][]float64
	for {
		updated := false
		d = derivativeOfAssignment(spn, as)
		for i := range x {
			if x[i] == -1 {
				xi0 := d[i][0]
				xi1 := d[i][1]

				if xi0 < baseline && xi1 < baseline {
					return baseline
				}
				if xi0 < baseline {
					x[i] = 1
					as[i][0] = 0
					updated = true
				}
				if xi1 < baseline {
					x[i] = 0
					as[i][1] = 0
					updated = true
				}
			}
		}
		if !updated {
			break
		}
	}
	maxVarID := -1
	maxValID := -1
	maxDer := math.Inf(-1)
	for i := range x {
		if x[i] == -1 {
			crtValID := 0
			crtDer := d[i][0]
			if d[i][0] < d[i][1] {
				crtValID = 1
				crtDer = d[i][1]
			}
			if maxVarID == -1 || maxDer < crtDer {
				maxVarID = i
				maxValID = crtValID
				maxDer = crtDer
			}
		}
	}
	if i := maxVarID; i != -1 {
		x[i] = maxValID
		baseline = math.Max(dfsORDERING(ctx, spn, x, baseline), baseline)
		x[i] = maxValID ^ 1
		baseline = math.Max(dfsORDERING(ctx, spn, x, baseline), baseline)
		return baseline
	}
	return math.Max(baseline, math.Max(d[0][0], d[0][1]))
}

func ExactSTAGE(ctx context.Context, spn SPN, x []int, best float64) float64 {
	select {
	case <-ctx.Done():
		return best
	default:
	}

	x2 := make([]int, len(x))
	copy(x2, x)
	x = x2
	var d [][]float64
	for {
		updated := false
		d = derivativeOfAssignmentX(spn, x)
		for i := range x {
			if x[i] == -1 {
				if d[i][0] <= best && d[i][1] <= best {
					return best
				}
				if d[i][0] <= best {
					x[i] = 1
					updated = true
				}
				if d[i][1] <= best {
					x[i] = 0
					updated = true
				}
			}
		}
		if !updated {
			break
		}
	}
	cnt := 0
	for i := range x {
		if x[i] == -1 {
			cnt++
		}
	}
	if cnt > 1 && len(x)-cnt >= 5 {
		spn = spn.StageSPN(x)
		x = make([]int, len(spn.Schema))
		for i := range x {
			x[i] = -1
		}
		d = derivativeOfAssignmentX(spn, x)
	}
	varID := -1
	valID := -1
	for i := range x {
		if x[i] == -1 {
			var valI int
			if d[i][0] < d[i][1] {
				valI = 1
			} else {
				valI = 0
			}
			if varID == -1 || d[varID][valID] < d[i][valI] {
				varID = i
				valID = valI
			}
		}
	}
	if varID == -1 {
		return math.Max(best, d[0][x[0]])
	}
	if cnt == 1 {
		return d[varID][valID]
	}
	x[varID] = valID
	best = ExactSTAGE(ctx, spn, x, best)
	x[varID] = 1 - valID
	best = ExactSTAGE(ctx, spn, x, best)
	return best
}

type ResTime struct {
	Result float64
	Time   float64
}

type summaryFunc func(resData [][]ResTime) []string

func summary(title string, sf summaryFunc) {
	if *TITLE != "" {
		title = *TITLE
	}
	cols := strings.Split(*COLUMN, ",")
	data := []byte{}
	data = append(data, []byte(title)...)
	for _, c := range cols {
		data = append(data, ',')
		data = append(data, []byte(c)...)
	}
	data = append(data, '\n')
	for _, dataset := range DATASETS {
		rtss := make([][]ResTime, len(cols))
		for ri := range rtss {
			resFile := fmt.Sprintf("%s%s/%s/%s/%s", RESULT_DIR, *QEH, cols[ri], "result", dataset)
			ress := readFloat(resFile)
			timeFile := fmt.Sprintf("%s%s/%s/%s/%s", RESULT_DIR, *QEH, cols[ri], "time", dataset)
			times := readFloat(timeFile)
			rts := make([]ResTime, len(ress))
			for j := range rts {
				rts[j] = ResTime{
					Result: ress[j],
					Time:   times[j],
				}
			}
			rtss[ri] = rts
		}
		data = append(data, []byte(dataset)...)
		row := sf(rtss)
		for _, r := range row {
			data = append(data, ',')
			data = append(data, []byte(r)...)
		}
		data = append(data, '\n')
	}
	os.Mkdir(SUMMARY_DIR+*QEH, 0777)
	if err := ioutil.WriteFile(SUMMARY_DIR+*QEH+"/"+title, data, 0777); err != nil {
		log.Fatal("WriteFile:", err)
	}
}

func readFloat(filename string) []float64 {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("ReadFile:", err)
	}
	raw = bytes.TrimSpace(raw)
	raws := bytes.Split(raw, []byte{'\n'})
	if len(raws) != QUERY_COUNT {
		log.Fatal("Result count is not equal:", len(raws), QUERY_COUNT)
	}
	fs := make([]float64, len(raws))
	for fi, r := range raws {
		fs[fi] = parseFloat(string(r))
	}
	return fs
}

func summaryWINCNT(resData [][]ResTime) []string {
	cnt := make([]int, len(resData))
	for j := range resData[0] {
		max := math.Inf(-1)
		for i := range resData {
			r := resData[i][j].Result
			if !math.IsNaN(r) {
				max = math.Max(max, r)
			}
		}
		for i := range resData {
			r := resData[i][j].Result
			if !math.IsNaN(r) && floatEqual(r, max) {
				cnt[i]++
			}
		}
	}
	res := i2s(cnt)
	return res
}
func i2s(cnt []int) []string {
	res := make([]string, len(cnt))
	for i := range cnt {
		res[i] = fmt.Sprint(cnt[i])
	}
	return res
}
func f2s(cnt []float64) []string {
	res := make([]string, len(cnt))
	for i := range cnt {
		res[i] = fmt.Sprint(cnt[i])
	}
	return res
}

const EPSILON = 1e-6

func floatEqual(x float64, y float64) bool {
	xsy := x - y
	if -EPSILON <= xsy && xsy <= EPSILON {
		return true
	}
	if -EPSILON <= x && x <= EPSILON || -EPSILON <= y && y <= EPSILON {
		return false
	}
	rx := xsy / x
	ry := xsy / y
	return (-EPSILON <= rx && rx <= EPSILON) || (-EPSILON <= ry && ry <= EPSILON)
}

func summaryFINISH(data [][]ResTime) []string {
	cnt := make([]int, len(data))
	for i := range data {
		for j := range data[i] {
			if data[i][j].Time < float64(*TIMEOUT-5) {
				cnt[i]++
			}
		}
	}
	return i2s(cnt)
}

func summaryTIMEAVG(data [][]ResTime) []string {
	res := make([]float64, len(data))
	for i := range data {
		sum := 0.0
		for j := range data[i] {
			sum += data[i][j].Time
		}
		sum /= float64(len(data[i]))
		res[i] = sum
	}
	return f2s(res)
}

func summaryRESAVG(data [][]ResTime) []string {
	res := make([]float64, len(data))
	for i := range data {
		sum := 0.0
		for j := range data[i] {
			r := data[i][j].Result
			if math.IsNaN(r) {
				sum = math.NaN()
				break
			}
			sum += r
		}
		sum /= float64(len(data[i]))
		res[i] = sum
	}
	return f2s(res)
}

func summaryRESLSE(data [][]ResTime) []string {
	res := make([]float64, len(data))
	for i := range data {
		sum := math.Inf(-1)
		for j := range data[i] {
			r := data[i][j].Result
			if math.IsNaN(r) {
				sum = math.NaN()
				break
			}
			sum = logSumExp(sum, r)
		}
		res[i] = sum
	}
	return f2s(res)
}

func summaryBATTLE(title string) {
	if *TITLE != "" {
		title = *TITLE
	}
	cols := strings.Split(*COLUMN, ",")
	data := []byte{}
	data = append(data, []byte(title)...)
	for _, c := range cols {
		data = append(data, ',')
		data = append(data, []byte(c)...)
	}
	data = append(data, '\n')
	res := make([][]int, len(cols))
	for i := range res {
		res[i] = make([]int, len(cols))
	}
	for _, dataset := range DATASETS {
		rtss := make([][]ResTime, len(cols))
		for ri := range rtss {
			resFile := fmt.Sprintf("%s%s/%s/%s/%s", RESULT_DIR, *QEH, cols[ri], "result", dataset)
			ress := readFloat(resFile)
			timeFile := fmt.Sprintf("%s%s/%s/%s/%s", RESULT_DIR, *QEH, cols[ri], "time", dataset)
			times := readFloat(timeFile)
			rts := make([]ResTime, len(ress))
			for j := range rts {
				rts[j] = ResTime{
					Result: ress[j],
					Time:   times[j],
				}
			}
			rtss[ri] = rts
		}
		battleDataset(res, rtss)
	}
	for i := range res {
		data = append(data, []byte(cols[i])...)
		for j := range res[i] {
			data = append(data, ',')
			data = append(data, []byte(fmt.Sprint(res[i][j]))...)
		}
		data = append(data, '\n')
	}
	os.Mkdir(SUMMARY_DIR+*QEH, 0777)
	if err := ioutil.WriteFile(SUMMARY_DIR+*QEH+"/"+title, data, 0777); err != nil {
		log.Fatal("WriteFile:", err)
	}
}
func battleDataset(res [][]int, data [][]ResTime) {
	for c := range data[0] {
		for i := range data {
			for j := range data {
				ri := data[i][c]
				rj := data[j][c]
				if !floatEqual(ri.Time, rj.Time) && ri.Time < float64(*TIMEOUT-5) && ri.Time < rj.Time &&
					!math.IsNaN(ri.Result) && (math.IsNaN(rj.Result) ||
					!floatEqual(ri.Result, rj.Result) && ri.Result > rj.Result) {
					res[i][j]++
				}
			}
		}
	}
}
