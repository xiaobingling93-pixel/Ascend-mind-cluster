#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025. Huawei Technologies Co.,Ltd. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================
from mindcluster_tools.interface import Queryable
from mindcluster_tools.utils import parse_eid

from bitarray import bitarray
from bitarray.util import ba2hex


class EIDGenerator(Queryable):
    """EID Generator class"""
    def query(self, *args, **kwargs):
        npu_id, die_id, port_id, fe_id, ch_id, eid_type = args
        eid = bitarray(parse_eid.EID_LENGTH)
        if eid_type == parse_eid.EID_TYPE_PHY:
            global_port = (npu_id % parse_eid.NPU_COUNT_IN_A_BOARD) * parse_eid.DIE_COUNT_IN_A_NPU * parse_eid.PHY_PORT_COUNT_IN_A_DIE \
                          + parse_eid.PHY_PORT_COUNT_IN_A_DIE * die_id \
                          + port_id \
                          + 1
        if eid_type == parse_eid.EID_TYPE_LOGIC:
            global_port = (npu_id % parse_eid.NPU_COUNT_IN_A_BOARD) * parse_eid.DIE_COUNT_IN_A_NPU * parse_eid.LOGIC_PORT_COUNT_IN_A_DIE \
                          + parse_eid.LOGIC_PORT_COUNT_IN_A_DIE * die_id \
                          + port_id \
                          + parse_eid.LOGIC_PORT_FLAG
        eid[parse_eid.PORT_ID_RANGE_START: parse_eid.PORT_ID_RANGE_END] = self._get_bitarray(global_port, parse_eid.PORT_ID_LENGTH)
        board = bitarray(bin(npu_id // parse_eid.NPU_COUNT_IN_A_BOARD)[2:].zfill(parse_eid.BOARD_ID_LENGTH))
        eid[parse_eid.BOARD_ID_RANGE_START: parse_eid.BOARD_ID_RANGE_END] = board
        eid[parse_eid.CHESSIS_ID_RANGE_START: parse_eid.CHESSIS_ID_RANGE_END] = self._get_bitarray(ch_id, parse_eid.CHESSIS_ID_LENGTH)
        # Fix to 233
        eid[parse_eid.UBC_223_RANGE_START: parse_eid.UBC_223_RANGE_END] = self._get_bitarray(223 + 223 * 256, parse_eid.UBC_223_LENGTH)
        eid[parse_eid.FE_ID_RANGE_START: parse_eid.FE_ID_RANGE_END] = self._get_bitarray(fe_id, parse_eid.FE_ID_LENGTH)
        # Bit 52 is fixed to 1
        eid[parse_eid.BIT_53TH_INDEX] = parse_eid.BIT_53TH_VALUE
        return ba2hex(eid)

    def _get_bitarray(self, decimal_num, length):
        return bitarray(bin(decimal_num)[2:].zfill(length))