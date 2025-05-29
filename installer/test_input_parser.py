import unittest
from unittest.mock import patch, MagicMock
from installer.input_parser import InputParser
import argparse

class TestInputParser(unittest.TestCase):
    def setUp(self):
        self.parser = InputParser()

    def test_setup_arg_parser(self):
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

    def test_setup_arg_parser_invalid_env(self):
        parser = self.parser._setup_arg_parser()
        with self.assertRaises(SystemExit):
            parser.parse_args(['--env', 'invalid_env'])

    def test_setup_arg_parser_short_options(self):
        parser = self.parser._setup_arg_parser()
        args = parser.parse_args(['-e', 'test@example.com', '-p', 'Test123!'])
        self.assertEqual(args.email, 'test@example.com')
        self.assertEqual(args.password, 'Test123!')

    def test_generate_strong_password(self):
        password = self.parser.generate_strong_password()
        self.assertEqual(len(password), 16)
        self.assertTrue(any(c.isupper() for c in password))
        self.assertTrue(any(c.islower() for c in password))
        self.assertTrue(any(c.isdigit() for c in password))
        self.assertTrue(any(c in '!@#$%^&*()_+-=[]{}|;:,.<>?' for c in password))

    def test_generate_strong_password_multiple_generations(self):
        passwords = set()
        for _ in range(10):
            password = self.parser.generate_strong_password()
            self.assertNotIn(password, passwords)
            passwords.add(password)

    def test_get_env_from_args(self):
        args = MagicMock()
        args.env = 'staging'
        self.assertEqual(self.parser.get_env_from_args(args), 'staging')
        
        args.env = None
        self.assertEqual(self.parser.get_env_from_args(args), 'production')

    def test_get_env_from_args_invalid_env(self):
        args = MagicMock()
        args.env = 'invalid_env'
        with self.assertRaises(SystemExit):
            self.parser.get_env_from_args(args)

    @patch('installer.input_parser.Validation')
    def test_get_domains_from_args(self, mock_validation):
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
        self.assertEqual(domains['api_domain'], 'nixopusapi.example.com')
        self.assertEqual(domains['app_domain'], 'app.example.com')

    @patch('installer.input_parser.Validation')
    def test_get_domains_from_args_invalid_domain(self, mock_validation):
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        mock_validation_instance.validate_domain.side_effect = SystemExit()
        
        args = MagicMock()
        args.api_domain = 'invalid.domain'
        args.app_domain = 'app.example.com'
        
        domains = self.parser.get_domains_from_args(args)
        self.assertIsNone(domains)

    @patch('installer.input_parser.Validation')
    def test_get_domains_from_args_no_domains(self, mock_validation):
        args = MagicMock()
        args.api_domain = None
        args.app_domain = None
        
        domains = self.parser.get_domains_from_args(args)
        self.assertIsNone(domains)

    @patch('installer.input_parser.Validation')
    def test_get_admin_credentials_from_args(self, mock_validation):
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        
        args = MagicMock()
        args.email = 'test@example.com'
        args.password = 'Test123!'
        
        email, password = self.parser.get_admin_credentials_from_args(args)
        self.assertEqual(email, 'test@example.com')
        self.assertEqual(password, 'Test123!')

    @patch('installer.input_parser.Validation')
    def test_get_admin_credentials_from_args_invalid_email(self, mock_validation):
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        mock_validation_instance.validate_email.side_effect = SystemExit()
        
        args = MagicMock()
        args.email = 'invalid.email'
        args.password = 'Test123!'
        
        with self.assertRaises(SystemExit):
            self.parser.get_admin_credentials_from_args(args)

    @patch('installer.input_parser.Validation')
    def test_get_admin_credentials_from_args_invalid_password(self, mock_validation):
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
    def test_ask_admin_credentials(self, mock_validation, mock_input):
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        mock_input.side_effect = ['test@example.com', 'Test123!']
        
        email, password = self.parser.ask_admin_credentials()
        self.assertEqual(email, 'test@example.com')
        self.assertEqual(password, 'Test123!')

    @patch('builtins.input')
    @patch('installer.input_parser.Validation')
    def test_ask_admin_credentials_empty_password(self, mock_validation, mock_input):
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        mock_input.side_effect = ['test@example.com', '']
        
        with patch.object(self.parser, 'generate_strong_password', return_value='Generated123!'):
            email, password = self.parser.ask_admin_credentials()
            self.assertEqual(email, 'test@example.com')
            self.assertEqual(password, 'Generated123!')

    @patch('builtins.input')
    @patch('installer.input_parser.Validation')
    def test_ask_admin_credentials_invalid_email_retry(self, mock_validation, mock_input):
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        mock_validation_instance.validate_email.side_effect = [SystemExit(), None]
        mock_input.side_effect = ['invalid.email', 'test@example.com', 'Test123!']
        
        email, password = self.parser.ask_admin_credentials()
        self.assertEqual(email, 'test@example.com')
        self.assertEqual(password, 'Test123!')

    @patch('builtins.input')
    @patch('installer.input_parser.Validation')
    def test_ask_for_domain(self, mock_validation, mock_input):
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        mock_input.return_value = 'example.com'
        
        domains = self.parser.ask_for_domain()
        self.assertEqual(domains['api_domain'], 'nixopusapi.example.com')
        self.assertEqual(domains['app_domain'], 'nixopus.example.com')

    @patch('builtins.input')
    @patch('installer.input_parser.Validation')
    def test_ask_for_domain_invalid_retry(self, mock_validation, mock_input):
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        mock_validation_instance.validate_domain.side_effect = [SystemExit(), None]
        mock_input.side_effect = ['invalid.domain', 'example.com']
        
        domains = self.parser.ask_for_domain()
        self.assertEqual(domains['api_domain'], 'nixopusapi.example.com')
        self.assertEqual(domains['app_domain'], 'nixopus.example.com')

    @patch('builtins.input')
    @patch('installer.input_parser.Validation')
    def test_ask_for_domain_empty_input(self, mock_validation, mock_input):
        mock_validation_instance = MagicMock()
        mock_validation.return_value = mock_validation_instance
        mock_input.side_effect = ['', 'example.com']
        
        domains = self.parser.ask_for_domain()
        self.assertEqual(domains['api_domain'], 'nixopusapi.example.com')
        self.assertEqual(domains['app_domain'], 'nixopus.example.com')

if __name__ == '__main__':
    unittest.main() 