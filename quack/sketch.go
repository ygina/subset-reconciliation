package quack

type Sketch struct {
	PowerSums []ModUint32
	Count     uint32
}

func NewSketch(d int) Sketch {
	return Sketch {
		PowerSums: make([]ModUint32, d),
		Count: 0,
	}
}

// AddSymbol inserts source symbol t to the set of which s is a sketch.
func (s *Sketch) AddSymbol(t HashType) {
	size := len(s.PowerSums)
	x := NewModUint32(t)
	y := x
	for i := 0; i < size - 1; i++ {
		s.PowerSums[i].AddAssign(y)
		y.MulAssign(x)
	}
	s.PowerSums[size - 1].AddAssign(y)
	s.Count += 1
}

// Subtract subtracts s2 from s by modifying s in place. s and s2 must be of
// equal length. If s is a sketch of set S and s2 is a sketch of set S2, then
// the result is a sketch of the symmetric difference between S and S2.
func (s *Sketch) Subtract(s2 Sketch) {
	if len(s.PowerSums) != len(s2.PowerSums) {
		panic("subtracting sketches of different sizes")
	}

	s.Count -= s2.Count
	for i := range s.PowerSums {
		s.PowerSums[i].SubAssign(s2.PowerSums[i])
	}
	return
}

func (s Sketch) ToCoeffs() []ModUint32 {
	coeffs := make([]ModUint32, s.Count)
	coeffs[0] = s.PowerSums[0].Neg()
	for i := 1; i < len(coeffs); i++ {
		coeffs[i] = ModUint32(0)
		for j := 0; j < i; j++ {
			coeffs[i] = coeffs[i].Sub(s.PowerSums[j].Mul(coeffs[i - j - 1]))
		}
		coeffs[i].SubAssign(s.PowerSums[i])
		coeffs[i].MulAssign(InverseTableUint32[i])
	}
	return coeffs
}

func EvalCoeffs(coeffs []ModUint32, x ModUint32) ModUint32 {
	size := len(coeffs)
	result := x
	for i := 0; i < size - 1; i++ {
		result.AddAssign(coeffs[i])
		result.MulAssign(x)
	}
	return result.Add(coeffs[size - 1])
}

// Decode tries to decode s, where s can be one of the following
//  1. A sketch of set S.
//  2. Content of s after calling s.Subtract(s2), where s is a sketch of set
//     S, and s2 is a sketch of set S2.
//
// When successful, indicated by succ being true, fwd contains all source
// symbols in S in case 1, or S \ S2 in case 2 (\ is the set subtraction
// operation). rev is empty in case 1, or S2 \ S in case 2.
func (s Sketch) Decode(log []HashType) (missing []HashType, succ bool) {
	if s.Count == 0 {
		return []HashType{}, true
	}
	if int(s.Count) > len(s.PowerSums) {
		// panic("number of elements must not exceed threshold")
		return []HashType{}, false
	}

	missing = []HashType{}
	coeffs := s.ToCoeffs()
	for _, x := range log {
		if EvalCoeffs(coeffs, NewModUint32(x)) == 0 {
			missing = append(missing, x)
		}
	}

	return missing, true
}
