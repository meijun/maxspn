package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"sync"
	"time"
)

const (
	DATA_DIR = "data/"

	// the top one is source, the followings are generated files
	ID_AC        = DATA_DIR + "id-ac/"
	ID_AC_SPN    = DATA_DIR + "id-ac-spn/"
	ID_AC_SPN_AC = DATA_DIR + "id-ac-spn-ac/"

	LR_SPN    = DATA_DIR + "lr-spn/"
	LR_SPN_AC = DATA_DIR + "lr-spn-ac/"

	TY_SPN = DATA_DIR + "ty-spn/"

	QUERY = DATA_DIR + "query/"
)

var SPN_AC = map[string]string{
	ID_AC_SPN: ID_AC_SPN_AC,
	LR_SPN:    LR_SPN_AC,
}

var DATA_NAMES = []string{
	"nltcs", "msnbc", "kdd", "plants", "baudio", "bnetflix", "jester", "accidents", "tretail", "pumsb_star",
	"dna", "kosarek", "msweb", "book", "tmovie", "cwebkb", "cr52", "c20ng", "bbc", "ad",
}
var VAR_CNT = []int{
	16, 17, 64, 69, 100, 100, 100, 111, 135, 163,
	180, 190, 294, 500, 500, 839, 889, 910, 1058, 1556,
}
var DATA_VAR_CNT = map[string]int{}

func init() {
	for i, name := range DATA_NAMES {
		DATA_VAR_CNT[name] = VAR_CNT[i]
	}
}

// Given source, generate transformed data.
func PrepareData() {
	for _, name := range DATA_NAMES {
		// rename original ac file
		os.Rename(ID_AC+name+".ac", ID_AC+name)

		// id-ac => id-ac-spn => id-ac-spn-ac
		if _, err := os.Stat(ID_AC_SPN_AC + name); os.IsNotExist(err) {
			ac := LoadAC(ID_AC + name)
			spn := AC2SPN(ac)
			spn.Save(ID_AC_SPN + name)
			spn.SaveAsAC(ID_AC_SPN_AC + name)
		}

		// lr-spn => lr-spn-ac
		if _, err := os.Stat(LR_SPN_AC + name); os.IsNotExist(err) {
			spn := LoadSPN(LR_SPN + name)
			spn.SaveAsAC(LR_SPN_AC + name)
		}

		// TODO generate MAP query
		if _, err := os.Stat(QUERY + name); os.IsNotExist(err) {
			varCnt := DATA_VAR_CNT[name]
			data := []byte{}
			for row := 0; row < 100; row++ {
				for i := 0; i < varCnt; i++ {
					var c byte
					switch rand.Intn(3) {
					case 0: // hidden
						c = '*'
					case 1: // evidence
						c = byte('0' + rand.Intn(2))
					case 2: // query
						c = '?'
					}
					if i != 0 {
						data = append(data, ' ')
					}
					data = append(data, c)
				}
				data = append(data, '\n')
			}
			data = append(data, []byte("EOF")...)
			ioutil.WriteFile(QUERY+name, data, 0666)
		}
	}
	log.Println("[DONE] PrepareData")
}

func Prb1kMethod(spn SPN) float64 {
	return PrbKMax(spn, 1000).P
}
func MaxMaxMethod(spn SPN) float64 {
	return spn.EvalX(MaxMax(spn))
}
func SumMaxMethod(spn SPN) float64 {
	return spn.EvalX(SumMax(spn))
}
func NaiveBayesMethod(spn SPN) float64 {
	return spn.EvalX(NaiveBayes(spn))
}
func ExactMethod(spn SPN) float64 {
	return Exact(spn, math.Inf(-1))
	//return Exact(spn, spn.EvalX(MaxMax(spn)))
}
func ExactOrderMethod(spn SPN) float64 {
	return ExactOrder(spn, math.Inf(-1))
	//return ExactOrder(spn, spn.EvalX(MaxMax(spn)))
}
func ExactOrderDerMethod(spn SPN) float64 {
	return ExactOrderDer(spn, math.Inf(-1))
	//return ExactOrderDer(spn, spn.EvalX(MaxMax(spn)))
}
func Prb1kBSMethod(spn SPN) float64 {
	return BeamSearch(spn, PrbK(spn, 1000), 31).P
}
func MaxMaxBSMethod(spn SPN) float64 {
	x := MaxMax(spn)
	return BeamSearch(spn, []XP{{x, spn.EvalX(x)}}, 31).P
}
func SumMaxBSMethod(spn SPN) float64 {
	x := SumMax(spn)
	return BeamSearch(spn, []XP{{x, spn.EvalX(x)}}, 31).P
}
func TopKMaxMaxMethod(spn SPN) float64 {
	xs := TopKMaxMax(spn, 1000)
	return MaxXP(EvalXBatch(spn, xs)).P
}
func TopKMaxMaxBSMethod(spn SPN) float64 {
	xs := TopKMaxMax(spn, 1000)
	return BeamSearch(spn, EvalXBatch(spn, xs), 31).P
}
func MCMethod(spn SPN) float64 {
	return MC(spn).P
}

