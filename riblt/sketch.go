package riblt

// Sketch is a prefix of the coded symbol sequence for a set of source symbols.
// When generating a prefix of predetermined length, compared to generating the
// prefix incrementally using an Encoder, it is more efficient to use Sketch.
// Sketch also allows inserting or deleting source symbols from the set after
// it has been created.
type Sketch []CodedSymbol

// AddSymbol inserts source symbol t to the set of which s is a sketch.
func (s Sketch) AddSymbol(t HashType) {
	m := randomMapping{t, 0}
	for int(m.lastIdx) < len(s) {
		idx := m.lastIdx
		s[idx].Count += 1
		s[idx].Hash ^= t
		m.nextIndex()
	}
}

// RemoveSymbol deletes source symbol t from the set of which s is a sketch.
func (s Sketch) RemoveSymbol(t HashType) {
	m := randomMapping{t, 0}
	for int(m.lastIdx) < len(s) {
		idx := m.lastIdx
		s[idx].Count -= 1
		s[idx].Hash ^= t
		m.nextIndex()
	}
}

// Subtract subtracts s2 from s by modifying s in place. s and s2 must be of
// equal length. If s is a sketch of set S and s2 is a sketch of set S2, then
// the result is a sketch of the symmetric difference between S and S2.
func (s Sketch) Subtract(s2 Sketch) {
	if len(s) != len(s2) {
		panic("subtracting sketches of different sizes")
	}

	for i := range s {
		s[i].Count = s[i].Count - s2[i].Count
		s[i].Hash ^= s2[i].Hash
	}
	return
}

// Decode tries to decode s, where s can be one of the following
//  1. A sketch of set S.
//  2. Content of s after calling s.Subtract(s2), where s is a sketch of set
//     S, and s2 is a sketch of set S2.
//
// When successful, indicated by succ being true, fwd contains all source
// symbols in S in case 1, or S \ S2 in case 2 (\ is the set subtraction
// operation). rev is empty in case 1, or S2 \ S in case 2.
func (s Sketch) Decode() (fwd []HashType, rev []HashType, succ bool) {
	dec := Decoder{}
	for _, c := range s {
		dec.AddCodedSymbol(c)
	}
	dec.TryDecode()
	return dec.Remote(), dec.Local(), dec.Decoded()
}
