package helperFunctions

import (
	"errors"
)

func polynomialLeastSquares(matrix [][]float64) [3]float64 {
	r := len(matrix)
	c := len(matrix[0])

	if r == c {
		FatalErrCheck(errors.New("rows != columns"), "polynomialLeastSquares: ")
	}

	const polDegConst = 2
	n := polDegConst

	var B [polDegConst+1][polDegConst+2]float64
	var a [polDegConst+1]float64

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