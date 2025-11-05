# Jira Sync - 증분 업데이트 기반 Jira 동기화 도구

Jira 이슈를 Qdrant 벡터 데이터베이스에 동기화하는 통합 CLI 도구입니다. 하이브리드 검색(BM42 + Dense Embedding)을 지원하며, 증분 업데이트를 통해 처리 시간을 최소화합니다.

## 주요 기능

- ✅ **증분 업데이트**: 첫 실행 이후 변경된 이슈만 수집하여 처리 시간 단축
- ✅ **자동 타임스탬프 관리**: 마지막 동기화 시간 자동 추적
- ✅ **Upsert 방식**: 기존 이슈는 업데이트, 새 이슈는 삽입하여 중복 방지
- ✅ **하이브리드 검색**: Dense Vector (임베딩) + Sparse Vector (BM42) 동시 지원
- ✅ **구조화된 로깅**: JSON 형식의 구조화된 로그 출력 (slog)
- ✅ **재시도 로직**: 네트워크 오류 시 자동 재시도 (3번, 3초 간격)

## 아키텍처

```
cmd/jira-sync/
├── app/                    # 애플리케이션 레이어
│   ├── adapter/            # 외부 서비스 어댑터
│   │   ├── jira.go        # Jira API 클라이언트
│   │   ├── embedder.go    # 임베딩 생성
│   │   └── qdrant.go      # Qdrant 업로드
│   ├── domain/            # 도메인 모델
│   │   └── sync.go        # 동기화 도메인 타입
│   ├── service/           # 비즈니스 로직
│   │   └── sync_service.go
│   └── app.go             # 애플리케이션 진입점
├── python/                # BM42 Python 스크립트
│   ├── bm42_indexer.py
│   └── requirements.txt
└── main.go                # 메인 CLI
```

## 실행 흐름

### 1️⃣ 명령 시작

```bash
jira-sync incremental
```

**플래그**:
- `--data-dir`: 임시 파일 저장 경로 (기본값: `/data`)
- `--state-dir`: 타임스탬프 영구 저장 경로 (기본값: `/state`)

### 2️⃣ 타임스탬프 확인

```go
tm := timestamp.NewManager(stateDir)
lastSync, err := tm.GetLastSync()
```

**두 가지 경로**:

#### A. 타임스탬프 파일이 없는 경우 (첫 실행)
```
❌ /state/.last_sync_timestamp 없음
→ "No previous sync found, running full sync instead..."
→ RunFull() 함수로 전환
```

#### B. 타임스탬프 파일이 있는 경우 (정상 증분 업데이트)
```
✅ 마지막 동기화: "2025-10-21 14:51"
→ 현재 시간 기록: "2025-10-22 00:30"
→ 증분 업데이트 진행
```

### 3️⃣ Step 1: Jira 이슈 수집 (adapter/jira.go)

**JQL 쿼리 생성**:
```sql
project=GS AND updated >= "2025-10-21 14:51" ORDER BY updated DESC
```

**실행 과정**:
1. Jira REST API 호출 (100개씩 페이징)
2. 재시도 로직 적용 (최대 3번, 3초 간격)
3. 응답 파싱 및 메모리에 수집
4. `/data/issues.json` 파일로 저장

**로그 출력 예시**:
```json
{"time":"2025-10-22T00:30:15Z","level":"INFO","msg":"fetching issues updated since","since":"2025-10-21 14:51"}
{"time":"2025-10-22T00:30:17Z","level":"INFO","msg":"collected issues","count":100,"total":100}
{"time":"2025-10-22T00:30:19Z","level":"INFO","msg":"collected issues","count":50,"total":150}
{"time":"2025-10-22T00:30:20Z","level":"INFO","msg":"issues saved","path":"/data/issues.json","count":150}
```

### 4️⃣ Step 2: 임베딩 생성 (adapter/embedder.go)

**실행 명령**:
```bash
prototype embedding \
  --input /data/issues.json \
  --output /data/embedding.json \
  --api-url http://llama-server:8080/v1
```

**처리 과정**:
1. 각 이슈의 요약(summary)과 설명(description) 추출
2. OpenAI API 호출하여 텍스트 → 벡터 변환
3. 결과를 `/data/embedding.json`에 저장

**임베딩 파일 구조**:
```json
[
  {
    "id": "GS-1234",
    "vector": [0.123, -0.456, 0.789, ...],  // 1536차원
    "payload": {
      "key": "GS-1234",
      "summary": "EDR 보안 점검 기능 추가",
      "description": "..."
    }
  }
]
```

### 5️⃣ Step 3a: Dense Vector Upsert (adapter/qdrant.go)

**실행 명령**:
```bash
prototype insert \
  --file /data/embedding.json \
  --qdrant-host qdrant:6334 \
  --collection jira_issues
```

