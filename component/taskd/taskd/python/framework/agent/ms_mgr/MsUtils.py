import os

from taskd.python.toolkit.constants import constants
from taskd.python.toolkit.logger.log import run_log

# check_monitor_res_valid to check whether mindspore monitor interface given a valid result
def check_monitor_res_valid(rank_status_dict: dict):
    if not isinstance(rank_status_dict, dict):
        run_log.warning("monitor result should be a dict")
        return False

    for rank, info in rank_status_dict.items():
        if not isinstance(info, dict):
            run_log.warning(f"monitor result for rank {rank} should be a dict")
            return False

        # 校验每个字典中必须有 'pid', 'status', 'global_rank' 键
        required_keys = ['pid', 'status', constants.GLOBAL_RANK_ID_KEY]
        for key in required_keys:
            if key not in info:
                run_log.warning(f"{rank} has no key: {key}")
                return False

        # 校验 'pid' 是整数
        if not isinstance(info['pid'], int):
            run_log.warning(f"info['pid']is not int, but {info['pid']}")
            return False

        # 校验 'status' 是整数
        if not isinstance(info['status'], int) and info['status'] is not None:
            run_log.warning(f"info['status']is not int,but {info['status']}")
            return False

        # 校验 'global_rank_id' 是整数
        if not isinstance(info[constants.GLOBAL_RANK_ID_KEY], int):
            run_log.warning(f"info['global_rank']is not int, but {info[constants.GLOBAL_RANK_ID_KEY]}")
            return False
    return True

# 从环境变量计算当前节点的global rank 数组
def calculate_global_rank():
    # 从环境变量中获取 MS_LOCAL_WORKER 和 MS_NODE_RANK
    ms_local_worker = os.getenv('MS_LOCAL_WORKER')
    ms_node_rank = os.getenv('MS_NODE_RANK')
    if ms_local_worker is None or ms_node_rank is None:
        run_log.error("环境变量 MS_LOCAL_WORKER 或 MS_NODE_RANK 未设置")
        return []
    try:
        ms_local_worker = int(ms_local_worker)
        ms_node_rank = int(ms_node_rank)
    except ValueError as e:
        run_log.info(f"failed to get MS_LOCAL_WORKER and MS_NODE_RANK from env, please set it: {e}")
        return []
    global_rank = []
    for local_worker in range(ms_local_worker):
        # 计算 global_rank
        global_rank.append(ms_node_rank * ms_local_worker + local_worker)
    return global_rank