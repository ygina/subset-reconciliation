package quack

import (
	"testing"
	"math/rand"
)

func BenchmarkQuackEncode(b *testing.B) {
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
		s := NewSketch(bench.size)
		b.Run(bench.name, func(b *testing.B) {
			b.SetBytes(HashTypeSize)
			for i := 0; i < b.N; i++ {
				s.AddSymbol(HashType(i))
			}
		})
	}
}

func BenchmarkQuackDecode(bc *testing.B) {
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
			b.SetBytes(HashTypeSize * int64(d))
			log := make([]HashType, d + n)
			var nextId uint32
			b.ResetTimer()
			b.StopTimer()
			for iter := 0; iter < b.N; iter++ {
				for i := 0; i < d + n; i++ {
					nextId += 1
					log[i] = nextId
				}
				slocal := NewSketch(d)
				sremote := NewSketch(d)
				for i := 0; i < d + n; i++ {
					slocal.AddSymbol(log[i])
				}
				for i := 0; i < n; i++ {
					sremote.AddSymbol(log[i])
				}
				InitInverseTableUint32(d)

				// Shuffle log
				for i := len(log) - 1; i > 0; i-- {
					j := rand.Intn(i + 1)
					log[i], log[j] = log[j], log[i]
				}

				// Decode
				b.StartTimer()
				slocal.Subtract(sremote)
				slocal.Decode(log)
				b.StopTimer()
			}
			b.ReportMetric(1, "symbols/diff")
		})
	}
}

func TestAddSymbol(t *testing.T) {
	d := 20
	s := NewSketch(d)
	if s.Count != 0 {
		t.Errorf("wrong count: expected=0, actual=%d", s.Count)
	}
	s.AddSymbol(10)
	if s.Count != 1 {
		t.Errorf("wrong count: expected=1, actual=%d", s.Count)
	}
	s.AddSymbol(20)
	s.AddSymbol(30)
	if s.Count != 3 {
		t.Errorf("wrong count: expected=3, actual=%d", s.Count)
	}
}

func TestToCoeffs(t *testing.T) {
	x1 := uint32(1)
	x2 := uint32(2)
	d := 20
	s := NewSketch(d)
	s.AddSymbol(x1)
	s.AddSymbol(x2)

	InitInverseTableUint32(d)
	coeffs := s.ToCoeffs()
	expected := []ModUint32{
		ModUint32(x1 + x2).Neg(),
		ModUint32(x1 * x2),
	}
	if len(coeffs) != 2 {
		t.Errorf("expected %d != %d coeffs", len(expected), len(coeffs))
	}
	if len(coeffs) >= 1 && coeffs[0] != expected[0] {
		t.Errorf("expected first coeff %d != %d", expected[0], coeffs[0])
	}
	if len(coeffs) >= 2 && coeffs[1] != expected[1] {
		t.Errorf("expected first coeff %d != %d", expected[1], coeffs[1])
	}
}

func checkDecode(t *testing.T, fwd []HashType, succ bool, expected_fwd []HashType, expected_succ bool) {
	fast_fail := succ != expected_succ
	fast_fail = fast_fail || (len(fwd) != len(expected_fwd))
	if !fast_fail {
		for i := 0; i < len(fwd); i++ {
			if fwd[i] != expected_fwd[i] {
				fast_fail = true
				break
			}
		}
	}
	if fast_fail {
		t.Errorf("decoding failed %v %t", fwd, succ)
	}
}