**Upsert 동작**:
- **이슈 키 → Point ID 매핑**: `hash("GS-1234") % 2^63`
- **기존 이슈**: 벡터 업데이트
- **새 이슈**: 새로운 포인트 삽입
- **중복 방지**: 동일한 이슈 키 = 동일한 ID

### 6️⃣ Step 3b: Sparse Vector (BM42) Upsert

**실행 명령**:
```bash
python3 /app/python/bm42_indexer.py \
  --input /data/issues.json \
  --host qdrant \
  --port 6333 \
  --collection jira_bm42_full
```

**BM42 처리 과정**:

1. **컬렉션 확인**:
   ```python
   collection_exists = client.get_collection("jira_bm42_full")
   if collection_exists:
       print("Collection already exists (will update)")
   else:
       client.create_collection(...)  # Sparse Vector 설정으로 생성
   ```

2. **Sparse Vector 생성**:
   - fastembed `Qdrant/bm42-all-minilm-l6-v2-attentions` 모델 사용
   - Attention 메커니즘으로 중요 토큰 추출
   - 각 토큰에 가중치 부여

   ```python
   from fastembed import SparseTextEmbedding
   model = SparseTextEmbedding(model_name="Qdrant/bm42-all-minilm-l6-v2-attentions")

   text = "EDR 보안 점검 기능 추가"
   embedding = model.embed([text])
   # → SparseVector(indices=[12, 45, 789, ...], values=[0.8, 0.6, 0.4, ...])
   ```

3. **해시 기반 ID로 Upsert**:
   ```python
   for doc in documents:
       doc_key = doc.get("key", f"doc_{i}")
       point_id = hash(doc_key) % (2**63 - 1)  # 안정적인 ID 생성

       client.upsert(
           collection_name=collection_name,
           points=[PointStruct(
               id=point_id,
               payload=metadata,
               vector={"bm42": sparse_vector}
           )]
       )
   ```

**로그 출력 예시**:
```
Collection already exists: jira_bm42_full (will update)
Processing 150 documents...
Generating BM42 embeddings...
Upserted 100/150 documents
Upserted 150/150 documents
Successfully indexed 150 documents
```

### 7️⃣ Step 4: 타임스탬프 업데이트

```go
tm.SaveLastSync(currentTime)  // "2025-10-22 00:30"
```

**저장 내용** (`/state/.last_sync_timestamp`):
```
2025-10-22 00:30
```

**다음 실행 시**:
- JQL 쿼리: `updated >= "2025-10-22 00:30"`
- 이 시간 이후 업데이트된 이슈만 수집

### 8️⃣ 완료

```json
{"time":"2025-10-22T00:30:45Z","level":"INFO","msg":"incremental synchronization completed successfully","issues":150,"duration":"30s"}
```

## 시퀀스 다이어그램

```
┌─────────────────────────────────────────┐
│ jira-sync incremental 시작              │
└────────────┬────────────────────────────┘
             │
   ┌─────────▼──────────┐
   │ 타임스탬프 확인    │
   │ (pkg/jira-sync/    │
   │  timestamp/)       │
   └────────┬───────────┘
            │
    ┌───────┴────────┐
    │                │
없음 │            있음 │
    │                │
┌───▼───┐      ┌─────▼──────────────┐
│ Full  │      │ JQL 쿼리 생성      │
│ Sync  │      │ updated >= "..."   │
└───────┘      └─────┬──────────────┘
                     │
          ┌──────────▼───────────────┐
          │ Jira API 호출            │
          │ (adapter/jira.go)        │
          │ - 페이징 (100개씩)       │
          │ - 재시도 (3번, 3초 간격) │
          │ → issues.json            │
          └──────────┬───────────────┘
                     │
          ┌──────────▼───────────────┐
          │ Embedding 생성           │
          │ (adapter/embedder.go)    │
          │ - prototype 바이너리     │
          │ - OpenAI API 호출        │
          │ → embedding.json         │
          └──────────┬───────────────┘
                     │
          ┌──────────▼───────────────┐
          │ Qdrant Upsert            │
          │ (adapter/qdrant.go)      │
          ├──────────────────────────┤
          │ 1. Dense Vectors         │
          │    - prototype insert    │
          │    - Port 6334           │
          ├──────────────────────────┤
          │ 2. BM42 Sparse Vectors   │
          │    - Python indexer      │
          │    - Port 6333           │
          └──────────┬───────────────┘
                     │
          ┌──────────▼───────────────┐
          │ 타임스탬프 저장          │
          │ (현재 시간)              │
          └──────────┬───────────────┘
                     │
              ┌──────▼──────┐
              │   완료 ✅    │
              └─────────────┘
```

## 명령어

### Full Sync (전체 동기화)

```bash
jira-sync full
```

**사용 시기**:
- 첫 실행
- 전체 재색인이 필요한 경우
- 타임스탬프 파일이 손상된 경우

