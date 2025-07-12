#!/usr/bin/python3
# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

from unittest.mock import patch, MagicMock
from unittest import TestCase
from taskd.python.toolkit.validator.cert_check import CertContentsChecker
from datetime import datetime, timedelta

class TestCertContentsChecker(TestCase):

    @patch('taskd.python.toolkit.validator.cert_check.ParseCertInfo')
    def test_check_cert_info(self, mock_parse_cert_info):
        cert_checker = CertContentsChecker()
        cert_bytes = b'test_cert_bytes'

        # Mocking the ParseCertInfo class
        mock_cert_info = MagicMock()
        mock_parse_cert_info.return_value = mock_cert_info

        # Testing valid certificate
        mock_cert_info.start_time = datetime.utcnow() - timedelta(days=1)
        mock_cert_info.end_time = datetime.utcnow() + timedelta(days=1)
        mock_cert_info.cert_version = cert_checker.X509_V3
        mock_cert_info.pubkey_type = 6
        mock_cert_info.signature_len = cert_checker.RSA_LEN_LIMIT + 1
        mock_cert_info.signature_algorithm = 'sha256WithRSAEncryption'
        mock_cert_info.extensions = {'basicConstraints': 'CA:TRUE', 'keyUsage': 'Digital Signature'}
        title = "CN"
        title_byte = title.encode('utf-8')
        name = "test.com"
        name_byte = name.encode('utf-8')
        mock_cert_info.subject_components = [(title_byte, name_byte)]

        domain_name = cert_checker.check_cert_info(cert_bytes)
        self.assertEqual(domain_name, 'test.com')

        # Testing invalid certificate validity period
        mock_cert_info.start_time = datetime.utcnow() + timedelta(days=1)
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)

        # Testing invalid certificate version
        mock_cert_info.start_time = datetime.utcnow() - timedelta(days=1)
        mock_cert_info.cert_version = 'Invalid Version'
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)

        # Testing invalid certificate pubkey type
        mock_cert_info.cert_version = cert_checker.X509_V3
        mock_cert_info.pubkey_type = 'Invalid Type'
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)

        # Testing invalid RSA pubkey length
        mock_cert_info.pubkey_type = 'EVP_PKEY_RSA'
        mock_cert_info.signature_len = cert_checker.RSA_LEN_LIMIT - 1
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)

        # Testing invalid EC pubkey length
        mock_cert_info.pubkey_type = 'EVP_PKEY_EC'
        mock_cert_info.signature_len = cert_checker.EC_LEN_LIMIT - 1
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)

        # Testing invalid signature algorithm
        mock_cert_info.signature_len = cert_checker.RSA_LEN_LIMIT + 1
        mock_cert_info.signature_algorithm = 'Invalid Algorithm'
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)

        # Testing missing 'CA' in basic constraints
        mock_cert_info.signature_algorithm = 'SHA256'
        mock_cert_info.extensions = {'basicConstraints': '', 'keyUsage': 'Digital Signature'}
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)

        # Testing missing 'Digital Signature' in key usage
        mock_cert_info.extensions = {'basicConstraints': 'CA:TRUE', 'keyUsage': ''}
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)