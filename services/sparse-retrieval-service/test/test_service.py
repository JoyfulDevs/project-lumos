#!/usr/bin/env python3
"""
Sparse Retrieval Service 테스트 스크립트
"""
import grpc
import sys
import os
import time

# protobuf 경로 추가
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../../gen/python'))

from retrieval.passage.v1 import service_pb2, service_pb2_grpc


def test_connection(host='localhost', port=50051):
    """gRPC 연결 테스트"""
    print(f"\n=== gRPC 연결 테스트 (host={host}, port={port}) ===")
    channel = grpc.insecure_channel(f'{host}:{port}')
    
    try:
        # 연결 시도 (timeout 5초)
        grpc.channel_ready_future(channel).result(timeout=5)
        print("✅ gRPC 서버 연결 성공")
        return channel
    except grpc.FutureTimeoutError:
        print("❌ gRPC 서버 연결 실패: 서버가 응답하지 않습니다")
        return None
    except Exception as e:
        print(f"❌ gRPC 서버 연결 실패: {e}")
        return None


def test_retrieve(channel, query="보안 점검"):
    """Retrieve 기능 테스트"""
    print(f"\n=== Retrieve 기능 테스트 ===")
    print(f"Query: {query}")
    
    stub = service_pb2_grpc.PassageRetrievalServiceStub(channel)
    request = service_pb2.RetrieveRequest(
        query=query,
        limit=5
    )
    
    try:
        response = stub.Retrieve(request, timeout=10)
        print(f"✅ Retrieve 성공: {len(response.passages)}개의 passage 반환")
        
        for i, passage in enumerate(response.passages[:3], 1):
            content_preview = passage.content.decode('utf-8')[:100] if passage.content else "No content"
            print(f"\n  Passage {i}:")
            print(f"    Score: {passage.score:.4f}")
            print(f"    Content: {content_preview}...")
        
        return True
    except grpc.RpcError as e:
        print(f"❌ Retrieve 실패: {e.code()} - {e.details()}")
        return False
    except Exception as e:
        print(f"❌ Retrieve 실패: {e}")
        return False


def test_empty_query(channel):
    """빈 쿼리 테스트 (에러 핸들링 확인)"""
    print(f"\n=== 빈 쿼리 테스트 ===")
    
    stub = service_pb2_grpc.PassageRetrievalServiceStub(channel)
    request = service_pb2.RetrieveRequest(
        query="",
        limit=5
    )
    
    try:
        response = stub.Retrieve(request, timeout=5)
        print("❌ 빈 쿼리가 허용됨 (예상: 에러)")
        return False
    except grpc.RpcError as e:
        if e.code() == grpc.StatusCode.INVALID_ARGUMENT:
            print(f"✅ 빈 쿼리 거부됨: {e.details()}")
            return True
        else:
            print(f"❌ 예상치 못한 에러: {e.code()} - {e.details()}")
            return False


def test_large_limit(channel):
    """큰 limit 값 테스트"""
    print(f"\n=== 큰 limit 값 테스트 ===")
    
    stub = service_pb2_grpc.PassageRetrievalServiceStub(channel)
    request = service_pb2.RetrieveRequest(
        query="테스트 쿼리",
        limit=200  # 최대값 100 초과
    )
    
    try:
        response = stub.Retrieve(request, timeout=10)
        actual_count = len(response.passages)
        print(f"✅ 요청 처리됨: {actual_count}개의 passage 반환")
        if actual_count <= 100:
            print("✅ Limit이 100으로 제한됨")
            return True
        else:
            print(f"❌ Limit 제한 실패: {actual_count}개 반환됨")
            return False
    except grpc.RpcError as e:
        print(f"❌ 테스트 실패: {e.code()} - {e.details()}")
        return False


def test_health_check(channel):
    """Health check 테스트 - Retrieve로 간접 확인"""
    print(f"\n=== Health Check 테스트 ===")
    
    # Health 엔드포인트가 proto에 정의되어 있지 않으므로
    # 간단한 쿼리로 서비스 상태 확인
    stub = service_pb2_grpc.PassageRetrievalServiceStub(channel)
    request = service_pb2.RetrieveRequest(
        query="health check test",
        limit=1
    )
    
    try:
        response = stub.Retrieve(request, timeout=5)
        print(f"✅ 서비스 정상 작동 (health check via retrieve)")
        return True
    except grpc.RpcError as e:
        if e.code() == grpc.StatusCode.UNAVAILABLE:
            print(f"⚠️  서비스 비정상: {e.details()}")
        else:
            print(f"❌ Health check 실패: {e.code()} - {e.details()}")
        return False


def main():
    """메인 테스트 실행"""
    print("=" * 50)
    print("Sparse Retrieval Service 테스트 시작")
    print("=" * 50)
    
    # 환경 변수에서 설정 읽기
    host = os.getenv('GRPC_HOST', 'localhost')
    port = int(os.getenv('GRPC_PORT', '50051'))
    
    # 연결 테스트
    channel = test_connection(host, port)
    if not channel:
        print("\n❌ 서버에 연결할 수 없습니다. 서버가 실행 중인지 확인하세요.")
        print("\n서버 시작 방법:")
        print("  cd cmd/sparse-retrieval-service")
        print("  python main.py")
        return
    
    # 각 테스트 실행
    results = []
    
    # Health check
    results.append(("Health Check", test_health_check(channel)))
    
    # 정상 retrieve
    results.append(("Retrieve (정상)", test_retrieve(channel, "보안 점검")))
    
    # 빈 쿼리
    results.append(("빈 쿼리 처리", test_empty_query(channel)))
    
    # 큰 limit
    results.append(("큰 limit 처리", test_large_limit(channel)))
    
    # 다른 쿼리 테스트
    test_queries = [
        "보안 점검 기능",
        "Live Response",
        "EDR 보안 점검 기능"
    ]
    
    print(f"\n=== 추가 쿼리 테스트 ===")
    for query in test_queries:
        success = test_retrieve(channel, query)
        results.append((f"Query: {query}", success))
    
    # 결과 요약
    print("\n" + "=" * 50)
    print("테스트 결과 요약")
    print("=" * 50)
    
    for test_name, success in results:
        status = "✅ PASS" if success else "❌ FAIL"
        print(f"{status}: {test_name}")
    
    passed = sum(1 for _, success in results if success)
    total = len(results)
    print(f"\n전체: {passed}/{total} 테스트 통과")
    
    # 채널 닫기
    channel.close()


if __name__ == "__main__":
    main()