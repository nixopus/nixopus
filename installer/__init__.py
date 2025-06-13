"""
Installer package for Nixopus.
"""

import logging

def setup_logger(debug: bool = False):
    logger = logging.getLogger("nixopus")
    logger.setLevel(logging.DEBUG if debug else logging.INFO)
    
    # Create console handler with formatting
    handler = logging.StreamHandler()
    formatter = logging.Formatter('%(levelname)s: %(message)s')
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    
    return logger

# Create a default logger instance
logger = setup_logger() 