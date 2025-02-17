import os
import unittest
from unittest.mock import patch
from component.taskd.taskd.python.framework.agent.ms_mgr.MsUtils import calculate_global_rank


class TestCalculateGlobalRank(unittest.TestCase):
    @patch('os.getenv')
    def test_valid_env_vars(self, mock_getenv):
        # 模拟环境变量
        mock_getenv.side_effect = ['2', '3']
        result = calculate_global_rank()
        expected = [3 * 2 + 0, 3 * 2 + 1]
        self.assertEqual(result, expected)

    @patch('os.getenv')
    def test_missing_env_vars(self, mock_getenv):
        # 模拟缺少环境变量
        mock_getenv.side_effect = [None, '3']
        result = calculate_global_rank()
        self.assertEqual(result, [])

    @patch('os.getenv')
    def test_invalid_env_vars(self, mock_getenv):
        # 模拟无效的环境变量
        mock_getenv.side_effect = ['abc', '3']
        result = calculate_global_rank()
        self.assertEqual(result, [])

if __name__ == '__main__':
    unittest.main()