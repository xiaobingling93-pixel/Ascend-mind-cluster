import unittest
from unittest.mock import MagicMock, patch
import os
from taskd.python.cython_api import cython_api
from taskd.python.framework.common.type import DEFAULT_SERVERRANK
from taskd.taskd.api import taskd_proxy_api


class TestTaskdProxyAPI(unittest.TestCase):
    def setUp(self):
        # Backup original cython_api.lib
        self.original_lib = cython_api.lib

    def tearDown(self):
        # Restore original cython_api.lib
        cython_api.lib = self.original_lib

    @patch('os.getenv')
    def test_init_taskd_proxy_success(self, mock_getenv):
        # Mock environment variables
        mock_getenv.side_effect = lambda key, default=None: DEFAULT_SERVERRANK if key in ["RANK", "MS_NODE_RANK"] else None

        # Mock cython_api.lib
        cython_api.lib = MagicMock()
        cython_api.lib.InitTaskdProxy.return_value = 0

        config = {}
        result = taskd_proxy_api.init_taskd_proxy(config)
        self.assertTrue(result)

    @patch('os.getenv')
    def test_init_taskd_proxy_lib_not_loaded(self, mock_getenv):
        # Mock environment variables
        mock_getenv.side_effect = lambda key, default=None: DEFAULT_SERVERRANK if key in ["RANK", "MS_NODE_RANK"] else None

        cython_api.lib = None
        result = taskd_proxy_api.init_taskd_proxy({})
        self.assertFalse(result)

    @patch('os.getenv')
    def test_init_taskd_proxy_init_failed(self, mock_getenv):
        # Mock environment variables
        mock_getenv.side_effect = lambda key, default=None: DEFAULT_SERVERRANK if key in ["RANK", "MS_NODE_RANK"] else None

        # Mock cython_api.lib
        cython_api.lib = MagicMock()
        cython_api.lib.InitTaskdProxy.return_value = 1

        config = {}
        result = taskd_proxy_api.init_taskd_proxy(config)
        self.assertFalse(result)

    @patch('os.getenv')
    def test_init_taskd_proxy_exception(self, mock_getenv):
        # Mock environment variables
        mock_getenv.side_effect = lambda key, default=None: DEFAULT_SERVERRANK if key in ["RANK", "MS_NODE_RANK"] else None

        # Mock cython_api.lib
        cython_api.lib = MagicMock()
        cython_api.lib.InitTaskdProxy.side_effect = Exception("Mock exception")

        config = {}
        result = taskd_proxy_api.init_taskd_proxy(config)
        self.assertFalse(result)

    def test_destroy_taskd_proxy_success(self):
        # Mock cython_api.lib
        cython_api.lib = MagicMock()
        cython_api.lib.DestroyTaskdProxy = MagicMock()

        result = taskd_proxy_api.destroy_taskd_proxy()
        self.assertTrue(result)

    def test_destroy_taskd_proxy_lib_not_loaded(self):
        cython_api.lib = None
        result = taskd_proxy_api.destroy_taskd_proxy()
        self.assertFalse(result)

    def test_destroy_taskd_proxy_exception(self):
        # Mock cython_api.lib
        cython_api.lib = MagicMock()
        cython_api.lib.DestroyTaskdProxy.side_effect = Exception("Mock exception")

        result = taskd_proxy_api.destroy_taskd_proxy()
        self.assertFalse(result)


if __name__ == '__main__':
    unittest.main()
    