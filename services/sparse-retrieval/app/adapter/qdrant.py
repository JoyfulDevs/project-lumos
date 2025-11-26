"""
Qdrant Client Adapter for Sparse Vector Search
"""
import logging
import time
from typing import List, Dict, Any, Optional, Tuple
from qdrant_client import QdrantClient as QdrantClientBase
from qdrant_client.models import SparseVector

logger = logging.getLogger(__name__)


class QdrantClient:
    """Sparse vector 연산을 위한 Qdrant 클라이언트 어댑터"""
    
    def __init__(self, url: str = None, host: str = "localhost", 
                 port: int = 6333, api_key: str = None):
        """Qdrant 클라이언트 초기화
        
        Args:
            url: Qdrant의 전체 URL (프로덕션용)
            host: Qdrant 호스트 (로컬 개발용)
            port: Qdrant 포트 (로컬 개발용)
            api_key: 보안 인스턴스의 API 키
        """
        try:
            if url:
                # URL을 사용한 프로덕션 모드
                self.client = QdrantClientBase(
                    url=url,
                    api_key=api_key,
                    https=url.startswith("https"),
                    timeout=30
                )
                logger.info(f"Qdrant URL에 연결됨: {url}")
            else:
                # 로컬 개발 모드
                self.client = QdrantClientBase(host=host, port=port)
                logger.info(f"Qdrant {host}:{port}에 연결됨")
                
        except Exception as e:
            logger.error(f"Qdrant 연결 실패: {e}")
            raise
    
    def search_sparse(self, collection_name: str, 
                     indices: List[int], values: List[float],
                     limit: int = 10, retry_count: int = 3) -> List[Dict[str, Any]]:
        """재시도 로직을 포함한 sparse vector 검색
        
        Args:
            collection_name: 검색할 컬렉션 이름
            indices: sparse vector indices
            values: sparse vector values
            limit: 최대 결과 수
            retry_count: 실패 시 재시도 횟수
            
        Returns:
            id, score, payload를 포함한 검색 결과 목록
        """
        last_error = None

        for attempt in range(retry_count):
            try:
                # sparse vector 생성
                sparse_vector = SparseVector(
                    indices=indices,
                    values=values
                )

                # 검색 수행
                # Note: qdrant-client v1.10.0+ 에서는 query_points를 사용합니다.
                # v1.16.0+ 에서는 query 메서드 사용을 권장합니다.
                results = self.client.query_points(
                    collection_name=collection_name,
                    query=sparse_vector,
                    using="bm42",
                    limit=limit,
                    with_payload=True
                ).points

                # 결과 포맷팅
                formatted_results = []
                for point in results:
                    result = {
                        "id": str(point.id),
                        "score": float(point.score),
                        "payload": point.payload or {}
                    }
                    formatted_results.append(result)

                logger.debug(f"검색이 {len(formatted_results)}개의 결과 반환")
                return formatted_results

            except Exception as e:
                last_error = e
                if attempt < retry_count - 1:
                    wait_time = 2 ** attempt  # exponential backoff
                    logger.warning(
                        f"검색 실패 (시도 {attempt + 1}/{retry_count}), "
                        f"{wait_time}초 후 재시도: {e}"
                    )
                    time.sleep(wait_time)
                else:
                    logger.error(f"{retry_count}회 시도 후 검색 실패: {e}")

        raise last_error
    
    def get_collection_info(self, collection_name: str) -> Dict[str, Any]:
        """컬렉션 정보 가져오기
        
        Args:
            collection_name: 컬렉션 이름
            
        Returns:
            컬렉션 정보
        """
        try:
            info = self.client.get_collection(collection_name)
            return {
                "vectors_count": info.points_count,
                "status": info.status,
                "config": info.config
            }
        except Exception as e:
            logger.error(f"컬렉션 정보 가져오기 실패: {e}")
            raise
    
    def health_check(self) -> bool:
        """Qdrant 상태 확인
        
        Returns:
            정상이면 True, 아니면 False
        """
        try:
            # 헬스체크로 컬렉션 목록 시도
            self.client.get_collections()
            return True
        except Exception as e:
            logger.error(f"헬스체크 실패: {e}")
            return False

    def close(self):
        """Qdrant 클라이언트 연결 종료"""
        try:
            # Qdrant 클라이언트에는 명시적인 close 메서드가 없음
            # 참조만 지움
            if hasattr(self, 'client'):
                self.client = None
                logger.info("Qdrant 연결 종료")
        except Exception as e:
            logger.error(f"Qdrant 연결 종료 실패: {e}")