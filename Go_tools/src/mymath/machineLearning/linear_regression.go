package machineLearning

import (
	"log"
	"math"
	"mymath"
)

//Linear_Regression takes 4/5 parameters.
//1.  x ---- 2-d slice, representing "feature"
//2.  y ---- 1-d slice, representing "target"
//3.  initial ---- initial value for intrested parameters
//4.  alpha ---- set up how long you want to "walk" for 1 step
//5.  iteration ---- maximux iteration, which is used under conditions that without convergence
func Linear_Regression(x [][]float64, y, initial []float64, alpha float64, iteration ...int64) []float64 {

	if len(x) != len(y) {
		log.Fatal("dimensions of x, y NOT the same!")
		return nil
	}

	var count int64
	var recursion int64
	if len(iteration) != 0 {
		recursion = iteration[0]
	} else {
		recursion = 10000
	}

	m := len(x)
	n := len(x[0])

	if alpha == 0.0 {
		alpha = 0.05
	} //default value

	h := func(xj []float64) float64 {
		sum := 0.0
		for i := 0; i != n; i++ {
			sum += initial[i+1] * xj[i]
		}
		return sum + initial[0]
	}

	c := make(chan float64)

	calc_next := func(initial_inner []float64) []float64 {
		next := make([]float64, n+1)
		for i := 1; i != n+1; i++ {
			go func(c chan float64) {
				sum := 0.0
				for j := 0; j != m; j++ {
					sum += (h(x[j]) - y[j]) * x[j][i-1]
				}
				c <- sum
			}(c)
			next[i] = initial_inner[i] - alpha*<-c
			next[0] = initial_inner[0] - alpha*(h(x[0])-y[0])
		}
		return next
	}

	initial_matrix := &mymath.Matrix{1, n + 1, [][]float64{initial}}

	res := make([]float64, n+1)
	for {
		next := calc_next(initial)
		initial_matrix.Element[0] = initial
		next_matrix := &mymath.Matrix{1, n + 1, [][]float64{next}}
		temp := math.Sqrt(initial_matrix.Times(next_matrix.T()).Element[0][0])
		if temp <= 1e-5 || count > recursion {
			res = next
			break
		}
		count++
		//println(calc_next(initial)[0])
		initial = next
	}

	return res
}
