package helperFunctions

import (
	"errors"
	"fmt"
	"math"
)

func EstimateYValueInHyperbolaFunction(x float64, fit []float64) float64 {
	y := fit[0] + x * fit[1]
	return InvertValue(y)
}

func EstimateXValueInHyperbolaFunction(y float64, fit []float64) float64 {
	x := (InvertValue(y) - fit[0]) / fit[1]
	return x
}

func HyperbolaLeastSquares(x []float64, y []float64) []float64 {
	invertedXs := InvertValues(x)

	fit := LinearLeastSquares(invertedXs, y)

	return fit
}

func InvertValues(values []float64) []float64 {
	invertedValues := make([]float64, len(values))

	for i, f := range values {
		invertedValues[i] = InvertValue(f)
	}

	return invertedValues
}

func InvertValue(f float64) float64 {
	if f == 0 {
		println("trying to invert 0 value, will result in +inf")
		return math.MaxFloat64 / 2.0
	}
	return 1 / f
}

func LinearLeastSquares(x []float64, y []float64) []float64 {
	r := len(x)
	c := len(y)

	if r != c {
		FatalErrCheck(errors.New("rows != columns"), "LinearLeastSquares: ")
	}

	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumXX := 0.0

	for i := 0; i < len(x); i++ {
		xVal := x[i]
		yVal := y[i]

		sumX += xVal
		sumY += yVal
		sumXY += xVal * yVal
		sumXX += xVal * xVal
	}

	base := float64(r)*sumXX - sumX*sumX
	x1 := (float64(r)*sumXY - sumX*sumY) / base
	x0 := (sumXX*sumY - sumXY*sumX) / base

	res := []float64{x0, x1}

	return res
}

func EstimateYValueInFunction(x float64, f []float64) float64 {
	res := 0.0
	for i := 0; i < len(f); i++ {
		res += f[i] * math.Pow(x, float64(i))
	}
	return res
}

func PolynomialLeastSquares(x []float64, y []float64) []float64 {
	r := len(x)
	c := len(y)

	fmt.Printf("len(x): %d len(y): %d\n", r, c)

	if r != c {
		FatalErrCheck(errors.New("rows != columns"), "PolynomialLeastSquares: ")
	}

	const polDegConst = 2
	n := polDegConst

	var X [2*polDegConst + 1]float64

	for i := 0; i < 2*n+1; i++ {
		X[i] = 0
		for j := 0; j < r; j++ {
			X[i] = X[i] + math.Pow(x[j], float64(i))
		}
	}

	var B [polDegConst + 1][polDegConst + 2]float64

	for i := 0; i <= n; i++ {
		for j := 0; j <= n; j++ {
			B[i][j] = X[i+j]
		}
	}

	var Y [polDegConst + 1]float64

	for i := 0; i < n+1; i++ {
		Y[i] = 0
		for j := 0; j < r; j++ {
			Y[i] = Y[i] + math.Pow(x[j], float64(i))*y[j]
		}
	}

	a := make([]float64, polDegConst + 1)

	for i := 0; i <= n; i++ {
		B[i][n+1] = Y[i]
	}

	n = n + 1
	for i := 0; i < n; i++ {
		for k := 0; k < n; k++ {
			if B[i][i] < B[k][i] {
				for j := 0; j <= n; j++ {
					temp := B[i][j]
					B[i][j] = B[k][j]
					B[k][j] = temp
				}
			}
		}
	}

	for i := 0; i < n-1; i++ {
		for k := i + 1; k < n; k++ {
			t := B[k][i] / B[i][i]
			for j := 0; j <= n; j++ {
				B[k][j] = B[k][j] - t*B[i][j]
			}
		}
	}

	for i := n - 1; i >= 0; i-- {
		a[i] = B[i][n]
		for j := 0; j < n; j++ {
			if j != i {
				a[i] = a[i] - B[i][j]*a[j]
			}
		}
		a[i] = a[i] / B[i][i]
	}

	return a
}
