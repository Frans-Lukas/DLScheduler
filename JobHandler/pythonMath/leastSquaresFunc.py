#!/bin/python

import sys

import numpy as np
from scipy.optimize import least_squares


def h(theta, x, y):
    return theta[2] * (x - theta[0]) ** 2 + theta[3] * (y - theta[1]) ** 2


def fun(theta):
    return (h(theta, xs, ys) - hs).flatten()


# example input
# -1. -0.89473684 -0.78947368 -0.68421053 -0.57894737 -0.47368421 -0.36842105 -0.26315789 -0.15789474 -0.05263158 0.05263158 0.15789474 0.26315789 0.36842105 0.47368421 0.57894737 0.68421053 0.78947368 0.89473684 1.
# -1. -0.89473684 -0.78947368 -0.68421053 -0.57894737 -0.47368421 -0.36842105 -0.26315789 -0.15789474 -0.05263158 0.05263158 0.15789474 0.26315789 0.36842105 0.47368421 0.57894737 0.68421053 0.78947368 0.89473684 1.
# 2.655 2.09876731 1.60901662 1.18574792 0.82896122 0.53865651 0.3148338 0.15749307 0.06663435 0.04225762 0.08436288 0.19295014 0.36801939 0.60957064 0.91760388 1.29211911 1.73311634 2.24059557 2.81455679 3.455
if __name__ == '__main__':
    if len(sys.argv) < 4:
        print("Too few arguments")
        exit(-1)
        

    xInput = sys.argv[1]
    xs = np.fromstring(xInput, dtype=float, sep=' ')

    yInput = sys.argv[2]
    ys = np.fromstring(yInput, dtype=float, sep=' ')

    hInput = sys.argv[3]
    hs = np.fromstring(hInput, dtype=float, sep=' ')

    theta0 = [0, 0, 1, 2]
    res3 = least_squares(fun, theta0)
    thetaRes = res3.get("x")

    print(thetaRes)

# See PyCharm help at https://www.jetbrains.com/help/pycharm/