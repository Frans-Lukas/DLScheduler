import json
import os
import subprocess


def start_from_nuclio(context, event):
    context.logger.info_with('Got invoked',
                             trigger_kind=event.trigger.kind,
                             event_body=event.body)
    body = json.loads(event.body)
    os.environ['DMLC_PS_ROOT_URI'] = body['ip']
    os.environ['DMLC_ROLE'] = body['role']
    os.environ['DMLC_NUM_SERVER'] = body['num_server']
    os.environ['DMLC_NUM_WORKER'] = body['num_worker']
    return start_real_nuclio()


def start_real_nuclio():
    # python /DLScheduler/MXNetImplementations/LeNet/MXNet/main.py input_args
    # return subprocess.check_output(["python", "/opt/nuclio/python_executioner.py"], env=os.environ)
    return subprocess.check_output(["python", "/DLScheduler/MXNetImplementations/LeNet/MXNet/main.py"], env=os.environ)

if __name__ == '__main__':
    print("teeest")