func Exp(dataSet string, method func(SPN) float64, label string) {
	tic := time.Now()
	res := make([]float64, len(DATA_NAMES))
	tim := make([]float64, len(DATA_NAMES))
	finally = append(finally, func() {
		log.Printf("[DONE][%s][TIME %.0f][RES] %s: %v\n", label, time.Since(tic).Seconds(), dataSet, res)
		log.Printf("[DONE][%s][TIME %.0f][TIME] %s: %v\n", label, time.Since(tic).Seconds(), dataSet, tim)
	})
	wg := sync.WaitGroup{}
	for i, name := range DATA_NAMES {
		wg.Add(1)
		i, name := i, name
		go func() {
			spn := LoadSPN(dataSet + name)
			ticMethod := time.Now()
			res[i] = method(spn)
			tim[i] = time.Since(ticMethod).Seconds()
			log.Printf("[DONE][TIME %f] %s%s: %v\n", tim[i], dataSet, DATA_NAMES[i], res[i])
			wg.Done()
		}()
	}
	wg.Wait()
}

func LibraSPNMPE(dataSet string) {
	res := make([]float64, len(DATA_NAMES))
	for i, name := range DATA_NAMES {
		res[i] = libraSPNMPE1(dataSet, name)
	}
	log.Println("[libra]", res)
}

func printResult1(label string, res []float64) {
	fmt.Print(label)
	for _, r := range res {
		fmt.Print("\t")
		fmt.Print(r)
	}
	fmt.Println()
}

func libraSPNMPE1(dataSet, dataName string) float64 {
	x := libraMPE1(SPN_AC[dataSet], dataName)
	if x != nil {
		spn := LoadSPN(dataSet + dataName)
		return spn.EvalX(x)
	} else {
		return math.Inf(-1)
	}
}

func libraMPE1(dataSet, dataName string) []int {
	varCnt := DATA_VAR_CNT[dataName]
	star := make([]byte, varCnt*2)
	for i := 0; i < varCnt; i++ {
		star[i*2] = '*'
		star[i*2+1] = ','
	}
	star[varCnt*2-1] = '\n'

	evFile, err := ioutil.TempFile("", "ev.")
	if err != nil {
		log.Fatalf("LibraMPE create ev TempFile error: %v\n", err)
	}
	defer os.Remove(evFile.Name())
	evFile.Write(star)
	evFile.Close()
	acFilename := dataSet + dataName
	cmd := exec.Command("libra", "acquery", "-m", acFilename, "-ev", evFile.Name(), "-mpe")
	res, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Libra error when(%s, %s): %v\n", dataSet, dataName, err)
		return nil
	}
	log.Printf("[libra] output: %s\n", string(res))
	x := make([]int, varCnt)
	for i := range x {
		if res[i*2] == '1' {
			x[i] = 1
		} else {
			x[i] = 0
		}
	}
	return x
}

func ACMaxMaxExp(dataSet string) {
	ps := make([]float64, len(DATA_NAMES))
	for i, name := range DATA_NAMES {
		ac := LoadAC(dataSet + name)
		x := ac.MaxMax()
		p := AC2SPN(ac).EvalX(x)
		log.Println(x, p)
		ps[i] = p
	}
	log.Println(ps)
}

