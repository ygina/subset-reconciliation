package riblt

// Decoder computes the symmetric difference between two sets A, B. The Decoder
// knows B (the local set) and expects coded symbols for A (the remote set). 
type Decoder struct {
	// coded symbols received so far
	cs []CodedSymbol
	// set of source symbols that are exclusive to the decoder
	local codingWindow
	// set of source symbols that the decoder initially has
	window codingWindow
	// set of source symbols that are exclusive to the encoder
	remote codingWindow
	// indices of coded symbols that can be decoded, i.e., degree equal to -1
	// or 1 or degree equal to 0 and sum of hash equal to 0
	decodable []int
	// number of coded symbols that are decoded
	decoded int
}

// Decoded returns true if and only if every existing coded symbols d received
// so far have been decoded.
func (d *Decoder) Decoded() bool {
	return d.decoded == len(d.cs)
}

// Local returns the list of source symbols that are present in B but not in A.
func (d *Decoder) Local() []HashType {
	return d.local.symbols
}

// Remote returns the list of source symbols that are present in A but not in B.
func (d *Decoder) Remote() []HashType {
	return d.remote.symbols
}

// AddSymbol adds a source symbol to B, the Decoder's local set. It is
// undefined behavior to call AddSymbol after AddCodedSymbol has been called
// one or multiple times.
func (d *Decoder) AddSymbol(s HashType) {
	d.AddHash(s)
}

// AddHash adds a source symbol to B, the Decoder's local set. It is
// undefined behavior to call AddHash after AddCodedSymbol has been
// called one or multiple times.
func (d *Decoder) AddHash(s HashType) {
	d.window.addHash(s)
}

// AddCodedSymbol passes the next coded symbol in A's sequence to the Decoder.
// Coded symbols must be passed in the same ordering as they are generated by
// A's Encoder.
func (d *Decoder) AddCodedSymbol(c CodedSymbol) {
	// scan through decoded symbols to peel off matching ones
	c = d.window.applyWindow(c, remove)
	c = d.remote.applyWindow(c, remove)
	c = d.local.applyWindow(c, add)
	// insert the new coded symbol
	d.cs = append(d.cs, c)
	// check if the coded symbol is decodable, and insert into decodable list if so
	if c.Count == 1 || c.Count == -1 {
		d.decodable = append(d.decodable, len(d.cs)-1)
	} else if c.Count == 0 && c.Hash == 0 {
		d.decodable = append(d.decodable, len(d.cs)-1)
	}
	return
}

func (d *Decoder) applyNewSymbol(t HashType, direction int64) randomMapping {
	m := randomMapping{t, 0}
	for int(m.lastIdx) < len(d.cs) {
		cidx := int(m.lastIdx)
		d.cs[cidx] = d.cs[cidx].apply(t, direction)
		// Check if the coded symbol is now decodable. We do not want to insert
		// a decodable symbol into the list if we already did, otherwise we
		// will visit the same coded symbol twice. To see how we achieve that,
		// notice the following invariant: if a coded symbol becomes decodable
		// with degree D (obviously -1 <= D <=1), it will stay that way, except
		// for that it's degree may become 0. For example, a decodable symbol
		// of degree -1 may not later become undecodable, or become decodable
		// but of degree 1. This is because each peeling removes a source
		// symbol from the coded symbol. So, if a coded symbol already contains
		// only 1 or 0 source symbol (the definition of decodable), the most we
		// can do is to peel off the only remaining source symbol.
		//
		// Meanwhile, notice that if a decodable symbol is of degree 0, then
		// there must be a point in the past when it was of degree 1 or -1 and
		// decodable, at which time we would have inserted it into the
		// decodable list. So, we do not insert degree-0 symbols to avoid
		// duplicates. On the other hand, it is fine that we insert all
		// degree-1 or -1 decodable symbols, because we only see them in such
		// state once.
		if d.cs[cidx].Count == -1 || d.cs[cidx].Count == 1 {
			d.decodable = append(d.decodable, cidx)
		}
		m.nextIndex()
	}
	return m
}

// TryDecode tries to decode all coded symbols received so far.
func (d *Decoder) TryDecode() {
	for didx := 0; didx < len(d.decodable); didx += 1 {
		cidx := d.decodable[didx]
		c := d.cs[cidx]
		// We do not need to compare Hash and Symbol.Hash() below, because we
		// have checked it before inserting into the decodable list. Per the
		// invariant mentioned in the comments in applyNewSymbol, a decodable
		// symbol does not turn undecodable, so there is no worry that
		// additional source symbols have been peeled off a coded symbol after
		// it was inserted into the decodable list and before we visit them
		// here.
		switch c.Count {
		case 1:
			// allocate a symbol and then XOR with the sum, so that we are
			// guaranted to copy the sum whether or not the symbol interface is
			// implemented as a pointer
			ns := c.Hash
			m := d.applyNewSymbol(ns, remove)
			d.remote.addHashWithMapping(ns, m)
			d.decoded += 1
		case -1:
			panic("only handle subset reconciliation")
		case 0:
			d.decoded += 1
		default:
			// a decodable symbol does not turn undecodable, so its degree must
			// be -1, 0, or 1
			panic("invalid degree for decodable coded symbol")
		}
	}
	d.decodable = d.decodable[:0]
}

// Reset clears d. It is more efficient to call Reset to reuse an existing
// Decoder than creating a new one.
func (d *Decoder) Reset() {
	if len(d.cs) != 0 {
		d.cs = d.cs[:0]
	}
	if len(d.decodable) != 0 {
		d.decodable = d.decodable[:0]
	}
	d.local.reset()
	d.remote.reset()
	d.window.reset()
	d.decoded = 0
}
