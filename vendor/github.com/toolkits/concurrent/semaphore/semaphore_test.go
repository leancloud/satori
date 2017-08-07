package semaphore

import (
	"runtime"
	"sync"
	"testing"
)

func TestSemaphore(t *testing.T) {
	sema := NewSemaphore(2)
	if !(sema.TryAcquire() && sema.TryAcquire() && !sema.TryAcquire()) {
		t.Error("error, TryAcquire")
	}

	sema.Release()
	sema.Release()
}

func BenchmarkSemaphore(b *testing.B) {
	b.StopTimer()
	sema := NewSemaphore(1)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if sema.TryAcquire() {
			sema.Release()
		}
	}
}

func BenchmarkSemaphoreConcurrent(b *testing.B) {
	b.StopTimer()
	sema := NewSemaphore(1)
	wg := sync.WaitGroup{}
	workers := runtime.NumCPU()
	each := b.N / workers
	wg.Add(workers)
	b.StartTimer()
	for i := 0; i < workers; i++ {
		go func() {
			for i := 0; i < each; i++ {
				if sema.TryAcquire() {
					sema.Release()
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
