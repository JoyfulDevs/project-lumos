# 다중 AI 빌드 서버 제어 도구

AWS EC2에서 여러 AI 모델을 실행할 수 있는 관리 도구입니다.
**모델 타입별 고정 포트 할당 방식**을 적용하여 포트 충돌을 방지합니다.

---

## 주요 특징

* **다중 모델 지원**: 임베딩 모델 / 생성 모델을 동시에 실행 가능
* **고정 포트 할당**: 모델 타입별 고정 포트 사용

  * 임베딩 모델: **8080**
  * 생성 모델: **8081**
* **세션 관리**: 모델 타입별 1개 세션만 허용 (중복 시 기존 세션 중지 여부 선택)
* **EC2 인스턴스 관리**: 시작/중지/상태 확인 및 자동 정리
* **포트 상태 모니터링**: 로컬/원격 포트 상태 실시간 확인
* **Windows/Linux/Mac 지원**

---

## 프로젝트 구조

```
ai-build-server/
├── run_model_server.py      # 메인 실행 파일
├── multi_build_server.py    # 통합 서버 클래스
├── config_manager.py        # 설정 관리
├── ec2_manager.py           # EC2 인스턴스 관리
├── port_manager.py          # 포트 관리 (고정 할당)
└── session_manager.py       # 모델 세션 관리
```

---

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

`config.json`에서 다음 항목들을 설정하세요:

```json
{
  "aws_access_key": "YOUR_AWS_ACCESS_KEY",
  "aws_secret_key": "YOUR_AWS_SECRET_KEY",
  "aws_region": "us-east-1",
  "instance_id": "i-1234567890abcdef0",
  "ssh_key_path": "./keys/server-key.pem",
  "ec2_user": "ubuntu",
  "server_work_dir": "/home/ubuntu/llama.cpp",
  "models": {
    "qwen3-embedding": {
      "name": "Qwen3 Embedding 0.6B",
      "path": "/home/ubuntu/llama.cpp/models/qwen3-embedding-0.6b/Qwen3-Embedding-0.6B-Q8_0.gguf",
      "gpu_layers": 32,
      "threads": 4,
      "embedding": true
    },
    "gpt-oss-20b": {
      "name": "GPT OSS 20B",
      "path": "/home/ubuntu/llama.cpp/models/gpt-oss-20b/gpt-oss-20b-f16.gguf",
      "gpu_layers": 40,
      "threads": 4,
      "embedding": false
    }
  }
}
```

---

## 사용법

### 기본 명령어

| 명령어                                            | 설명                 |
| ---------------------------------------------- | ------------------ |
| `python run_model_server.py start [model_id]`  | 새 세션 시작 (포트 자동 할당) |
| `python run_model_server.py stop-session <id>` | 특정 세션 중지           |
| `python run_model_server.py stop-all`          | 모든 세션 중지           |
| `python run_model_server.py status`            | 서버 및 세션 상태 확인      |
| `python run_model_server.py models`            | 사용 가능한 모델 목록 표시    |

### 디버깅 명령어

| 명령어                                               | 설명          |
| ------------------------------------------------- | ----------- |
| `python run_model_server.py debug-ports`          | 포트 상태 디버깅   |
| `python run_model_server.py kill-ports 8080 8081` | 특정 포트 강제 정리 |

### 설정 관리

| 명령어                                    | 설명        |
| -------------------------------------- | --------- |
| `python run_model_server.py add-model` | 새 모델 추가   |
| `python run_model_server.py template`  | 설정 템플릿 생성 |

---

## 세션 관리 규칙

* 같은 타입(임베딩/생성)의 모델은 동시에 **1개 세션만 실행 가능**
* 같은 타입을 새로 실행할 경우 → 기존 세션 중지 여부를 선택
* 세션 ID는 `모델명_포트_타임스탬프` 형식으로 생성

---

## 포트 관리

* **고정 포트 사용**:

  * 임베딩 모델: **8080**
  * 생성 모델: **8081**
* `--port` 옵션은 무시됨 (강제 고정)
* 충돌 시 에러 메시지 및 해결 방법 안내 (세션 중지, 포트 강제 정리)

---

## 문제 해결

### 포트 충돌

```bash
python run_model_server.py debug-ports
python run_model_server.py kill-ports 8080
```

### SSH 문제

```bash
ls -la keys/yerin-linux.pem
chmod 600 keys/yerin-linux.pem
```

### 세션 정리

```bash
python run_model_server.py stop-all
```

---

## 라이선스

MIT License