func TestSubtract(t *testing.T) {
	const (
		x1 uint32 = 3616712547
		x2 uint32 = 2333013068
		x3 uint32 = 2234311686
		x4 uint32 = 448751902
		x5 uint32 = 918748965
	)

	d := 20
	s1 := NewSketch(d)
	s2 := NewSketch(d)
	InitInverseTableUint32(d)
	s1.AddSymbol(x4)
	s1.AddSymbol(x5)
	s2.AddSymbol(x1)
	s2.AddSymbol(x2)
	s2.AddSymbol(x3)
	s2.AddSymbol(x4)
	s2.AddSymbol(x5)
	if s1.Count != 2 {
		t.Errorf("s1 wrong count %d", s1.Count)
	}
	if s2.Count != 5 {
		t.Errorf("s2 wrong count %d", s2.Count)
	}
	if s1.PowerSums[0] != ModUint32(x4).Add(ModUint32(x5)) {
		t.Errorf("s1 wrong power sums %v", s1.PowerSums)
	}
	if s2.PowerSums[0] != ModUint32(x1).Add(ModUint32(x2)).Add(ModUint32(x3)).Add(ModUint32(x4)).Add(ModUint32(x5)) {
		t.Errorf("s2 wrong power sums %v", s2.PowerSums)
	}

	s2.Subtract(s1)
	if s2.Count != 3 {
		t.Errorf("s2 wrong count %d", s2.Count)
	}
	if s2.PowerSums[0] != ModUint32(x1).Add(ModUint32(x2)).Add(ModUint32(x3)) {
		t.Errorf("s2 wrong power sums %v", s2.PowerSums)
	}
}

func TestDecode(t *testing.T) {
	const (
		x1 uint32 = 3616712547
		x2 uint32 = 2333013068
		x3 uint32 = 2234311686
		x4 uint32 = 448751902
		x5 uint32 = 918748965
	)

	d := 20
	s := NewSketch(d)
	InitInverseTableUint32(d)
	s.AddSymbol(x1)
	s.AddSymbol(x2)
	s.AddSymbol(x3)

	// different orderings
	// NOTE: the spec of `decode_with_log` doesn't guarantee an order but
	// here we assume the elements appear in the same order as the list.
	fwd, succ := s.Decode([]HashType{x1, x2, x3})
	checkDecode(t, fwd, succ, []HashType{x1, x2, x3}, true)
	fwd, succ = s.Decode([]HashType{x3, x1, x2})
	checkDecode(t, fwd, succ, []HashType{x3, x1, x2}, true)

	// one extra element in log
	fwd, succ = s.Decode([]HashType{x1, x2, x3, x4})
	checkDecode(t, fwd, succ, []HashType{x1, x2, x3}, true)
	fwd, succ = s.Decode([]HashType{x1, x4, x2, x3})
	checkDecode(t, fwd, succ, []HashType{x1, x2, x3}, true)
	fwd, succ = s.Decode([]HashType{x4, x1, x2, x3})
	checkDecode(t, fwd, succ, []HashType{x1, x2, x3}, true)

	// two extra elements in log
	fwd, succ = s.Decode([]HashType{x1, x5, x2, x3, x4})
	checkDecode(t, fwd, succ, []HashType{x1, x2, x3}, true)

	// not all roots are in log
	fwd, succ = s.Decode([]HashType{x1, x2})
	checkDecode(t, fwd, succ, []HashType{x1, x2}, true)
	fwd, succ = s.Decode([]HashType{})
	checkDecode(t, fwd, succ, []HashType{}, true)
	fwd, succ = s.Decode([]HashType{x1, x2, x4})
	checkDecode(t, fwd, succ, []HashType{x1, x2}, true)
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
		// {"d=50000", 50000},
		// {"d=100000", 100000},
	}
	for _, tc := range cases {
		var nextId uint32

		d := tc.size
		n := tc.size
		log := make([]HashType, d + n)
		for i := 0; i < d + n; i++ {
			nextId += 1
			log[i] = nextId
		}
		slocal := NewSketch(d)
		sremote := NewSketch(d)
		for i := 0; i < d + n; i++ {
			slocal.AddSymbol(log[i])
		}
		for i := 0; i < n; i++ {
			sremote.AddSymbol(log[i])
		}
		InitInverseTableUint32(d)

		// Shuffle log
		for i := len(log) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			log[i], log[j] = log[j], log[i]
		}

		// Decode
		slocal.Subtract(sremote)
		missing, succ := slocal.Decode(log)
		if !succ {
			t.Errorf("(size=%d) failed to decode at all", tc.size)
		}
		if len(missing) != d {
			t.Errorf("(size=%d) missing symbols: %d local", tc.size, len(missing))
		}
	}
}
