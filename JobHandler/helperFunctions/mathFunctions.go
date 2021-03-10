package helperFunctions

import (
	"errors"
	"fmt"
	"math"
)

func EstimateValueInHyperbola(x float64, fit []float64) float64 {
	y := fit[0] + x*fit[1]
	return invertValue(y)
}

func HyperbolaLeastSquares(x []float64, y []float64) []float64 {
	invertedXs := invertValues(x)

	fit := LinearLeastSquares(invertedXs, y)

	return fit
}

func invertValues(values []float64) []float64 {
	for i, f := range values {
		values[i] = invertValue(f)
	}

	return values
}

func invertValue(f float64) float64 {
	return 1 / f
}

func LinearLeastSquares(x []float64, y []float64) []float64 {
	r := len(x)
	c := len(y)
	fmt.Printf("r: %d, c: %d", r, c)

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

		fmt.Printf("x: %f, y: %f", x[i], y[i])

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

func PerformEstimationWithFunctions(f [3]float64, x float64) float64 {
	return f[0] + f[1]*x + f[2]*(x*x)
}

func PolynomialLeastSquares(x []float64, y []float64) [3]float64 {
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

	var a [polDegConst + 1]float64

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
