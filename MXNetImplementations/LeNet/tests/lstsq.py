import numpy as np
import spectrochempy as scp
from spectrochempy import ur

def func(t, v, var):
    d = v * t + (np.random.rand(len(t)) - 0.5) * var
    d[0].data = 0.
    return d


time = scp.LinearCoord.linspace(0, 10, 20, title='time', units='hour')
d = scp.NDDataset.fromfunction(func, v=100. * ur('km/hr'), var=60. * ur('km'),
                               # extra arguments passed to the function v, var
                               coordset=scp.CoordSet(t=time), name='mydataset', title='distance travelled')