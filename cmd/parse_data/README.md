# JIRA 임베딩 전처리기

JIRA 데이터를 h4 섹션별로 분할하여 OpenAI 임베딩 요청 형식으로 전처리하는 도구입니다.

## 주요 기능

- JIRA description을 h4 섹션별로 자동 분할
- N/A, 빈 내용 등 의미없는 섹션 자동 제외
- OpenAI 임베딩 배치 API 요청 형식으로 변환
- JSONL 및 JSON 출력 형식 지원
- 모듈화된 구조로 재사용 가능

## 파일 구조

```
├── jira_processor.py    # 메인 처리 로직 (모듈화됨)
├── cli.py              # CLI 인터페이스
└── README.md           # 이 파일
```

## 설치 및 실행

### 필수 패키지

```bash
# Python 3.7+ 필요
# 기본 라이브러리만 사용 (추가 설치 불필요)
```

### CLI 사용법

```bash
# 기본 사용법
python cli.py input.json -o output.jsonl

# 샘플 데이터로 테스트
python cli.py --test

# 모든 옵션
python cli.py input.json \
    --output output.jsonl \
    --format jsonl \
    --include-empty
```

#### CLI 옵션

- `input_file`: 입력 JSON 파일 경로
- `-o, --output`: 출력 파일 경로 (기본값: embedding_requests.jsonl)
- `-f, --format`: 출력 형식 (jsonl 또는 json, 기본값: jsonl)
- `--test`: 샘플 데이터로 테스트 실행
- `--include-empty`: N/A나 빈 내용도 포함

## 모듈 사용법

### 1. 간단한 편의 함수 사용

```python
from jira_processor import process_jira_file, process_jira_data

# 파일에서 처리
success = process_jira_file('input.json', 'output.jsonl')

# 데이터 리스트로 처리
jira_data = [...]  # JIRA 데이터 리스트
success = process_jira_data(jira_data, 'output.jsonl')
```

### 2. 클래스 기반 사용

```python
from jira_processor import JiraEmbeddingProcessor

# 프로세서 생성 (빈 내용 제외)
processor = JiraEmbeddingProcessor(include_empty=False)

# 파일 처리
processor.process_from_file('input.json', 'output.jsonl', 'jsonl')

# 샘플 데이터 처리
processor.process_sample_data('sample_output.jsonl')

# 직접 데이터 처리
processor.process_data_list(jira_data_list, 'output.jsonl')
```

### 3. 개별 컴포넌트 사용

```python
from jira_processor import (
    JiraDataLoader,
    JiraDescriptionProcessor, 
    EmbeddingRequestSaver
)

# 데이터 로드
loader = JiraDataLoader()
jira_data = loader.load_from_file('input.json')

# 설명 처리
processor = JiraDescriptionProcessor(include_empty=False)
sections = processor.split_by_h4_sections(description)
requests = processor.create_embedding_requests("JIRA-123", sections)

# 결과 저장
saver = EmbeddingRequestSaver()
saver.save_requests(requests, 'output.jsonl', 'jsonl')
```

## 입력 데이터 형식

### JIRA JSON 형식

```json
[
  {
    "key": "GS-10691",
    "fields": {
      "description": "h4. Why will be improved\r\n개선 이유...\r\n\r\nh4. What will be improved\r\n개선 목표..."
    }
  }
]
```

또는 단일 객체:

```json
{
  "key": "GS-10691", 
  "fields": {
    "description": "..."
  }
}
```

## 출력 데이터 형식

### JSONL 형식 (기본값)

```jsonl
{"custom_id": "GS-10691-content-1", "method": "POST", "url": "/v1/embeddings", "body": {"model": "text-embedding-3-small", "input": "h4. Why will be improved\r\n개선 이유...", "encoding_format": "float"}}
{"custom_id": "GS-10691-content-2", "method": "POST", "url": "/v1/embeddings", "body": {"model": "text-embedding-3-small", "input": "h4. What will be improved\r\n개선 목표...", "encoding_format": "float"}}
```

### JSON 배열 형식

```json
[
  {
    "custom_id": "GS-10691-content-1",
    "method": "POST", 
    "url": "/v1/embeddings",
    "body": {
      "model": "text-embedding-3-small",
      "input": "h4. Why will be improved\r\n개선 이유...",
      "encoding_format": "float"
    }
  }
]
```

## 주요 클래스

### `JiraEmbeddingProcessor`
메인 처리 클래스로, 전체 워크플로우를 관리합니다.

### `JiraDescriptionProcessor`
JIRA description의 h4 섹션 분할과 의미있는 내용 필터링을 담당합니다.

### `JiraDataLoader`
JSON 파일에서 JIRA 데이터를 로드합니다.

### `EmbeddingRequestSaver`
임베딩 요청을 JSONL/JSON 형식으로 저장합니다.

## 섹션 필터링 규칙

다음과 같은 섹션은 자동으로 제외됩니다:

- 빈 내용이거나 공백만 있는 섹션
- N/A, 없음, TBD 등의 값만 있는 섹션  
- 2글자 이하의 매우 짧은 내용
- h4 헤더만 있고 실제 내용이 없는 섹션

`--include-empty` 옵션으로 이러한 섹션도 포함할 수 있습니다.

## 에러 처리

- 잘못된 JSON 형식: 에러 메시지 출력 후 계속 진행
- 누락된 필드: 경고 메시지 출력 후 해당 항목 스킵
- 파일 I/O 오류: 적절한 에러 메시지와 함께 종료

## 확장성

모듈화된 설계로 각 컴포넌트를 독립적으로 확장하거나 수정할 수 있습니다:

- 다른 임베딩 모델로 변경
- 다른 섹션 분할 규칙 적용
- 추가 전처리 로직 삽입
- 다른 출력 형식 지원