package riblt

import (
	"testing"
)

func BenchmarkRIBLTEncode(b *testing.B) {
	benches := []struct {
		name string
		size int
	}{
		{"m=10", 10},
		{"m=20", 20},
		{"m=40", 40},
		{"m=80", 80},
		{"m=160", 160},
		{"m=320", 320},
		{"m=1000", 1000},
		{"m=10000", 10000},
		{"m=100000", 100000},
		{"m=1000000", 1000000},
		{"m=10000000", 10000000},
	}
	for _, bench := range benches {
		s := make(Sketch, bench.size)
		b.Run(bench.name, func(b *testing.B) {
			b.SetBytes(HashTypeSize)
			for i := 0; i < b.N; i++ {
				s.AddSymbol(HashType(i))
			}
		})
	}
}

func BenchmarkRIBLTDecode(bc *testing.B) {
	cases := []struct {
		name string
		size int
	}{
		{"d=10", 10},
		{"d=20", 20},
		{"d=40", 40},
		{"d=80", 80},
		{"d=160", 160},
		{"d=320", 320},
		{"d=1000", 1000},
		{"d=10000", 10000},
		// {"d=50000", 50000},
		// {"d=100000", 100000},
	}
	for _, tc := range cases {
		bc.Run(tc.name, func(b *testing.B) {
			d := tc.size
			n := tc.size
			b.SetBytes(HashTypeSize * int64(tc.size))
			log := make([]HashType, d + n)
			ncw := 0
			var nextId uint64
			b.ResetTimer()
			b.StopTimer()
			for iter := 0; iter < b.N; iter++ {
				for i := 0; i < d + n; i++ {
					nextId += 1
					log[i] = nextId
				}

				multiplier := 2
				for {
					slocal := make(Sketch, d * multiplier)
					sremote := make(Sketch, d * multiplier)
					for i := 0; i < d + n; i++ {
						slocal.AddSymbol(log[i])
					}
					for i := 0; i < n; i++ {
						sremote.AddSymbol(log[i])
					}

					b.StartTimer()
					slocal.Subtract(sremote)
					_, _, succ := slocal.Decode()
					b.StopTimer()
					if succ {
						ncw += multiplier
						break
					}
					multiplier *= 2
				}
			}
			b.ReportMetric(float64(ncw)/float64(b.N), "symbols/diff")
		})
	}
}

func TestFixedEncodeAndDecode(t *testing.T) {
	cases := []struct {
		name string
		size int
	}{
		{"d=10", 10},
		{"d=20", 20},
		{"d=40", 40},
		{"d=100", 100},
		{"d=1000", 1000},
		{"d=10000", 10000},
		{"d=50000", 50000},
		{"d=100000", 100000},
	}
	for _, tc := range cases {
		nlocal := tc.size
		ncommon := tc.size
		var nextId uint64
		slocal := make(Sketch, nlocal * 2)
		sremote := make(Sketch, nlocal * 2)
		for i := 0; i < nlocal; i++ {
			s := newTestSymbol(nextId)
			nextId += 1
			slocal.AddSymbol(s.Hash())
		}
		for i := 0; i < ncommon; i++ {
			s := newTestSymbol(nextId)
			nextId += 1
			slocal.AddSymbol(s.Hash())
			sremote.AddSymbol(s.Hash())
		}

		// Decode
		slocal.Subtract(sremote)
		fwd, rev, succ := slocal.Decode()
		if !succ {
			t.Errorf("(size=%d) failed to decode at all", tc.size)
		}
		if len(rev) != 0 {
			t.Errorf("(size=%d) failed to detect subset", tc.size)
		}
		if len(fwd) != nlocal {
			t.Errorf("(size=%d) missing symbols: %d local", tc.size, len(fwd))
		}
	}
}
