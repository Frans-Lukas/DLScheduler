package helperFunctions

import (
	"errors"
	"math"
)

func PolynomialLeastSquares(x []float64, y []float64) [3]float64 {
	r := len(x)
	c := len(y)

	if r != c {
		FatalErrCheck(errors.New("rows != columns"), "polynomialLeastSquares: ")
	}

	const polDegConst = 2
	n := polDegConst

	var X [2 * polDegConst + 1]float64

	for i := 0 ; i < 2*n+1 ; i++ {
		X[i]=0
		for j := 0 ; j < r ; j++ {
			X[i]=X[i]+math.Pow(x[j],float64(i))
		}
	}

	var B [polDegConst+1][polDegConst+2]float64

	for i := 0 ; i <= n ; i++ {
		for j := 0 ; j <= n ; j++ {
			B[i][j] = X[i+j]
		}
	}

	var Y [polDegConst + 1]float64

	for i := 0 ; i < n+1 ; i++ {
		Y[i] = 0
		for j := 0 ; j < r ; j++ {
			Y[i] = Y[i] + math.Pow(x[j], float64(i))*y[j]
		}
	}

	var a [polDegConst+1]float64

	for i := 0 ; i <= n ; i++ {
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
			t := B[k][i]/ B[i][i]
			for j := 0; j <= n; j++ {
				B[k][j]= B[k][j]-t*B[i][j]
			}
		}
	}

	for i := n -1 ; i >= 0 ; i-- {
		a[i]= B[i][n]
		for j := 0 ; j < n ; j++ {
			if j != i {
				a[i]=a[i]-B[i][j]*a[j]
			}
		}
		a[i]=a[i]/B[i][i]
	}

	return a
}