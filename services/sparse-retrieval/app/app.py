"""
Application Main Logic
"""
import os
import signal
import sys
import logging
import threading

from app.adapter.qdrant import QdrantClient
from app.adapter.bm42 import BM42Embedder
from app.service.service import PassageRetrievalService
from app.service.server import PassageRetrievalServer

logger = logging.getLogger(__name__)

shutdown_event = threading.Event()


def run():
    """Sparse retrieval service 실행"""
    qdrant_client = None
    server = None

    try:
        # 환경 변수에서 설정 가져오기
        grpc_port = int(os.getenv("GRPC_PORT", "50051"))
        qdrant_url = os.getenv("QDRANT_URL")
        qdrant_host = os.getenv("QDRANT_HOST", "localhost")
        qdrant_port = int(os.getenv("QDRANT_PORT", "6333"))
        qdrant_api_key = os.getenv("QDRANT_API_KEY")
        collection_name = os.getenv("COLLECTION_NAME", "jira_bm42_full")
        max_workers = int(os.getenv("MAX_WORKERS", "10"))

        # 어댑터 초기화
        logger.info("Initializing adapters...")

        # Qdrant 클라이언트 생성
        if qdrant_url:
            logger.info(f"Connecting to Qdrant URL: {qdrant_url}")
            qdrant_client = QdrantClient(url=qdrant_url, api_key=qdrant_api_key)
        else:
            logger.info(f"Connecting to Qdrant at {qdrant_host}:{qdrant_port}")
            qdrant_client = QdrantClient(host=qdrant_host, port=qdrant_port)

        # BM42 embedder 생성
        bm42_embedder = BM42Embedder()

        # 서비스 생성
        logger.info(f"Creating service with collection: {collection_name}")
        service = PassageRetrievalService(
            retriever=qdrant_client,
            embedder=bm42_embedder,
            collection_name=collection_name,
        )

        # 서버 생성 및 시작
        server = PassageRetrievalServer(
            service=service, port=grpc_port, max_workers=max_workers
        )

        # 시그널 핸들러 설정
        def signal_handler(signum, _):
            logger.info(f"시그널 {signum} 수신, 정상 종료 시작...")
            shutdown_event.set()
            
            # 유예 기간과 함께 서버 중지
            if server:
                logger.info("gRPC 서버 중지 중...")
                server.stop(grace=5)
            
            # 리소스 정리
            resources_to_close = []
            if qdrant_client:
                resources_to_close.append(("Qdrant 클라이언트", qdrant_client))
            
            for name, resource in resources_to_close:
                try:
                    if hasattr(resource, 'close'):
                        resource.close()
                        logger.info(f"{name} 종료됨")
                except Exception as e:
                    logger.error(f"{name} 종료 중 오류: {e}")
            
            logger.info("정상 종료 완료")
            sys.exit(0)

        signal.signal(signal.SIGINT, signal_handler)
        signal.signal(signal.SIGTERM, signal_handler)

        # 서버 시작
        server.start()
        server.wait_for_termination()

    except Exception as e:
        logger.error(f"Sparse retrieval service에서 치명적인 오류 발생: {e}", exc_info=True)

        # 오류 시 정리
        if qdrant_client:
            try:
                qdrant_client.close()
            except:
                pass

        if server:
            try:
                server.stop(0)
            except:
                pass

        sys.exit(1)