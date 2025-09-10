"""
BM42 Sparse Embedding Adapter
"""
import logging
from typing import List, Tuple
import numpy as np
from fastembed import SparseTextEmbedding

logger = logging.getLogger(__name__)


class BM42Embedder:
    """BM42 sparse text embedding adapter"""
    
    def __init__(self, model_name: str = "Qdrant/bm42-all-minilm-l6-v2-attentions"):
        """Initialize BM42 embedder
        
        Args:
            model_name: Name of the BM42 model to use
        """
        self.model_name = model_name
        logger.info(f"Initializing BM42 embedder with model: {model_name}")
        self.model = SparseTextEmbedding(model_name=model_name)
        logger.info("BM42 embedder initialized successfully")
    
    def embed(self, text: str) -> Tuple[List[int], List[float]]:
        """Generate sparse embedding for text
        
        Args:
            text: Input text to embed
            
        Returns:
            Tuple of (indices, values) for sparse vector
        """
        try:
            # Generate embedding
            embedding = list(self.model.embed([text]))[0]
            
            # Convert to lists
            indices = embedding.indices.tolist()
            values = embedding.values.tolist()
            
            return indices, values
            
        except Exception as e:
            logger.error(f"Failed to generate embedding: {e}")
            raise
    
    def embed_batch(self, texts: List[str]) -> List[Tuple[List[int], List[float]]]:
        """Generate sparse embeddings for multiple texts
        
        Args:
            texts: List of input texts
            
        Returns:
            List of (indices, values) tuples
        """
        try:
            embeddings = list(self.model.embed(texts))
            results = []
            
            for embedding in embeddings:
                indices = embedding.indices.tolist()
                values = embedding.values.tolist()
                results.append((indices, values))
            
            return results
            
        except Exception as e:
            logger.error(f"Failed to generate batch embeddings: {e}")
            raise