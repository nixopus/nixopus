"""
Unit tests for the InputParser class.

This module contains comprehensive test cases for the InputParser class, covering:
- Command line argument parsing and validation
- Password generation and validation
- Domain validation and handling
- Admin credential management
- Environment configuration
"""

import argparse
import sys
import unittest
from unittest.mock import patch, MagicMock

from installer.input_parser import InputParser
from installer.validation import Validation


class TestInputParser(unittest.TestCase):
    """Test suite for InputParser class functionality.
    
    Tests cover all major functionality including argument parsing, validation,
    password generation, and credential management.
    """
    
    def setUp(self) -> None:
        """Set up test fixtures before each test method."""
        self.parser = InputParser()

    def test_setup_arg_parser(self) -> None:
        """Test argument parser setup and basic argument parsing functionality."""
        parser = self.parser._setup_arg_parser()
        self.assertIsInstance(parser, argparse.ArgumentParser)

        args = parser.parse_args([
            '--api-domain', 'api.example.com',
            '--app-domain', 'app.example.com',
            '--email', 'test@example.com',
            '--password', 'Test123!',
            '--env', 'production'
        ])

        self.assertEqual(args.api_domain, 'api.example.com')
        self.assertEqual(args.app_domain, 'app.example.com')
        self.assertEqual(args.email, 'test@example.com')
        self.assertEqual(args.password, 'Test123!')
        self.assertEqual(args.env, 'production')

    def test_setup_arg_parser_invalid_env(self) -> None:
        """Test argument parser with invalid environment value."""
        parser = self.parser._setup_arg_parser()
        with self.assertRaises(SystemExit):
            parser.parse_args(['--env', 'invalid_env'])

    def test_setup_arg_parser_short_options(self) -> None:
        """Test argument parser with short form options."""
        parser = self.parser._setup_arg_parser()
        args = parser.parse_args(['-e', 'test@example.com', '-p', 'Test123!'])
        self.assertEqual(args.email, 'test@example.com')
        self.assertEqual(args.password, 'Test123!')

    def test_generate_strong_password(self) -> None:
        """Test strong password generation with validation."""
        password = self.parser.generate_strong_password()
        self.assertEqual(len(password), 16)
        self.assertTrue(any(c.isupper() for c in password))
        self.assertTrue(any(c.islower() for c in password))
        self.assertTrue(any(c.isdigit() for c in password))
        validation = Validation()
        self.assertTrue(any(c in validation.SPECIAL_CHARS for c in password))

    def test_generate_strong_password_multiple_generations(self) -> None:
        """Test multiple password generations for uniqueness."""
        passwords = set()
        for _ in range(10):
            password = self.parser.generate_strong_password()
            self.assertNotIn(password, passwords)
            passwords.add(password)

    def test_get_env_from_args(self) -> None:
        """Test environment extraction from arguments."""
        args = MagicMock()
        args.env = 'staging'
        self.assertEqual(self.parser.get_env_from_args(args), 'staging')
        
        args.env = None
        self.assertEqual(self.parser.get_env_from_args(args), 'production')

    def test_get_env_from_args_invalid_env(self) -> None:
        """Test environment validation with invalid value."""
        args = MagicMock()
        args.env = 'invalid_env'
        with self.assertRaises(SystemExit):
            self.parser.get_env_from_args(args)

    @patch('installer.input_parser.Validation')
    def test_get_domains_from_args(self, mock_validation: MagicMock) -> None:
        """Test domain extraction and validation from arguments."""
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        
        args = MagicMock()
        args.api_domain = 'api.example.com'
        args.app_domain = 'app.example.com'
        
        domains = self.parser.get_domains_from_args(args)
        self.assertEqual(domains['api_domain'], 'api.example.com')
        self.assertEqual(domains['app_domain'], 'app.example.com')
        
        args.api_domain = None
        args.app_domain = 'app.example.com'
        domains = self.parser.get_domains_from_args(args)
        self.assertIsNone(domains)

    @patch('installer.input_parser.Validation')
    def test_get_domains_from_args_invalid_domain(self, mock_validation: MagicMock) -> None:
        """Test domain validation with invalid domain."""
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        mock_validation_instance.validate_domain.side_effect = SystemExit()
        
        args = MagicMock()
        args.api_domain = 'invalid.domain'
        args.app_domain = 'app.example.com'
        
        domains = self.parser.get_domains_from_args(args)
        self.assertIsNone(domains)

    @patch('installer.input_parser.Validation')
    def test_get_domains_from_args_no_domains(self, mock_validation: MagicMock) -> None:
        """Test domain handling when no domains are provided."""
        args = MagicMock()
        args.api_domain = None
        args.app_domain = None
        
        domains = self.parser.get_domains_from_args(args)
        self.assertIsNone(domains)

    @patch('installer.input_parser.Validation')
    def test_get_admin_credentials_from_args(self, mock_validation: MagicMock) -> None:
        """Test admin credential extraction and validation."""
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        
        args = MagicMock()
        args.email = 'test@example.com'
        args.password = 'Test123!'
        
        email, password = self.parser.get_admin_credentials_from_args(args)
        self.assertEqual(email, 'test@example.com')
        self.assertEqual(password, 'Test123!')

    @patch('installer.input_parser.Validation')
    def test_get_admin_credentials_from_args_invalid_email(self, mock_validation: MagicMock) -> None:
        """Test admin credential validation with invalid email."""
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        mock_validation_instance.validate_email.side_effect = SystemExit()
        
        args = MagicMock()
        args.email = 'invalid.email'
        args.password = 'Test123!'
        
        with self.assertRaises(SystemExit):
            self.parser.get_admin_credentials_from_args(args)

    @patch('installer.input_parser.Validation')
    def test_get_admin_credentials_from_args_invalid_password(self, mock_validation: MagicMock) -> None:
        """Test admin credential validation with invalid password."""
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        mock_validation_instance.validate_password.side_effect = SystemExit()
        
        args = MagicMock()
        args.email = 'test@example.com'
        args.password = 'weak'
        
        with self.assertRaises(SystemExit):
            self.parser.get_admin_credentials_from_args(args)

    @patch('builtins.input')
    @patch('installer.input_parser.Validation')
    def test_ask_admin_credentials(self, mock_validation: MagicMock, mock_input: MagicMock) -> None:
        """Test interactive admin credential collection."""
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        mock_input.side_effect = ['test@example.com', 'Test123!']
        
        email, password = self.parser.ask_admin_credentials()
        self.assertEqual(email, 'test@example.com')
        self.assertEqual(password, 'Test123!')

    @patch('builtins.input')
    @patch('installer.input_parser.Validation')
    def test_ask_admin_credentials_empty_password(self, mock_validation: MagicMock, mock_input: MagicMock) -> None:
        """Test admin credential collection with empty password."""
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        mock_input.side_effect = ['test@example.com', '']
        
        with patch.object(self.parser, 'generate_strong_password', return_value='Generated123!'):
            email, password = self.parser.ask_admin_credentials()
            self.assertEqual(email, 'test@example.com')
            self.assertEqual(password, 'Generated123!')

    @patch('builtins.input')
    @patch('installer.input_parser.Validation')
    def test_ask_admin_credentials_invalid_email_retry(self, mock_validation: MagicMock, mock_input: MagicMock) -> None:
        """Test admin credential collection with invalid email retry."""
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        mock_validation_instance.validate_email.side_effect = [SystemExit(), None]
        mock_input.side_effect = ['invalid.email', 'test@example.com', 'Test123!']
        
        email, password = self.parser.ask_admin_credentials()
        self.assertEqual(email, 'test@example.com')
        self.assertEqual(password, 'Test123!')


if __name__ == '__main__':
    unittest.main() 