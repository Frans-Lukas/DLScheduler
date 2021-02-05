# ResNet50 implementation training on CIFAR-10 dataset

Used with Python 3.8.
To install requirements simply call
`pip install -r requirements.txt`
and run with 
`python main.py`

The file will automatically download the dataset if it does not exist.

Thanks to annytab.com for providing the implementation at [https://www.annytab.com/resnet50-image-classification-in-python/](https://www.annytab.com/resnet50-image-classification-in-python/) and thanks to @YoongiKim for providing the test and training dataset at https://github.com/YoongiKim/CIFAR-10-images

 
Training with the validation model for 1562 epochs gives the following results:

`1562/1562 [==============================] - 1769s 1s/step - loss: 3.4929 - accuracy: 0.1872 - val_loss: 2.1343 - val_accuracy: 0.2936`
`1562/1562 [==============================] - 1783s 1s/step - loss: 1.5285 - accuracy: 0.4444 - val_loss: 1.5359 - val_accuracy: 0.4493`
`1562/1562 [==============================] - 1715s 1s/step - loss: 1.3679 - accuracy: 0.5095 - val_loss: 3.6235 - val_accuracy: 0.4485`