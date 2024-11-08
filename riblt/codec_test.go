package riblt

import (
	"encoding/binary"
	"github.com/dchest/siphash"
	"testing"
	"unsafe"
)

const testSymbolSize = 64

type testSymbol [testSymbolSize]byte

func (d testSymbol) XOR(t2 testSymbol) testSymbol {
	dw := (*[testSymbolSize / 8]uint64)(unsafe.Pointer(&d))
	t2w := (*[testSymbolSize / 8]uint64)(unsafe.Pointer(&t2))
	for i := 0; i < testSymbolSize/8; i++ {
		(*dw)[i] ^= (*t2w)[i]
	}
	return d
}

func (d testSymbol) Hash() HashType {
	hash64 := siphash.Hash(567, 890, d[:])
	hash32 := uint32(hash64 & 0xFFFFFFFF)
	return hash32
}

func newTestSymbol(i uint64) testSymbol {
	data := testSymbol{}
	binary.LittleEndian.PutUint64(data[0:8], i)
	return data
}

func BenchmarkEncodeAndDecode(bc *testing.B) {
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
		bc.Run(tc.name, func(b *testing.B) {
			b.SetBytes(testSymbolSize * int64(tc.size))
			nlocal := 0
			nremote := tc.size
			ncommon := tc.size
			ncw := 0
			var nextId uint64
			b.ResetTimer()
			b.StopTimer()
			for iter := 0; iter < b.N; iter++ {
				enc := Encoder{}
				dec := Decoder{}

				for i := 0; i < nlocal; i++ {
					s := newTestSymbol(nextId)
					nextId += 1
					dec.AddSymbol(s.Hash())
				}
				for i := 0; i < nremote; i++ {
					s := newTestSymbol(nextId)
					nextId += 1
					enc.AddSymbol(s.Hash())
				}
				for i := 0; i < ncommon; i++ {
					s := newTestSymbol(nextId)
					nextId += 1
					enc.AddSymbol(s.Hash())
					dec.AddSymbol(s.Hash())
				}
				b.StartTimer()
				for {
					dec.AddCodedSymbol(enc.ProduceNextCodedSymbol())
					dec.TryDecode()
					ncw += 1
					if dec.Decoded() {
						break
					}
				}
				b.StopTimer()
			}
			b.ReportMetric(float64(ncw)/float64(b.N * tc.size), "symbols/diff")
		})
	}
}

func TestEncodeAndDecode(t *testing.T) {
	enc := Encoder{}
	dec := Decoder{}
	local := make(map[HashType]struct{})
	remote := make(map[HashType]struct{})

	var nextId uint64
	nlocal := 0
	nremote := 1000
	ncommon := 1000
	for i := 0; i < nlocal; i++ {
		s := newTestSymbol(nextId)
		nextId += 1
		dec.AddSymbol(s.Hash())
		local[s.Hash()] = struct{}{}
	}
	for i := 0; i < nremote; i++ {
		s := newTestSymbol(nextId)
		nextId += 1
		enc.AddSymbol(s.Hash())
		remote[s.Hash()] = struct{}{}
	}
	for i := 0; i < ncommon; i++ {
		s := newTestSymbol(nextId)
		nextId += 1
		enc.AddSymbol(s.Hash())
		dec.AddSymbol(s.Hash())
	}

	ncw := 0
	for {
		dec.AddCodedSymbol(enc.ProduceNextCodedSymbol())
		ncw += 1
		dec.TryDecode()
		if dec.Decoded() {
			break
		}
		if ncw % 100000 == 0 {
			t.Errorf("%d coded symbols, %d remote %d local", ncw, len(dec.Remote()), len(dec.Local()))
		}
	}
	for _, v := range dec.Remote() {
		delete(remote, v)
	}
	for _, v := range dec.Local() {
		delete(local, v)
	}
	if len(remote) != 0 || len(local) != 0 {
		t.Errorf("missing symbols: %d remote and %d local", len(remote), len(local))
	}
	if !dec.Decoded() {
		t.Errorf("decoder not marked as decoded")
	}
}

