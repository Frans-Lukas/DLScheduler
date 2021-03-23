package constants

const DEPLOY_FUNCTION_SCRIPT = "./nuclio/deploy_nuclio_docker_container.sh"
const INVOKE_FUNCTION_SCRIPT = "./nuclio/invoke_nuclio_function.sh"
const DELETE_FUNCTIONS_SUBSTRING_SCRIPT = "./nuclio/kill_all_nuclio_funcs_with_substr_and_range.sh"
const TRAIN_JOB_TYPE = "train"
const AGGREGATE_JOB_TYPE = "average"
const PY_PATH_ENV_NAME = "PY_PATH"
const PYTHON_LEAST_SQUARES = "pythonMath/leastSquaresFunc.py"
const JOB_TYPE_SCHEDULER = "scheduler"
const JOB_TYPE_WORKER = "worker"
const JOB_TYPE_SERVER = "server"
const KUBERNETES_NAMESPACE = "nuclio"
const MAX_SERVERS_PER_NODE = 2
const MAX_WORKERS_PER_NODE = 5
