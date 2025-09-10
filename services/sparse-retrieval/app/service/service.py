"""
Passage Retrieval Service Implementation
"""
import logging
from typing import List, Dict, Any

from app.adapter.bm42 import BM42Embedder
from app.adapter.qdrant import QdrantClient

logger = logging.getLogger(__name__)


class PassageRetrievalService:
    """BM42 sparse embedding을 사용한 passage 검색 서비스"""
    
    def __init__(self, retriever: QdrantClient, embedder: BM42Embedder, 
                 collection_name: str):
        """Passage 검색 서비스 초기화
        
        Args:
            retriever: vector 검색을 위한 Qdrant 클라이언트
            embedder: sparse vector 생성을 위한 BM42 embedder
            collection_name: 검색할 collection 이름
        """
        self.retriever = retriever
        self.embedder = embedder
        self.collection_name = collection_name
        
        logger.info(f"Passage 검색 서비스 초기화됨, collection: {collection_name}")
        
        # collection 존재 여부 확인 (Qdrant 버전 호환성 문제로 실패할 수 있음)
        try:
            info = retriever.get_collection_info(collection_name)
            logger.info(f"Collection '{collection_name}'에 {info['vectors_count']}개의 vector 존재")
        except Exception as e:
            # Qdrant client와 서버 버전 호환성 문제일 수 있으므로 warning으로 처리
            logger.warning(f"Collection '{collection_name}' 정보 확인 실패 (서비스는 계속 실행됨): {e}")
    
    def retrieve(self, query: str, limit: int = 10) -> List[Dict[str, Any]]:
        """쿼리에 대한 passage 검색
        
        Args:
            query: 검색 query 텍스트
            limit: 최대 결과 수
            
        Returns:
            score와 metadata를 포함한 검색된 passage 목록
        """
        try:
            logger.debug(f"Query 처리 중: '{query[:100]}...'")
            
            # query에 대한 sparse embedding 생성
            indices, values = self.embedder.embed(query)
            logger.debug(f"{len(indices)}개의 non-zero element로 sparse vector 생성됨")
            
            # Qdrant에서 검색
            results = self.retriever.search_sparse(
                collection_name=self.collection_name,
                indices=indices,
                values=values,
                limit=limit
            )
            
            # 결과 포맷팅
            passages = []
            for result in results:
                passage = {
                    "id": result["id"],
                    "score": result["score"],
                    "content": result["payload"].get("content", ""),
                    "metadata": {}
                }
                
                # metadata 필드 추출
                payload = result["payload"]
                if "key" in payload:
                    passage["metadata"]["key"] = payload["key"]
                if "title" in payload:
                    passage["metadata"]["title"] = payload["title"]
                    
                # 기타 metadata 필드 추가
                for key, value in payload.items():
                    if key not in ["content", "key", "title"]:
                        passage["metadata"][key] = str(value)
                
                passages.append(passage)
            
            logger.info(f"Query에 대해 {len(passages)}개의 passage 검색됨")
            return passages
            
        except Exception as e:
            logger.error(f"Passage 검색 실패: {e}")
            raise
    
    def batch_retrieve(self, queries: List[str], limit: int = 10) -> List[List[Dict[str, Any]]]:
        """여러 query에 대한 passage 검색
        
        Args:
            queries: 검색 query 목록
            limit: query당 최대 결과 수
            
        Returns:
            각 query에 대한 결과 목록
        """
        try:
            results = []
            for query in queries:
                passages = self.retrieve(query, limit)
                results.append(passages)
            return results
            
        except Exception as e:
            logger.error(f"Batch passage 검색 실패: {e}")
            raise