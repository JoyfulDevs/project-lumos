"""
gRPC Server Implementation
"""
import logging
import sys
import os
from concurrent import futures

import grpc

# 생성된 protobuf 경로 추가
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../../../gen/python'))

from retrieval.passage.v1 import service_pb2, service_pb2_grpc
from app.service.service import PassageRetrievalService

logger = logging.getLogger(__name__)


class PassageRetrievalServicer(service_pb2_grpc.PassageRetrievalServiceServicer):
    """Passage 검색을 위한 gRPC 서비스 구현"""
    
    def __init__(self, service: PassageRetrievalService):
        """gRPC servicer 초기화
        
        Args:
            service: passage 검색 서비스 인스턴스
        """
        self.service = service
        logger.info("PassageRetrievalServicer 초기화됨")
    
    def Retrieve(self, request, context):
        """Passage 검색 요청 처리
        
        Args:
            request: 클라이언트로부터의 RetrieveRequest
            context: gRPC context
            
        Returns:
            검색된 passage를 포함한 RetrieveResponse
        """
        try:
            query = request.query
            limit = request.limit if request.limit > 0 else 10
            
            # 입력 검증
            if not query or not query.strip():
                logger.warning("빈 query 수신")
                context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                context.set_details("Query는 비어있을 수 없습니다")
                return service_pb2.RetrieveResponse()  # 빈 응답 반환
            
            if limit > 100:
                logger.warning(f"제한 {limit}이 최대값 초과, 100으로 제한")
                limit = 100
            
            logger.info(f"검색 요청 수신: query='{query[:50] if query else ''}...', limit={limit}")
            
            # passage 검색
            passages = self.service.retrieve(query, limit)
            
            # 응답 구성
            response = service_pb2.RetrieveResponse()
            
            for passage in passages:
                proto_passage = response.passages.add()
                proto_passage.score = passage["score"]
                # content를 byte로 인코딩
                proto_passage.content = passage["content"].encode('utf-8')
            
            logger.info(f"{len(response.passages)}개의 passage 반환")
            return response
            
        except ValueError as e:
            logger.error(f"검색 검증 오류: {e}")
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, str(e))
        except Exception as e:
            logger.error(f"검색 오류: {e}", exc_info=True)
            context.abort(grpc.StatusCode.INTERNAL, str(e))

    def Health(self, request, context):
        """gRPC health check 구현"""
        try:
            health_checks = {}
            overall_healthy = True
            
            # Qdrant 연결 확인
            try:
                collection_info = self.service.retriever.get_collection_info(
                    self.service.collection_name
                )
                health_checks['qdrant'] = {
                    'status': 'healthy',
                    'vectors_count': collection_info.get('vectors_count', 0)
                }
            except Exception as e:
                overall_healthy = False
                health_checks['qdrant'] = {
                    'status': 'unhealthy',
                    'error': str(e)
                }
                logger.error(f"Qdrant health check 실패: {e}")
            
            # BM42 모델 상태 확인
            if hasattr(self.service.embedder, 'model') and self.service.embedder.model is not None:
                health_checks['bm42'] = {'status': 'healthy'}
            else:
                overall_healthy = False
                health_checks['bm42'] = {
                    'status': 'unhealthy',
                    'error': 'Model not loaded'
                }
                logger.error("BM42 model이 로드되지 않음")
            
            if not overall_healthy:
                context.abort(
                    grpc.StatusCode.UNAVAILABLE, 
                    f"서비스 비정상: {health_checks}"
                )
            
            return service_pb2.HealthResponse(status="healthy")
        except Exception as e:
            logger.error(f"Health check 오류: {e}", exc_info=True)
            context.abort(grpc.StatusCode.INTERNAL, str(e))


class PassageRetrievalServer:
    """Passage 검색을 위한 gRPC 서버"""
    
    def __init__(self, service: PassageRetrievalService, 
                 port: int = 50051, max_workers: int = 10):
        """gRPC 서버 초기화
        
        Args:
            service: passage 검색 서비스
            port: 수신 포트
            max_workers: 최대 worker thread 수
        """
        self.service = service
        self.port = port
        self.max_workers = max_workers
        self.server = None

        logger.info(f"PassageRetrievalServer가 포트 {port}에 구성됨")
    
    def start(self):
        """gRPC 서버 시작"""
        try:
            # 서버 생성
            self.server = grpc.server(
                futures.ThreadPoolExecutor(max_workers=self.max_workers)
            )
            
            # servicer 추가
            service_pb2_grpc.add_PassageRetrievalServiceServicer_to_server(
                PassageRetrievalServicer(self.service),
                self.server
            )
            
            # 포트에 바인딩
            self.server.add_insecure_port(f'[::]:{self.port}')
            
            # 서버 시작
            self.server.start()
            logger.info(f"Sparse retrieval service가 포트 {self.port}에서 시작됨")
            
        except Exception as e:
            logger.error(f"서버 시작 실패: {e}")
            raise
    
    def wait_for_termination(self):
        """서버 종료 대기"""
        if self.server:
            self.server.wait_for_termination()
    
    def stop(self, grace=5):
        """gRPC 서버 중지
        
        Args:
            grace: 종료 유예 기간(초)
        """
        if self.server:
            logger.info("Sparse retrieval service 중지 중...")
            self.server.stop(grace)
            logger.info("Sparse retrieval service 중지됨")