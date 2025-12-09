package utils

import (
	"math"

	"gonum.org/v1/gonum/dsp/fourier"
)

func MAV(x []int) float64 {
	sum := 0.0
	for _, v := range x {
		sum += math.Abs(float64(v))
	}
	return sum / float64(len(x))
}

func RMS(x []int) float64 {
	sum := 0.0
	for _, v := range x {
		f := float64(v)
		sum += f * f
	}
	return math.Sqrt(sum / float64(len(x)))
}

func WL(x []int) float64 {
	sum := 0.0
	for i := 1; i < len(x); i++ {
		sum += math.Abs(float64(x[i] - x[i-1]))
	}
	return sum
}

func VAR(x []int) float64 {
	mean := 0.0
	for _, v := range x {
		mean += float64(v)
	}
	mean /= float64(len(x))

	variance := 0.0
	for _, v := range x {
		diff := float64(v) - mean
		variance += diff * diff
	}
	return variance / float64(len(x))
}

func ZeroCross(x []int) float64 {
	count := 0.0
	for i := 1; i < len(x); i++ {
		if float64(x[i])*float64(x[i-1]) < 0 {
			count++
		}
	}
	return count
}

func SSC(x []int) float64 {
	count := 0.0
	for i := 1; i < len(x)-1; i++ {
		a := float64(x[i-1])
		b := float64(x[i])
		c := float64(x[i+1])
		if (b-a)*(b-c) > 0 {
			count++
		}
	}
	return count
}

func Max(x []int) float64 {
	m := x[0]
	for _, v := range x {
		if v > m {
			m = v
		}
	}
	return float64(m)
}

func Min(x []int) float64 {
	m := x[0]
	for _, v := range x {
		if v < m {
			m = v
		}
	}
	return float64(m)
}

func IEMG(x []int) float64 {
	sum := 0.0
	for _, v := range x {
		sum += math.Abs(float64(v))
	}
	return sum
}

func KF(x []int) float64 {
	sumSq := 0.0
	for _, v := range x {
		f := float64(v)
		sumSq += f * f
	}
	return math.Sqrt(sumSq) / float64(len(x))
}

func MeanFreq(x []int) float64 {
	N := len(x)
	rfft := fourier.NewFFT(N).Coefficients(nil, intsToFloat64(x))
	mags := make([]float64, N/2+1)

	sum := 0.0
	for i := 0; i < len(mags); i++ {
		mags[i] = cmplxAbs(rfft[i])
		sum += mags[i]
	}
	return sum / float64(len(mags))
}

func PeakFreq(x []int) float64 {
	N := len(x)
	rfft := fourier.NewFFT(N).Coefficients(nil, intsToFloat64(x))
	maxIdx := 0
	maxVal := 0.0

	for i := 0; i < N/2+1; i++ {
		v := cmplxAbs(rfft[i])
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return float64(maxIdx)
}

// utils

func intsToFloat64(x []int) []float64 {
	out := make([]float64, len(x))
	for i, v := range x {
		out[i] = float64(v)
	}
	return out
}

func cmplxAbs(c complex128) float64 {
	return math.Sqrt(real(c)*real(c) + imag(c)*imag(c))
}

func ExtractFeatures(raw []int) []float64 {
	return []float64{
		MAV(raw),
		RMS(raw),
		WL(raw),
		VAR(raw),
		ZeroCross(raw),
		SSC(raw),
		Max(raw),
		Min(raw),
		IEMG(raw),
		KF(raw),
		MeanFreq(raw),
		PeakFreq(raw),
	}
}
