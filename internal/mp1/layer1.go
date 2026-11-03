package mp1

import "math"

// ScaleFactors returns the Layer I scalefactor index per channel and band.
func ScaleFactors(subband [][][]float64) [][]int {
	channels := len(subband)
	out := make([][]int, channels)
	for ch := 0; ch < channels; ch++ {
		out[ch] = make([]int, 32)
		for band := 0; band < 32; band++ {
			var peak float64
			for slot := 0; slot < 12; slot++ {
				if v := math.Abs(subband[ch][slot][band]); v > peak {
					peak = v
				}
			}
			ires := 0
			for i := 62; i >= 0; i-- {
				if peak < scaleFactor[i] {
					ires = i
					break
				}
			}
			out[ch][band] = ires
		}
	}
	return out
}

// Allocate distributes bits over bands using the SNR/masking ratio.
func Allocate(smr [][]float64, maxBits int, totalBits *int, alloc [][]int) {
	channels := len(smr)
	rmnr := make([][]float64, channels)
	for ch := 0; ch < channels; ch++ {
		rmnr[ch] = make([]float64, 32)
		for band := 0; band < 32; band++ {
			rmnr[ch][band] = smr[ch][band] - snr[0]
		}
	}

	for {
		remaining := maxBits - *totalBits
		var band, channel int
		switch {
		case remaining < 12:
			return
		case remaining < 30:
			band, channel = maxLoc(rmnr, alloc, func(b int, bits int) bool {
				return bits < 13 && bits != 0
			})
		default:
			band, channel = maxLoc(rmnr, alloc, func(b int, bits int) bool {
				return bits < 13
			})
			if band >= 0 && rmnr[channel][band] < -0.44 {
				b2, c2 := maxLoc(rmnr, alloc, func(b int, bits int) bool {
					return bits == 0
				})
				if b2 >= 0 {
					band, channel = b2, c2
				} else {
					band, channel = maxLoc(rmnr, alloc, func(b int, bits int) bool {
						return bits < 13
					})
				}
			}
		}
		if band < 0 {
			return
		}

		k := alloc[channel][band]
		if k == 0 {
			*totalBits += 6 + 24
		} else {
			*totalBits += 12
		}
		alloc[channel][band] = k + 1
		rmnr[channel][band] = smr[channel][band] - snr[k+1]
	}
}

// maxLoc mirrors Fortran maxloc over rmnr(32, channels): band varies fastest.
func maxLoc(rmnr [][]float64, alloc [][]int, keep func(band, bits int) bool) (int, int) {
	bestBand, bestChannel := -1, -1
	best := 0.0
	for ch := range rmnr {
		for band := 0; band < 32; band++ {
			if !keep(band, alloc[ch][band]) {
				continue
			}
			if bestBand < 0 || rmnr[ch][band] > best {
				best = rmnr[ch][band]
				bestBand, bestChannel = band, ch
			}
		}
	}
	return bestBand, bestChannel
}

// Quantize converts subband samples to integers using the allocated bits.
func Quantize(alloc [][]int, subband [][][]float64, scale [][]int) [][][]int {
	channels := len(subband)
	out := make([][][]int, channels)
	for ch := 0; ch < channels; ch++ {
		out[ch] = make([][]int, 12)
		for slot := 0; slot < 12; slot++ {
			out[ch][slot] = make([]int, 32)
		}
		for band := 0; band < 32; band++ {
			k := alloc[ch][band]
			if k == 0 {
				continue
			}
			m := scale[ch][band]
			factor := math.Pow(2, float64(k))
			for slot := 0; slot < 12; slot++ {
				s := quantA[k]*subband[ch][slot][band]/scaleFactor[m] + quantB[k]
				out[ch][slot][band] = int(math.Floor((s + 1) * factor))
			}
		}
	}
	return out
}
