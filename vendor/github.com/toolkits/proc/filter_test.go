package proc

import (
	"runtime"
	"strconv"
	"sync"
	"testing"
)

func TestFilter(t *testing.T) {
	// legalOpt
	if !(legalOpt("eq") && legalOpt("ne") && legalOpt("gt") && legalOpt("lt")) {
		t.Error("error, legalOpt")
	}
	if !(!legalOpt("negt")) {
		t.Error("error, legalOpt")
	}
	// compute
	if !(compute("eq", 1.2345679999, 1.2345673333)) {
		t.Error("error, compute")
	}
	if !(compute("ne", 1.2345689999, 1.2345673333)) {
		t.Error("error, compute")
	}
	if !(compute("gt", 1.2345689999, 1.2345673333)) {
		t.Error("error, compute")
	}
	if !(compute("lt", 1.2345673333, 1.2345689999)) {
		t.Error("error, compute")
	}
	// DataFilter
	f := NewDataFilter("test", 2)
	f.SetFilter("pk123", "gt", 10.0)
	f.Filter("pk123", 10.0, "10.0")
	if !(len(f.GetAllFiltered()) != 1) {
		t.Error("error, Filter")
	}
	f.Filter("pk123", 10.000001, "10.000001")
	if !(len(f.GetAllFiltered()) != 1) {
		t.Error("error, Filter")
	}
	f.Filter("pk123", 10.0000005, "10.0000005")
	if !(len(f.GetAllFiltered()) != 2) {
		t.Error("error, Filter")
	}
	f.Filter("pk123", 10.0, "10.0")
	if !(len(f.GetAllFiltered()) != 2) {
		t.Error("error, Filter")
	}
}

func BenchmarkFilter(b *testing.B) {
	b.StopTimer()
	b.N = 5000000
	f := NewDataFilter("name", 3)
	f.SetFilter("1000", "gt", 2000.0)
	strmap := make([]string, 0)
	for i := 0; i < 10000; i++ {
		strmap = append(strmap, strconv.Itoa(i))
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		f.Filter(strmap[i%10000], float64(i), i)
	}
	if len(f.GetAllFiltered()) != 3 {
		b.Error("error, Filter")
	}
}

func BenchmarkFilterConcurrent(b *testing.B) {
	b.StopTimer()
	b.N = 5000000
	f := NewDataFilter("name", 3)
	f.SetFilter("1000", "gt", 2000.0)
	strmap := make([]string, 0)
	for i := 0; i < 10000; i++ {
		strmap = append(strmap, strconv.Itoa(i))
	}

	wg := sync.WaitGroup{}
	workers := runtime.NumCPU()
	each := b.N / workers
	wg.Add(workers)

	b.StartTimer()

	for i := 0; i < workers; i++ {
		go func() {
			for i := 0; i < each; i++ {
				f.Filter(strmap[i%10000], float64(i), i)
			}
			wg.Done()
		}()
	}

	wg.Wait()
	if len(f.GetAllFiltered()) != 3 {
		b.Error("error, Filter")
	}
}
