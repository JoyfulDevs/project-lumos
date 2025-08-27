# 다중 AI 빌드 서버 제어 도구

AWS EC2에서 여러 AI 모델을 동시에 실행할 수 있는 포트 충돌 해결 기능이 포함된 서버 관리 도구입니다.

## 주요 특징

- **다중 모델 지원**: 여러 AI 모델을 동시에 실행
- **포트 충돌 자동 해결**: 로컬 및 원격 포트 상태 확인 후 자동 할당
- **EC2 인스턴스 자동 관리**: 시작/중지/상태 모니터링
- **세션 관리**: 개별 세션 제어 및 자동 정리
- **Windows/Linux/Mac 지원**: 플랫폼별 최적화

## 프로젝트 구조

```
ai-build-server/
├── run_model_server.py      # 메인 실행 파일
├── multi_build_server.py    # 통합 서버 클래스
├── config_manager.py        # 설정 관리
├── ec2_manager.py          # EC2 인스턴스 관리
├── port_manager.py         # 포트 관리 및 충돌 해결
└──  session_manager.py      # AI 모델 세션 관리           
```

## 설치 및 설정

### 1. 의존성 설치

```bash
pip install boto3
```

### 2. 설정 파일 생성

```bash
# 템플릿 생성
python run_model_server.py template

# config.json 편집
vim config.json
```

### 3. SSH 키 권한 설정

```bash
chmod 600 keys/server-key.pem
```

### 4. AWS 자격 증명 설정

config.json에서 다음 항목들을 설정하세요:

```json
{
  "aws_access_key": "YOUR_AWS_ACCESS_KEY",
  "aws_secret_key": "YOUR_AWS_SECRET_KEY",
  "aws_region": "us-east-1",
  "instance_id": "i-1234567890abcdef0",
  "ssh_key_path": "./keys/server-key.pem",
  "ec2_user": "ubuntu",
  "base_port": 8080,
  "server_work_dir": "/home/ubuntu/llama.cpp"
}
```

## 사용법

### 기본 명령어

| 명령어 | 설명 |
|--------|------|
| `python run_model_server.py start [model_id] [--port N]` | 새 세션 시작 |
| `python run_model_server.py stop-session <id>` | 특정 세션 중지 |
| `python run_model_server.py stop-all` | 모든 세션 중지 |
| `python run_model_server.py status` | 전체 상태 확인 |
| `python run_model_server.py models` | 사용 가능한 모델 목록 |

### 디버깅 명령어

| 명령어 | 설명 |
|--------|------|
| `python run_model_server.py debug-ports` | 포트 상태 디버깅 |
| `python run_model_server.py kill-ports 8080 8081` | 특정 포트 강제 정리 |

### 설정 관리

| 명령어 | 설명 |
|--------|------|
| `python run_model_server.py add-model` | 기존 config에 새 모델 추가 |
| `python run_model_server.py template` | 다중 모델 템플릿 생성 |

## 사용 예시

### 기본 사용

```bash
# 모델 선택 후 시작
python run_model_server.py start

# 특정 모델로 시작
python run_model_server.py start qwen3-embedding

# 특정 포트로 시작
python run_model_server.py start gpt-oss-20b --port 8085

# 상태 확인
python run_model_server.py status
```

### 디버깅

```bash
# 포트 충돌 디버깅
python run_model_server.py debug-ports

# 문제가 있는 포트 강제 정리
python run_model_server.py kill-ports 8080 8081 8082
```

### 모델 관리

```bash
# 사용 가능한 모델 확인
python run_model_server.py models

# 새 모델 추가
python run_model_server.py add-model
```

## 포트 충돌 해결 기능

### 자동 포트 할당

- **순차 검색**: 8080부터 시작하여 사용 가능한 포트 자동 탐지
- **실시간 확인**: 로컬 및 EC2 원격 서버의 포트 사용 상태 동시 확인
- **랜덤 할당**: 순차 검색 실패 시 8100-8999 범위에서 랜덤 포트 할당
- **중복 방지**: 메모리상 활성 세션과의 포트 충돌 방지

### 포트 상태 모니터링

```bash
# 상세한 포트 상태 확인
python run_model_server.py debug-ports
```

출력 예시:
```
🔍 포트 충돌 디버깅
--------------------------------------------------
🔍 로컬 포트 상태:
   포트 8080: 🟢 사용가능
   포트 8081: 🔴 사용중
   포트 8082: 🟢 사용가능

🔍 메모리상 활성 세션: 2
   qwen3-embedding_8080_12345: 포트 8080
   gpt-oss-20b_8082_67890: 포트 8082

🌐 EC2 원격 포트 상태 (3.25.123.45):
   🔌 EC2 사용중 포트: 8080, 8082, 22
```

## 개선된 기능

### 세션 관리

- **고유 세션 ID**: 모델명_포트_타임스탬프 형식으로 중복 방지
- **다중 세션**: 여러 모델을 동시에 실행 가능
- **자동 정리**: 프로그램 종료 시 모든 리소스 안전 정리
- **실시간 로그**: 각 세션별 독립적인 로그 출력

### EC2 관리

- **자동 시작**: 세션 시작 시 EC2 자동 부팅
- **SSH 준비 대기**: 서비스 준비 완료까지 자동 대기
- **안전한 종료**: Ctrl+C 시 사용자 확인 후 EC2 중지 선택

### 에러 처리

- **인코딩 처리**: Windows/Linux 환경에서 UTF-8 인코딩 자동 처리
- **네트워크 타임아웃**: SSH 연결 및 명령 실행 시 타임아웃 설정
- **예외 복구**: 각종 예외 상황에서 안전한 복구 메커니즘

## 종료 및 정리

### 정상 종료

```bash
# 실행 중인 서버에서 Ctrl+C
# → 세션 정리 → EC2 중지 확인 → 종료
```

### 개별 세션 중지

```bash
# 특정 세션만 중지
python run_model_server.py stop-session qwen3-embedding_8080_12345

# 모든 세션 중지 (EC2는 유지)
python run_model_server.py stop-all
```

### 강제 정리

```bash
# 문제 상황 시 원격 포트 강제 정리
python run_model_server.py kill-ports 8080 8081
```

## 문제 해결

### 포트 충돌 문제

1. **현재 포트 상태 확인**:
   ```bash
   python run_model_server.py debug-ports
   ```

2. **문제 포트 강제 정리**:
   ```bash
   python run_model_server.py kill-ports 8080
   ```

3. **다른 포트로 시도**:
   ```bash
   python run_model_server.py start --port 8085
   ```

### SSH 연결 문제

1. **키 파일 권한 확인**:
   ```bash
   ls -la keys/server-key.pem
   chmod 600 keys/server-key.pem
   ```

2. **EC2 상태 확인**:
   ```bash
   python run_model_server.py status
   ```

### 인코딩 에러 (Windows)

- UTF-8 인코딩이 자동으로 처리되도록 업데이트되었습니다
- 여전히 문제가 있다면 Windows Terminal 사용을 권장합니다

## 라이선스

MIT License

## 기여

이슈나 개선 제안은 GitHub Issues를 통해 제출해주세요.