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
class Topo(object):
    """Read topology file to generate topology object"""
    def __init__(self, topo_file_path):
        self._topo = {}

    def get_ports_by_level_and_die(self, local_id, level, die_id):
        pass


class TopoSingleFactory:
    """Singleton factory for topology objects"""
    _topo_path = None
    _topo = None

    @staticmethod
    def set_topo_path(topo_path):
        TopoSingleFactory._topo_path = topo_path

    @staticmethod
    def get_topo():
        if TopoSingleFactory._topo is None:
            TopoSingleFactory._topo = Topo(TopoSingleFactory._topo_path)
        return TopoSingleFactory._topo