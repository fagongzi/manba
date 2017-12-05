package goetty

import (
	"testing"
)

func Test_SyncPool_AllocSmall(t *testing.T) {
	pool := NewSyncPool(128, 1024, 2)
	mem := pool.Alloc(64)
	EqualNow(t, len(mem), 64)
	EqualNow(t, cap(mem), 128)
	pool.Free(mem)
}

func Test_SyncPool_AllocLarge(t *testing.T) {
	pool := NewSyncPool(128, 1024, 2)
	mem := pool.Alloc(2048)
	EqualNow(t, len(mem), 2048)
	EqualNow(t, cap(mem), 2048)
	pool.Free(mem)
}

func Benchmark_SyncPool_AllocAndFree_128(b *testing.B) {
	pool := NewSyncPool(128, 1024, 2)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.Free(pool.Alloc(128))
		}
	})
}

func Benchmark_SyncPool_AllocAndFree_256(b *testing.B) {
	pool := NewSyncPool(128, 1024, 2)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.Free(pool.Alloc(256))
		}
	})
}

func Benchmark_SyncPool_AllocAndFree_512(b *testing.B) {
	pool := NewSyncPool(128, 1024, 2)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.Free(pool.Alloc(512))
		}
	})
}

func Benchmark_SyncPool_CacheMiss_128(b *testing.B) {
	pool := NewSyncPool(128, 1024, 2)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.Alloc(128)
		}
	})
}

func Benchmark_SyncPool_CacheMiss_256(b *testing.B) {
	pool := NewSyncPool(128, 1024, 2)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.Alloc(256)
		}
	})
}

func Benchmark_SyncPool_CacheMiss_512(b *testing.B) {
	pool := NewSyncPool(128, 1024, 2)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.Alloc(512)
		}
	})
}

func Benchmark_Make_128(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var x []byte
		for pb.Next() {
			x = make([]byte, 128)
		}
		x = x[:0]
	})
}

func Benchmark_Make_256(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var x []byte
		for pb.Next() {
			x = make([]byte, 256)
		}
		x = x[:0]
	})
}

func Benchmark_Make_512(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var x []byte
		for pb.Next() {
			x = make([]byte, 512)
		}
		x = x[:0]
	})
}
