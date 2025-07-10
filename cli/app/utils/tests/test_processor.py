import unittest
import time
from app.utils.lib import ParallelProcessor

class TestParallelProcessor(unittest.TestCase):
    
    def test_basic_processing(self):
        """Test basic parallel processing functionality"""
        def square(x):
            return x * x
        
        items = [1, 2, 3, 4, 5]
        results = ParallelProcessor.process_items(items, square)
        
        # Results are in completion order, not input order
        self.assertEqual(len(results), 5)
        self.assertEqual(set(results), {1, 4, 9, 16, 25})
    
    def test_error_handling(self):
        """Test error handling in parallel processing"""
        def process_with_error(x):
            if x == 3:
                raise ValueError("Test error")
            return x * 2
        
        def error_handler(item, error):
            return f"Error processing {item}: {str(error)}"
        
        items = [1, 2, 3, 4, 5]
        results = ParallelProcessor.process_items(
            items, 
            process_with_error, 
            error_handler=error_handler
        )
        
        self.assertEqual(len(results), 5)
        # Check that we have the expected results (order may vary)
        expected_results = {2, 4, 8, 10}  # 1*2, 2*2, 4*2, 5*2
        error_results = [r for r in results if "Error processing 3" in str(r)]
        normal_results = [r for r in results if isinstance(r, int)]
        
        self.assertEqual(len(error_results), 1)
        self.assertEqual(set(normal_results), expected_results)
    
    def test_timeout_behavior(self):
        """Test that processing respects timeout behavior"""
        def slow_process(x):
            time.sleep(0.1)
            return x * 2
        
        items = list(range(10))
        start_time = time.time()
        results = ParallelProcessor.process_items(items, slow_process, max_workers=5)
        end_time = time.time()
        
        self.assertEqual(len(results), 10)
        # Results are in completion order, not input order
        self.assertEqual(set(results), {0, 2, 4, 6, 8, 10, 12, 14, 16, 18})
        
        # With 5 workers and 10 items taking 0.1s each, should complete in ~0.2s
        # (2 batches of 5 items each)
        self.assertLess(end_time - start_time, 0.5)
    
    def test_empty_list(self):
        """Test processing empty list"""
        def process(x):
            return x * 2
        
        results = ParallelProcessor.process_items([], process)
        self.assertEqual(results, [])
    
    def test_single_item(self):
        """Test processing single item"""
        def process(x):
            return x * 2
        
        results = ParallelProcessor.process_items([5], process)
        self.assertEqual(results, [10])

if __name__ == '__main__':
    unittest.main() 