package mp1

import (
	"math"
	"math/cmplx"
	"testing"
)

func TestFFTMatchesNaiveDFT(t *testing.T) {
	p := &Psycho{}
	p.init(44100)

	x := make([]float64, fftSize)
	for i := range x {
		x[i] = math.Sin(0.01*float64(i)) + 0.5*math.Cos(0.03*float64(i))
	}

	got := p.fftWindow(x)

	for k := 0; k < 32; k++ {
		var sum complex128
		for n := 0; n < fftSize; n++ {
			angle := 2 * math.Pi * float64(k*n) / fftSize
			sum += complex(x[n]*p.hann[n]/fftSize, 0) * cmplx.Rect(1, angle)
		}
		if diff := cmplx.Abs(got[k] - sum); diff > 1e-9 {
			t.Fatalf("bin %d: diff %g (got %v, want %v)", k, diff, got[k], sum)
		}
	}
}
