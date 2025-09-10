# Sparse Retrieval Service

BM42 기반 Sparse Vector 검색을 제공하는 gRPC 마이크로서비스입니다.

## 개요

이 서비스는 BM42 (Best Matching 42) 알고리즘을 사용하여 텍스트 검색을 수행합니다. Attention 메커니즘 기반의 sparse embedding을 생성하고 Qdrant 벡터 데이터베이스를 통해 고속 검색을 제공합니다.

## 아키텍처

```
services/sparse-retrieval-service/
├── main.py                 # 서비스 진입점
├── app/
│   ├── app.py             # 애플리케이션 초기화 및 설정
│   ├── adapter/           # 외부 서비스 어댑터
│   │   ├── bm42.py       # BM42 임베딩 생성
│   │   └── qdrant.py     # Qdrant 클라이언트
│   └── service/          # 비즈니스 로직
│       ├── service.py    # 검색 서비스 구현
│       └── server.py     # gRPC 서버
├── requirements.txt       # Python 의존성
├── Dockerfile            # Docker 이미지 빌드
├── .dockerignore         # Docker 빌드 제외 파일
├── .env.local            # 로컬 환경 설정
├── .env.production       # 프로덕션 환경 설정
├── pyproject.toml        # Python 프로젝트 설정
├── .flake8               # Flake8 린터 설정
├── .gitignore            # Git 제외 파일
└── test/                 # 테스트 스크립트
    └── test_service.py
```

## 주요 기능

- **BM42 Sparse Embedding**: Attention 기반 중요 토큰 추출
- **gRPC API**: Protocol Buffer 기반 고성능 통신
- **Qdrant 통합**: Sparse vector 저장 및 검색
- **환경별 설정**: 로컬/프로덕션 자동 전환
- **Docker 지원**: 컨테이너 기반 배포

## 환경 변수

| 변수명 | 기본값 | 설명 |
|--------|--------|------|
| `GRPC_PORT` | 50051 | gRPC 서버 포트 |
| `QDRANT_URL` | - | Qdrant 서버 URL (프로덕션) |
| `QDRANT_HOST` | localhost | Qdrant 호스트 (로컬) |
| `QDRANT_PORT` | 6333 | Qdrant 포트 (로컬) |
| `QDRANT_API_KEY` | - | Qdrant API 키 (선택) |
| `COLLECTION_NAME` | jira_bm42_full | 검색 대상 컬렉션 |
| `MAX_WORKERS` | 10 | gRPC 워커 스레드 수 |

## 빠른 시작

### 1. Docker로 실행

```bash
# 이미지 빌드
docker build -f cmd/sparse-retrieval-service/Dockerfile -t sparse-retrieval-service:latest .

# 컨테이너 실행
docker run -d --name sparse-retrieval-service \
  -p 50051:50051 \
  -e QDRANT_HOST=host.docker.internal \
  -e QDRANT_PORT=6333 \
  -e COLLECTION_NAME=jira_bm42_full \
  sparse-retrieval-service:latest
```

### 2. 로컬 개발

```bash
# 의존성 설치
pip install -r requirements.txt

# 환경 변수 설정
export GRPC_PORT=50051
export QDRANT_HOST=localhost
export QDRANT_PORT=6333
export COLLECTION_NAME=jira_bm42_full

# 서비스 실행
python main.py
```

## API 사양

### PassageRetrievalService

#### Retrieve RPC

텍스트 쿼리를 받아 관련 문서를 검색합니다.

**Request:**
```protobuf
message RetrieveRequest {
  string query = 1;  // 검색 쿼리
  int32 limit = 2;   // 최대 결과 수
}
```

**Response:**
```protobuf
message RetrieveResponse {
  repeated Passage passages = 1;
}

message Passage {
  float score = 1;    // 관련성 점수
  bytes content = 2;  // 문서 내용
}
```

## 모니터링

### 로그 확인

```bash
# Docker 로그
docker logs -f sparse-retrieval-service

# Docker Compose 로그
docker-compose logs -f sparse-retrieval
```

### 헬스체크

```bash
# 컨테이너 상태
docker inspect sparse-retrieval-service --format='{{.State.Health.Status}}'

# gRPC 연결 테스트
grpcurl -plaintext localhost:50051 list
```

## 트러블슈팅

### Qdrant 연결 실패

```bash
# Qdrant 상태 확인
curl http://localhost:6333/collections

# 네트워크 연결 확인
docker exec sparse-retrieval-service ping host.docker.internal
```
