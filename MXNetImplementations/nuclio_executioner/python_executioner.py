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
    os.environ['JOB_ID'] = body['job_id']
    os.environ['GOOGLE_APPLICATION_CREDENTIALS'] = "/opt/nuclio/secret_key.json"

    if 'num_parts' in body:
        os.environ['NUM_PARTS'] = body['num_parts']
    script_path = body['script_path']
    return start_real_nuclio(script_path)


def start_real_nuclio(script_path: str):
    # python /DLScheduler/MXNetImplementations/LeNet/MXNet/main.py input_args
    # return subprocess.check_output(["python", "/opt/nuclio/python_executioner.py"], env=os.environ)
    # "/DLScheduler/MXNetImplementations/LeNet/MXNet/main.py"
    try:
        return subprocess.check_output(["python", script_path], env=os.environ)
    except subprocess.CalledProcessError as e:
        return e.output


if __name__ == '__main__':
    script_path = "../Simple/main.py"
    print(subprocess.check_output(["python3", script_path], env=os.environ))
