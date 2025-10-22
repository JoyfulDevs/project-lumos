#!/usr/bin/env python3
"""
BM42 Indexer for Qdrant
Uses fastembed library with BM42 model for sparse vector indexing
"""

import json
from typing import List, Dict, Any
from fastembed import SparseTextEmbedding
from qdrant_client import QdrantClient
from qdrant_client.models import (
    PointStruct,
    SparseVector,
    Distance,
    SparseVectorParams,
    SparseIndexParams,
    Modifier
)
import argparse
from tqdm import tqdm


class BM42Indexer:
    def __init__(self, qdrant_host: str = "localhost", qdrant_port: int = 6333):
        """Initialize BM42 indexer with Qdrant client and embedding model"""
        self.client = QdrantClient(host=qdrant_host, port=qdrant_port)
        self.model = SparseTextEmbedding(
            model_name="Qdrant/bm42-all-minilm-l6-v2-attentions"
        )

    def create_collection(self, collection_name: str, recreate: bool = False):
        """Create a Qdrant collection with BM42 sparse vector configuration

        Args:
            collection_name: Name of the collection
            recreate: If True, delete existing collection first
        """
        collection_exists = False
        try:
            self.client.get_collection(collection_name)
            collection_exists = True
        except:
            pass

        if collection_exists:
            if recreate:
                print(f"Deleting existing collection: {collection_name}")
                self.client.delete_collection(collection_name)
            else:
                print(f"Collection already exists: {collection_name} (will update)")
                return

        # Create collection with sparse vectors (vectors_config can be empty dict for sparse-only)
        self.client.create_collection(
            collection_name=collection_name,
            vectors_config={},  # Empty for sparse-only collection
            sparse_vectors_config={
                "bm42": SparseVectorParams(
                    index=SparseIndexParams(
                        on_disk=False,
                    ),
                    modifier=Modifier.IDF,
                )
            }
        )
        print(f"Created collection: {collection_name}")

    def load_documents(self, file_path: str) -> List[Dict[str, Any]]:
        """Load documents from JSON file"""
        with open(file_path, 'r', encoding='utf-8') as f:
            documents = json.load(f)
        return documents

    def extract_text(self, doc: Dict[str, Any]) -> str:
        """Extract text from document based on format"""
        parts = []

        # Handle Jira format
        if 'fields' in doc:
            fields = doc['fields']
            if 'summary' in fields:
                parts.append(fields['summary'])
            if 'description' in fields:
                parts.append(fields['description'])
            # Add comments if available
            if 'comment' in fields and 'comments' in fields['comment']:
                for comment in fields['comment']['comments']:
                    if 'body' in comment:
                        parts.append(comment['body'])

        # Handle key
        if 'key' in doc:
            parts.append(doc['key'])

        return ' '.join(filter(None, parts))

    def index_documents(self, collection_name: str, documents: List[Dict[str, Any]], batch_size: int = 32):
        """Index documents using BM42 sparse embeddings"""
        points = []
        texts = []
        metadata = []
        doc_keys = []

        print(f"Processing {len(documents)} documents...")

        for i, doc in enumerate(documents):
            text = self.extract_text(doc)
            texts.append(text)

            # Use issue key for stable ID (hash it to get numeric ID)
            doc_key = doc.get("key", f"doc_{i}")
            doc_keys.append(doc_key)

            # Prepare metadata
            meta = {
                "key": doc_key,
                "title": doc.get("fields", {}).get("summary", "") if "fields" in doc else doc.get("summary", ""),
                "content": text[:1000]  # Store first 1000 chars
            }
            metadata.append(meta)

        # Generate embeddings in batches
        print("Generating BM42 embeddings...")
        embeddings = list(self.model.embed(texts, batch_size=batch_size))

        # Create points for Qdrant
        for i, (embedding, meta, doc_key) in enumerate(zip(embeddings, metadata, doc_keys)):
            # Use hash of key as numeric ID for upsert
            point_id = hash(doc_key) % (2**63 - 1)  # Ensure positive 64-bit int

            point = PointStruct(
                id=point_id,
                payload=meta,
                vector={
                    "bm42": SparseVector(
                        indices=embedding.indices.tolist(),
                        values=embedding.values.tolist()
                    )
                }
            )
            points.append(point)

            # Batch upload with upsert
            if len(points) >= 100 or i == len(embeddings) - 1:
                self.client.upsert(
                    collection_name=collection_name,
                    points=points
                )
                print(f"Upserted {i+1}/{len(documents)} documents")
                points = []

        print(f"Successfully indexed {len(documents)} documents")


def main():
    parser = argparse.ArgumentParser(description="Index documents using BM42")
    parser.add_argument("--input", "-i", required=True, help="Input JSON file")
    parser.add_argument("--collection", "-c", default="jira_bm42", help="Collection name")
    parser.add_argument("--host", default="localhost", help="Qdrant host")
    parser.add_argument("--port", type=int, default=6333, help="Qdrant port")
    parser.add_argument("--recreate", action="store_true", help="Recreate collection (default: update existing)")

    args = parser.parse_args()

    # Initialize indexer
    indexer = BM42Indexer(args.host, args.port)

    # Create collection (or skip if exists and not recreating)
    indexer.create_collection(args.collection, recreate=args.recreate)

    # Load and index documents
    documents = indexer.load_documents(args.input)
    indexer.index_documents(args.collection, documents)


if __name__ == "__main__":
    main()
