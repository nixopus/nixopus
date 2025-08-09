import signal
import time
import unittest
from unittest.mock import Mock, patch

from app.utils.timeout import TimeoutWrapper
from app.commands.install.messages import timeout_error


class TestTimeoutWrapper(unittest.TestCase):
    def setUp(self):
        self.original_signal = signal.signal
        self.original_alarm = signal.alarm

    def tearDown(self):
        signal.signal = self.original_signal
        signal.alarm = self.original_alarm

    def test_timeout_wrapper_zero_timeout(self):
        with TimeoutWrapper(0) as wrapper:
            self.assertEqual(wrapper.timeout, 0)
            time.sleep(0.1)

    def test_timeout_wrapper_negative_timeout(self):
        with TimeoutWrapper(-1) as wrapper:
            self.assertEqual(wrapper.timeout, -1)
            time.sleep(0.1)

    def test_timeout_wrapper_positive_timeout_success(self):
        with TimeoutWrapper(5) as wrapper:
            self.assertEqual(wrapper.timeout, 5)
            time.sleep(0.1)

    @patch('signal.signal')
    @patch('signal.alarm')
    def test_timeout_wrapper_signal_setup(self, mock_alarm, mock_signal):
        mock_signal.return_value = None
        
        with TimeoutWrapper(10) as wrapper:
            mock_signal.assert_called_once_with(signal.SIGALRM, unittest.mock.ANY)
            mock_alarm.assert_called_once_with(10)

    @patch('signal.signal')
    @patch('signal.alarm')
    def test_timeout_wrapper_signal_cleanup(self, mock_alarm, mock_signal):
        mock_signal.return_value = None
        
        with TimeoutWrapper(10):
            pass
        
        mock_alarm.assert_has_calls([
            unittest.mock.call(10),
            unittest.mock.call(0)
        ])
        self.assertEqual(mock_signal.call_count, 2)

    @patch('signal.signal')
    @patch('signal.alarm')
    def test_timeout_wrapper_zero_timeout_no_signal_setup(self, mock_alarm, mock_signal):
        with TimeoutWrapper(0):
            pass
        
        mock_signal.assert_not_called()
        mock_alarm.assert_not_called()

    @patch('signal.signal')
    @patch('signal.alarm')
    def test_timeout_wrapper_negative_timeout_no_signal_setup(self, mock_alarm, mock_signal):
        with TimeoutWrapper(-5):
            pass
        
        mock_signal.assert_not_called()
        mock_alarm.assert_not_called()

    def test_timeout_wrapper_timeout_triggered(self):
        with self.assertRaises(TimeoutError) as context:
            with TimeoutWrapper(1):
                time.sleep(2)
        
        expected_message = timeout_error.format(timeout=1)
        self.assertEqual(str(context.exception), expected_message)

    @patch('signal.signal')
    @patch('signal.alarm')
    def test_timeout_wrapper_exception_handling(self, mock_alarm, mock_signal):
        mock_signal.return_value = None
        
        with self.assertRaises(ValueError):
            with TimeoutWrapper(10):
                raise ValueError("Test exception")
        
        mock_alarm.assert_has_calls([
            unittest.mock.call(10),
            unittest.mock.call(0)
        ])
        mock_signal.assert_has_calls([
            unittest.mock.call(signal.SIGALRM, unittest.mock.ANY),
            unittest.mock.call(signal.SIGALRM, None)
        ])

    @patch('signal.signal')
    @patch('signal.alarm')
    def test_timeout_wrapper_signal_handler_registration(self, mock_alarm, mock_signal):
        mock_signal.return_value = None
        
        with TimeoutWrapper(5) as wrapper:
            mock_signal.assert_called_once()
            call_args = mock_signal.call_args
            self.assertEqual(call_args[0][0], signal.SIGALRM)
            self.assertTrue(callable(call_args[0][1]))

    @patch('signal.signal')
    @patch('signal.alarm')
    def test_timeout_wrapper_signal_handler_raises_timeout(self, mock_alarm, mock_signal):
        mock_signal.return_value = None
        
        with TimeoutWrapper(5) as wrapper:
            signal_handler = mock_signal.call_args[0][1]
            
            with self.assertRaises(TimeoutError) as context:
                signal_handler(signal.SIGALRM, None)
            
            expected_message = timeout_error.format(timeout=5)
            self.assertEqual(str(context.exception), expected_message)

    def test_timeout_wrapper_nested_usage(self):
        with TimeoutWrapper(10) as outer:
            with TimeoutWrapper(5) as inner:
                self.assertEqual(outer.timeout, 10)
                self.assertEqual(inner.timeout, 5)
                time.sleep(0.1)

    @patch('signal.signal')
    @patch('signal.alarm')
    def test_timeout_wrapper_multiple_instances(self, mock_alarm, mock_signal):
        mock_signal.return_value = None
        
        wrapper1 = TimeoutWrapper(10)
        wrapper2 = TimeoutWrapper(5)
        
        with wrapper1:
            with wrapper2:
                pass
        
        self.assertEqual(mock_alarm.call_count, 4)

    def test_timeout_wrapper_return_value(self):
        with TimeoutWrapper(10) as wrapper:
            self.assertIsInstance(wrapper, TimeoutWrapper)
            self.assertEqual(wrapper.timeout, 10)

    @patch('signal.signal')
    @patch('signal.alarm')
    def test_timeout_wrapper_signal_restoration(self, mock_alarm, mock_signal):
        original_handler = Mock()
        mock_signal.return_value = original_handler
        
        with TimeoutWrapper(10):
            pass
        
        mock_signal.assert_has_calls([
            unittest.mock.call(signal.SIGALRM, unittest.mock.ANY),
            unittest.mock.call(signal.SIGALRM, original_handler)
        ])

    def test_timeout_wrapper_large_timeout_value(self):
        with TimeoutWrapper(999999) as wrapper:
            self.assertEqual(wrapper.timeout, 999999)
            time.sleep(0.1)

    @patch('signal.signal')
    @patch('signal.alarm')
    def test_timeout_wrapper_signal_error_handling(self, mock_alarm, mock_signal):
        mock_signal.side_effect = OSError("Signal not supported")
        
        with self.assertRaises(OSError):
            with TimeoutWrapper(10):
                pass

    @patch('signal.signal')
    @patch('signal.alarm')
    def test_timeout_wrapper_alarm_error_handling(self, mock_alarm, mock_signal):
        mock_signal.return_value = None
        mock_alarm.side_effect = OSError("Alarm not supported")
        
        with self.assertRaises(OSError):
            with TimeoutWrapper(10):
                pass


if __name__ == "__main__":
    unittest.main() 