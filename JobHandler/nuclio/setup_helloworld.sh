#!/bin/bash

nuctl deploy helloworld -n nuclio -p https://raw.githubusercontent.com/nuclio/nuclio/master/hack/examples/golang/helloworld/helloworld.go --registry docker.io/franslukas --http-trigger-service-type=NodePort


#Exception caught in handler - "Command '['python', '/DLScheduler/MXNetImplementations/LeNet/MXNet/main.py']' returned non-zero exit status 1.": Traceback (most recent call last):
#  File "/opt/nuclio/_nuclio_wrapper.py", line 116, in serve_requests
#    self._handle_event(event)
#  File "/opt/nuclio/_nuclio_wrapper.py", line 274, in _handle_event
#    entrypoint_output = self._entrypoint(self._context, event)
#  File "/opt/nuclio/python_executioner.py", line 15, in start_from_nuclio
#    return start_real_nuclio()
#  File "/opt/nuclio/python_executioner.py", line 21, in start_real_nuclio
#    return subprocess.check_output(["python", "/DLScheduler/MXNetImplementations/LeNet/MXNet/main.py"], env=os.environ)
#  File "/usr/local/lib/python3.6/subprocess.py", line 356, in check_output
#    **kwargs).stdout
#  File "/usr/local/lib/python3.6/subprocess.py", line 438, in run
#    output=stdout, stderr=stderr)
#subprocess.CalledProcessError: Command '['python', '/DLScheduler/MXNetImplementations/LeNet/MXNet/main.py']' returned non-zero exit status 1.
#
#
#2021/03/16 16:00:19 InvokeFunction, regexp:  unexpected end of JSON input