func ExpQuerySPN(dataSet string) {
	for _, name := range DATA_NAMES {
		qf, err := ioutil.ReadFile(QUERY + name)
		if err != nil {
			log.Fatalf("ExpQuerySPN read file %s: %v\n", QUERY+name, err)
		}
		qs := bytes.Split(qf, []byte{'\n'})
		qCnt := len(qs) - 1
		for qCnt > 0 && string(qs[qCnt]) != "EOF" {
			qCnt--
		}
		if qCnt == 0 {
			log.Fatalf("Empty query file: %s\n", QUERY+name)
		}
		qs = qs[:qCnt]
		spn := LoadSPN(dataSet + name)
		for qj, q := range qs {
			qt := bytes.Split(q, []byte{' '})
			q = q[:len(qt)]
			for k := range qt {
				q[k] = qt[k][0]
			}
			qSPN := spn.QuerySPN(q)
			mm := qSPN.EvalX(MaxMax(qSPN))
			sm := qSPN.EvalX(SumMax(qSPN))
			ex := ExactOrderDer(qSPN, mm)
			eq := func(v float64) int {
				if math.Abs(v-ex) < 1e-6 {
					return 1
				}
				return 0
			}
			if math.Abs(ex-mm) > 1e-6 {
				t10 := MaxXP(EvalXBatch(qSPN, TopKMaxMax(qSPN, 10))).P
				t100 := MaxXP(EvalXBatch(qSPN, TopKMaxMax(qSPN, 100))).P
				t1k := MaxXP(EvalXBatch(qSPN, TopKMaxMax(qSPN, 1000))).P
				log.Printf("Found %s line: %d, mm: %f, exact: %f, %d%d%d\n", dataSet+name, qj+1, mm, ex, eq(t10), eq(t100), eq(t1k))
			}
			log.Printf("[DONE] %d %d%d\n", qj, eq(mm), eq(sm))
		}
		log.Printf("[DONE] %s\n", dataSet+name)
	}
}

const (
	QEH_Q = -1
	QEH_H = -2
)

func GenMAPQuery() [][][][]int {
	qehsss := [][][][]int{}
	for h := 1; h <= 8; h++ {
		q := 1
		//e := 10 - q - h
		//dir := fmt.Sprintf("%s%d%d%d.qeh", DATA_DIR, q, e, h)
		//os.Mkdir(dir, 0666)
		qehss := [][][]int{}
		for di /*, name*/ := range DATA_NAMES {
			//data := []byte{}
			qehs := [][]int{}
			for row := 0; row < 100; row++ {
				qeh := make([]int, VAR_CNT[di])
				qc := len(qeh) * q / 10
				hc := len(qeh) * h / 10
				for i := range qeh {
					switch {
					case i < qc:
						qeh[i] = QEH_Q
					case i < qc+hc:
						qeh[i] = QEH_H
					default: // evidence
						qeh[i] = rand.Intn(2)
					}
				}
				qeh = shuffle(qeh)
				qehs = append(qehs, qeh)
				//for i, v := range qeh {
				//	if i != 0 {
				//		data = append(data, ' ')
				//	}
				//	switch v {
				//	case QEH_Q:
				//		data = append(data, '?')
				//	case QEH_H:
				//		data = append(data, '*')
				//	default:
				//		data = append(data, byte(v + '0'))
				//	}
				//}
				//data = append(data, '\n')
			}
			//data = append(data, []byte("EOF")...)
			//if err := ioutil.WriteFile(dir + "/" + name, data, 0666); err != nil {
			//	log.Fatal(err)
			//}
			qehss = append(qehss, qehs)
		}
		qehsss = append(qehsss, qehss)
	}
	return qehsss
}
func shuffle(is []int) []int {
	rs := make([]int, len(is))
	for i, v := range rand.Perm(len(is)) {
		rs[i] = is[v]
	}
	return rs
}

func ExpMAP() {
	qsss := GenMAPQuery()
	for h := 1; h <= 8; h++ {
		for ni, name := range DATA_NAMES {
			spn := LoadSPN(LR_SPN + name)
			qs := qsss[h-1][ni]
			cntMC, cntT100 := 0, 0
			timMC, timT100 := 0., 0.
			mux := sync.Mutex{}
			wg := sync.WaitGroup{}
			for _, q := range qs {
				q := q
				wg.Add(1)
				go func() {
					qb := make([]byte, len(q))
					for i, v := range q {
						switch v {
						case QEH_Q:
							qb[i] = '?'
						case QEH_H:
							qb[i] = '*'
						default:
							qb[i] = byte('0' + v)
						}
					}
					qSPN := spn.QuerySPN(qb)
					tic := time.Now()
					mc := MC(qSPN).P
					mux.Lock()
					timMC += time.Since(tic).Seconds()
					mux.Unlock()
					tic = time.Now()
					t100 := MaxXP(EvalXBatch(qSPN, TopKMaxMax(qSPN, 100))).P
					mux.Lock()
					timT100 += time.Since(tic).Seconds()
					if math.Abs(mc-t100) > 1e-6 {
						if mc > t100 {
							cntMC++
						} else {
							cntT100++
						}
					}
					mux.Unlock()
					wg.Done()
				}()
			}
			wg.Wait()
			log.Printf("[h %d][%20s] cntMC: %3d, cntT100: %3d, timMC: %5.0f, timT100: %5.0f\n",
				h, name, cntMC, cntT100, timMC, timT100)
		}
	}
}
