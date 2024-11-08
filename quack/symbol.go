package quack

type HashType = uint32
const HashTypeSize int64 = 4

type Symbol[T any] interface {
	Modulus() T
	AddAssign(rhs T)
	SubAssign(rhs T)
	MulAssign(rhs T)
	Pow(power T) T
	Neg() T
	Inv() T
	Add(rhs T) T
	Sub(rhs T) T
	Mul(rhs T) T
	Eq(rhs T) bool
}

type ModUint32 uint32

const ModulusUint32Small uint32 = 4294967291
const ModulusUint32Big uint64 = uint64(ModulusUint32Small)

var inverseTableUint32Threshold int
var InverseTableUint32 []ModUint32

func InitInverseTableUint32(d int) {
	InverseTableUint32 = make([]ModUint32, d)
	idx := ModUint32(1)
	for i := 0; i < d; i++ {
		InverseTableUint32[i] = idx.Inv()
		idx.AddAssign(ModUint32(1))
	}
}

func NewModUint32(n uint32) ModUint32 {
	if n > ModulusUint32Small {
		return ModUint32(n - ModulusUint32Small)
	} else {
		return ModUint32(n)
	}
}

func (lhs *ModUint32) AddAssign(rhs ModUint32) {
	sum := uint64(*lhs) + uint64(rhs)
	if sum >= ModulusUint32Big {
		*lhs = ModUint32(sum - ModulusUint32Big)
	} else {
		*lhs = ModUint32(sum)
	}
}

func (lhs *ModUint32) SubAssign(rhs ModUint32) {
	diff := uint64(*lhs) + uint64(rhs.Neg())
	if diff >= ModulusUint32Big {
		*lhs = ModUint32(diff - ModulusUint32Big)
	} else {
		*lhs = ModUint32(diff)
	}
}

func (lhs *ModUint32) MulAssign(rhs ModUint32) {
	prod := uint64(*lhs) * uint64(rhs)
	*lhs = ModUint32(prod % ModulusUint32Big)
}

func (x ModUint32) Pow(power ModUint32) ModUint32 {
	if power == 0 {
		return 1
	} else if power == 1 {
		return x
	} else {
		result := x.Pow(power >> 1)
		result.MulAssign(result)
		if power & 1 == 1 {
			result.MulAssign(x)
		}
		return result
	}
}

func (x ModUint32) Neg() ModUint32 {
	if x == 0 {
		return 0
	} else {
		return ModUint32(ModulusUint32Small) - x
	}
}

func (x ModUint32) Inv() ModUint32 {
	return x.Pow(ModUint32(ModulusUint32Big) - 2)
}

func (lhs ModUint32) Add(rhs ModUint32) ModUint32 {
	lhs.AddAssign(rhs)
	return lhs
}

func (lhs ModUint32) Sub(rhs ModUint32) ModUint32 {
	lhs.SubAssign(rhs)
	return lhs
}

func (lhs ModUint32) Mul(rhs ModUint32) ModUint32 {
	lhs.MulAssign(rhs)
	return lhs
}

func (lhs ModUint32) Eq(rhs ModUint32) bool {
	return lhs == rhs
}
