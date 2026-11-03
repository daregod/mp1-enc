package mp1

import (
	"math"
	"math/cmplx"
)

const (
	fftOrder = 9
	fftSize  = 1 << fftOrder
	fftBins  = fftSize/2 + 1
)

// Psycho implements the psychoacoustic model (ISO layer I).
type Psycho struct {
	sampleRate int
	ready      bool
	freq       [fftBins]float64
	ath        [fftBins]float64
	crbw       [fftBins]float64
	cbwl       [fftBins]float64
	pseudo     [fftBins]float64
	sp         [fftBins][fftBins]float64
	scale      float64
	omega      [fftSize]complex128
	indx       [fftSize]int
	hann       [fftSize]float64
}

// Smr returns the signal-to-mask ratio per channel and subband.
func (p *Psycho) Smr(pcm [][]float64, sampleRate int) [][]float64 {
	if !p.ready || p.sampleRate != sampleRate {
		p.init(sampleRate)
	}

	smr := make([][]float64, len(pcm))
	for ch := range pcm {
		smr[ch] = make([]float64, 32)
		cfft := p.fftWindow(pcm[ch][160:672])
		smr[ch] = p.calcSMR(cfft)
	}
	return smr
}

func (p *Psycho) init(sampleRate int) {
	p.sampleRate = sampleRate
	p.ready = true

	twoPi := 2 * math.Pi
	for i := 0; i < fftSize; i++ {
		p.omega[i] = complex(math.Cos(twoPi*float64(i)/fftSize), math.Sin(twoPi*float64(i)/fftSize))
		p.hann[i] = 0.5 * (1 - math.Cos(twoPi*float64(i)/fftSize))
	}
	for i := 0; i < fftSize; i++ {
		n := 0
		k := i
		for j := fftOrder - 1; j >= 0; j-- {
			n += (k % 2) << uint(j)
			k /= 2
		}
		p.indx[i] = n
	}

	for i := 0; i < fftBins; i++ {
		freq := float64(sampleRate) / 2 / 1000 * float64(i) / float64(fftBins)
		p.ath[i] = 3.64*math.Pow(freq, -0.8) -
			6.5*math.Exp(-0.6*math.Pow(freq-3.3, 2)) +
			0.001*math.Pow(freq, 4)
		p.freq[i] = freq * 1000
	}
	p.scale = float64(sampleRate) / 2 / 256
	for i := 0; i < fftBins; i++ {
		p.crbw[i] = 25 + 75*math.Pow(1+1.4*math.Pow(p.freq[i]/1000, 2), 0.69)
		p.cbwl[i] = decibel(p.crbw[i])
	}

	var tmp float64
	for m := 1; m < fftBins; m++ {
		tmp += 1 / p.crbw[m]
		p.pseudo[m] = tmp * p.scale
	}
	for i := 1; i < fftBins; i++ {
		for m := 1; m < fftBins; m++ {
			p.sp[i][m] = spreading(p.pseudo[i] - p.pseudo[m])
		}
	}
}

func spreading(z float64) float64 {
	var v float64
	if z > 0 {
		v = -25 * z
	} else {
		v = 75 * z
	}
	return math.Max(-160, v)
}

func decibel(x float64) float64 {
	return 10 * math.Log10(math.Max(1e-100, x))
}

func pow10(x float64) float64 { return math.Pow(10, x) }

func (p *Psycho) fftWindow(x []float64) []complex128 {
	y := make([]complex128, fftSize)
	for i := 0; i < fftSize; i++ {
		y[i] = complex(p.hann[i]*x[i]/fftSize, 0)
	}
	p.fft(y)
	return y
}

func (p *Psycho) fft(y []complex128) {
	tmp := make([]complex128, fftSize)
	for i := 0; i < fftSize; i++ {
		tmp[i] = y[p.indx[i]]
	}
	copy(y, tmp)

	for k2 := 1; k2 <= fftOrder; k2++ {
		kn2 := 1 << uint(k2-1)
		for i := 0; i < fftSize; i += 2 * kn2 {
			for j := 0; j < kn2; j++ {
				iphase := (1 << uint(fftOrder-k2)) * j
				c1 := p.omega[iphase%fftSize]
				m1 := i + j
				m2 := m1 + kn2
				t1 := y[m1]
				t2 := c1 * y[m2]
				y[m1] = t1 + t2
				y[m2] = t1 - t2
			}
		}
	}
}

func (p *Psycho) calcSMR(cfft []complex128) []float64 {
	const alpha = 0.27

	var xa, ya, za [fftBins]float64
	for i := 0; i < fftBins; i++ {
		xa[i] = 2 * decibel(cmplx.Abs(cfft[i]))
	}

	for i := 1; i < fftBins; i++ {
		var sum float64
		for m := 1; m < fftBins; m++ {
			sum += pow10(((p.sp[i][m]+xa[m]-p.ath[m])*alpha - p.cbwl[m]) / 10)
		}
		ya[i] = math.Max(decibel(sum*p.scale)/alpha-11.5, p.ath[i]-90.3)
	}

	for i := 1; i < fftBins; i++ {
		m := int(math.Round(p.crbw[i] / p.scale))
		i0 := maxInt(i-m/2, 1)
		i1 := minInt(i0+m-1, fftBins-1)
		var sum float64
		for j := i0; j <= i1; j++ {
			sum += pow10((xa[j]*alpha - p.cbwl[i]) / 10)
		}
		za[i] = decibel(sum*p.scale) / alpha
	}

	smr := make([]float64, 32)
	for band := 1; band <= 32; band++ {
		m := (band-1)*8 + 1
		i0 := maxInt(m-int(math.Round(p.crbw[m]/2/p.scale)), 1)
		m2 := m + 7
		i1 := minInt(m2+int(math.Round(p.crbw[m2]/2/p.scale)), fftBins-1)

		snrMax := za[i0]
		mnrMin := ya[i0]
		for i := i0 + 1; i <= i1; i++ {
			if za[i] > snrMax {
				snrMax = za[i]
			}
			if ya[i] < mnrMin {
				mnrMin = ya[i]
			}
		}
		smr[band-1] = snrMax - mnrMin
	}
	return smr
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
