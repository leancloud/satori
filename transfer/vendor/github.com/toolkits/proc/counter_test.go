package proc

import (
	"runtime"
	"sync"
	"testing"
)

func BenchmarkSCounterBaseIncr(b *testing.B) {
	b.StopTimer()
	b.N = 5000000
	cnt := NewSCounterBase("cnt.base")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cnt.Incr()
	}
	if int(cnt.Cnt) != b.N {
		b.Error("error, SCounterBase.Incr not safe")
	}
}

func BenchmarkSCounterBaseIncrConcurrent(b *testing.B) {
	b.StopTimer()
	b.N = 5000000
	cnt := NewSCounterBase("cnt.base")
	wg := sync.WaitGroup{}
	workers := runtime.NumCPU()
	each := b.N / workers
	wg.Add(workers)
	b.StartTimer()
	for i := 0; i < workers; i++ {
		go func() {
			for i := 0; i < each; i++ {
				cnt.Incr()
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if int(cnt.Cnt) != b.N {
		b.Error("error, SCounterBase.Incr concurrently not safe")
	}
}

func BenchmarkSCounterQpsIncr(b *testing.B) {
	b.StopTimer()
	b.N = 5000000
	cnt := NewSCounterQps("cnt.qps")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cnt.Incr()
	}
	if int(cnt.Cnt) != b.N {
		b.Error("error, SCounterQps.Incr not safe")
	}
}

func BenchmarkSCounterQpsIncrConcurrent(b *testing.B) {
	b.StopTimer()
	b.N = 5000000
	cnt := NewSCounterQps("cnt.qps")
	wg := sync.WaitGroup{}
	workers := runtime.NumCPU()
	each := b.N / workers
	wg.Add(workers)
	b.StartTimer()
	for i := 0; i < workers; i++ {
		go func() {
			for i := 0; i < each; i++ {
				cnt.Incr()
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if int(cnt.Cnt) != b.N {
		b.Error("error, SCounterQps.Incr concurrently not safe")
	}
}
