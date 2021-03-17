package helperFunctions

import (
	"errors"
	"fmt"
	"jobHandler/constants"
	"math"
	"regexp"
	"strconv"
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

//example input []float64{-1., -0.89473684, -0.78947368, -0.68421053, -0.57894737, -0.47368421, -0.36842105, -0.26315789, -0.15789474, -0.05263158, 0.05263158, 0.15789474, 0.26315789, 0.36842105, 0.47368421, 0.57894737, 0.68421053, 0.78947368, 0.89473684, 1.}, []float64{-1., -0.89473684, -0.78947368, -0.68421053, -0.57894737, -0.47368421, -0.36842105, -0.26315789, -0.15789474, -0.05263158, 0.05263158, 0.15789474, 0.26315789, 0.36842105, 0.47368421, 0.57894737, 0.68421053, 0.78947368, 0.89473684, 1.}, []float64{2.655, 2.09876731, 1.60901662, 1.18574792, 0.82896122, 0.53865651, 0.3148338, 0.15749307, 0.06663435, 0.04225762, 0.08436288, 0.19295014, 0.36801939, 0.60957064, 0.91760388, 1.29211911, 1.73311634, 2.24059557, 2.81455679, 3.455}, []float64{0, 0, 1, 2}
func Python3DParabolaLeastSquares(xs []float64, ys []float64, hs []float64, initialGuess []float64) []float64 {
	var xString string
	for _, x := range xs {
		tmp := fmt.Sprintf("%f", x)
		if len(xString) == 0 {
			xString = tmp
		} else {
			xString = xString + " " + tmp
		}
	}

	var yString string
	for _, y := range ys {
		tmp := fmt.Sprintf("%f", y)
		if len(yString) == 0 {
			yString = tmp
		} else {
			yString = yString + " " + tmp
		}
	}

	var hString string
	for _, h := range hs {
		tmp := fmt.Sprintf("%f", h)
		if len(hString) == 0 {
			hString = tmp
		} else {
			hString = hString + " " + tmp
		}
	}

	var guess string
	for _, g := range initialGuess {
		tmp := fmt.Sprintf("%f", g)
		if len(guess) == 0 {
			guess = tmp
		} else {
			guess = guess + " " + tmp
		}
	}

	out, stderr, err := ExecuteFunction(constants.PYTHON, constants.PYTHON_LEAST_SQUARES, xString, yString, hString, guess)

	FatalErrCheck(err, "Python3DParabolaLeastSquares: " + stderr.String())

	outString := out.String()

	re := regexp.MustCompile("\\[")
	splitString := re.Split(outString, -1)

	re = regexp.MustCompile("]")
	splitString = re.Split(splitString[1], -1)

	re = regexp.MustCompile("[ ]+")
	splitString = re.Split(splitString[0], -1)

	output := make([]float64, 4)

	currPos := 0
	for _, s := range splitString {
		tmp, err := strconv.ParseFloat(s, 64)
		if err == nil {
			output[currPos] = tmp
			currPos++
		}
	}

	if currPos != 4 {
		FatalErrCheck(errors.New("too few variables returned, need 4 got: " + outString), "Python3DParabolaLeastSquares: ")
	}

	return output
}

func Python3DParabolaLeastSquaresEstimateH(x float64, y float64, f []float64) float64 {
	return f[2] * math.Pow(x - f[0], 2) + f[3] * math.Pow(y - f[1], 2)
}