### Incremental Sync (증분 업데이트)

```bash
jira-sync incremental
```

**사용 시기**:
- 일반적인 크론잡 실행
- 마지막 동기화 이후 변경사항만 반영

## 환경 변수

| 변수 | 필수 | 기본값 | 설명 |
|------|------|--------|------|
| `JIRA_API_TOKEN` | ✅ | - | Jira API 토큰 |
| `JIRA_SERVER` | ✅ | - | Jira 서버 URL (예: `https://jira.company.com/rest/api/2/search`) |
| `JIRA_PROJECT_KEY` | ✅ | - | Jira 프로젝트 키 (예: `GS`) |
| `EMBEDDING_API_URL` | ✅ | - | 임베딩 API URL (예: `http://llama-server:8080/v1`) |
| `QDRANT_HOST` | ✅ | - | Qdrant 호스트 (예: `qdrant`) |
| `QDRANT_PORT` | ✅ | - | Qdrant 포트 (예: `6333`) |
| `COLLECTION_NAME` | ❌ | `jira_issues` | Dense vector 컬렉션 이름 |
| `BM42_COLLECTION` | ❌ | `jira_bm42_full` | Sparse vector 컬렉션 이름 |
| `DATA_DIR` | ❌ | `/data` | 임시 파일 저장 경로 |
| `STATE_DIR` | ❌ | `/state` | 타임스탬프 저장 경로 |

## 로컬 실행

```bash
# 빌드
go build -o jira-sync ./cmd/jira-sync

# 환경 변수 설정
export JIRA_API_TOKEN="your-token"
export JIRA_SERVER="https://jira.company.com/rest/api/2/search"
export JIRA_PROJECT_KEY="GS"
export EMBEDDING_API_URL="http://localhost:8080/v1"
export QDRANT_HOST="localhost"
export QDRANT_PORT="6333"
export COLLECTION_NAME="jira_issues"
export BM42_COLLECTION="jira_bm42_full"

# 전체 동기화 (첫 실행)
./jira-sync full

# 증분 업데이트
./jira-sync incremental
```

## Docker 실행

```bash
# 빌드
docker build -f build/jira-sync/Dockerfile -t jira-sync:latest .

# 실행
docker run --rm \
  --env-file .env \
  -v /path/to/state:/state \
  jira-sync:latest \
  jira-sync incremental
```

## Kubernetes CronJob

전체 구성은 `/cronjob` 디렉토리 참조:

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: jira-sync-cronjob
spec:
  schedule: "0 17 * * *"  # 매일 새벽 2시 (KST)
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: jira-sync
            image: your-registry/jira-sync:latest
            command: ["jira-sync", "incremental"]
            env:
              - name: DATA_DIR
                value: "/data"
              - name: STATE_DIR
                value: "/state"
              # ... (환경 변수들)
            volumeMounts:
              - name: data
                mountPath: /data
              - name: state
                mountPath: /state
          volumes:
            - name: data
              emptyDir: {}
            - name: state
              persistentVolumeClaim:
                claimName: jira-sync-state
```

## 성능 최적화

### 첫 실행 (Full Sync)
- 11,093개 이슈 수집
- 예상 시간: ~10분
- 네트워크: ~110 API 요청

### 이후 실행 (Incremental)
- 변경된 이슈만 수집 (예: 150개)
- 예상 시간: ~30초
- 네트워크: ~2 API 요청
- **98.6% 시간 단축** 🚀

## 특징

### ✅ 멱등성 (Idempotency)
- 같은 이슈를 여러 번 실행해도 안전
- Upsert 방식으로 중복 생성 방지
- 이슈 키 기반 안정적인 ID 생성

### ⏱️ 증분 처리
- 타임스탬프 기반 변경 감지
- JQL 필터로 API 요청 최소화
- 불필요한 임베딩 재생성 방지

### 🔄 재시도 로직
- 네트워크 오류 시 자동 재시도 (3번)
- 지수 백오프 (3초 대기)
- Jira API, Embedding API 모두 적용

### 💾 영구 저장
- `/state` → PersistentVolume (타임스탬프)
- `/data` → emptyDir (임시 파일, Job 종료 시 삭제)

## 트러블슈팅

### "no previous sync found"
- 첫 실행이거나 타임스탬프 파일이 없는 경우
- 자동으로 Full Sync 실행됨
- `/state/.last_sync_timestamp` 파일 확인

### "JIRA_API_TOKEN is not set"
- 환경 변수가 설정되지 않음
- ConfigMap과 Secret 설정 확인

### "HTTP 400/401 error"
- Jira URL 또는 토큰 확인
- URL에 `/search` 엔드포인트 포함 필요

### "collection does not exist"
- Qdrant 컬렉션이 생성되지 않음
- BM42 indexer가 자동으로 생성함 (첫 실행)

## 라이선스

MIT
