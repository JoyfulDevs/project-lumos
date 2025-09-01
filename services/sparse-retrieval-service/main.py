#!/usr/bin/env python3
"""
Sparse Retrieval Service Main Entry Point
"""
import logging
import sys
import os

# Add parent directory to path for imports
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from app.app import run


def main():
    """Main entry point"""
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        handlers=[logging.StreamHandler()]
    )
    
    logger = logging.getLogger(__name__)
    logger.info("Sparse Retrieval Service starting")
    
    try:
        run()
    except Exception as e:
        logger.error(f"Failed to run sparse retrieval service: {e}", exc_info=True)
        sys.exit(1)
    
    logger.info("Sparse Retrieval Service finished")


if __name__ == "__main__":
    main()