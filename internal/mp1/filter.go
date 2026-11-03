package mp1

import "math"

// Polyphase implements the ISO analysis subband filter.
type Polyphase struct {
	Channels int
	pcm      [][]float64
	subband  [][][]float64 // [channel][slot][band]
	coeff    [32][64]float64
}

// NewPolyphase creates the filter bank for the given channel count.
func NewPolyphase(channels int) *Polyphase {
	p := &Polyphase{Channels: channels}
	const pi = math.Pi
	for i := 0; i < 32; i++ {
		for j := 0; j < 64; j++ {
			arg := float64(((2*i+1)*(j-16))%128) * pi / 64
			p.coeff[i][j] = round(math.Cos(arg), 9)
		}
	}
	p.pcm = make([][]float64, channels)
	p.subband = make([][][]float64, channels)
	for ch := 0; ch < channels; ch++ {
		p.pcm[ch] = make([]float64, 864)
		p.subband[ch] = make([][]float64, 12)
		for i := 0; i < 12; i++ {
			p.subband[ch][i] = make([]float64, 32)
		}
	}
	return p
}

// PushFrame appends 384 interleaved samples to the analysis buffer.
func (p *Polyphase) PushFrame(samples []int16) {
	for ch := 0; ch < p.Channels; ch++ {
		buf := p.pcm[ch]
		copy(buf[384:864], buf[:480])
		for i := 0; i < 384; i++ {
			buf[383-i] = float64(samples[i*p.Channels+ch]) / 32768
		}
	}
}

// Filter computes the 12 subband samples of the current frame.
func (p *Polyphase) Filter() {
	var z [512]float64
	var y [64]float64
	for i := 0; i < 12; i++ {
		end := 862 - 32*i
		start := end - 511
		for ch := 0; ch < p.Channels; ch++ {
			buf := p.pcm[ch]
			for k := 0; k < 512; k++ {
				idx := start + k
				if idx < 0 {
					z[k] = 0
					continue
				}
				z[k] = prototype[k] * buf[idx]
			}
			for j := 0; j < 64; j++ {
				var sum float64
				for m := 0; m < 8; m++ {
					sum += z[j+64*m]
				}
				y[j] = sum
			}
			for band := 0; band < 32; band++ {
				var sum float64
				for j := 0; j < 64; j++ {
					sum += p.coeff[band][j] * y[j]
				}
				p.subband[ch][i][band] = sum
			}
		}
	}
}

// Subband returns the subband samples of the last frame: [channel][slot][band].
func (p *Polyphase) Subband() [][][]float64 { return p.subband }

// PCM returns the internal 864-sample analysis buffer per channel.
func (p *Polyphase) PCM() [][]float64 { return p.pcm }
