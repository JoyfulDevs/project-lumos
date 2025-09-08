#!/usr/bin/env python3
"""
다중 AI 빌드 서버 메인 실행 파일
"""
import sys
import time
from multi_build_server import MultiBuildServer
from config_manager import ConfigManager


def kill_remote_ports():
    """EC2에서 특정 포트 범위의 프로세스 강제 종료"""
    if len(sys.argv) < 3:
        print("사용법: python run_model_server.py kill-ports <포트1> [포트2] [포트3]...")
        print("예시: python run_model_server.py kill-ports 8080 8081 8082")
        return
    
    ports_to_kill = sys.argv[2:]
    
    try:
        controller = MultiBuildServer()
        success = controller.kill_remote_ports(ports_to_kill)
        sys.exit(0 if success else 1)
        
    except Exception as e:
        print(f" 포트 정리 실패: {e}")
        sys.exit(1)


def main():
    """메인 함수"""
    if len(sys.argv) < 2:
        print("""
다중 AI 빌드 서버 제어 도구 (포트 충돌 해결)

기본 명령어:
  python run_model_server.py start [model_id] [--port N]  - 새 세션 시작
  python run_model_server.py stop-session <id>           - 특정 세션 중지
  python run_model_server.py stop-all                    - 모든 세션 중지
  python run_model_server.py status                      - 전체 상태 확인
  python run_model_server.py models                      - 사용 가능한 모델 목록

디버깅 명령어:
  python run_model_server.py debug-ports                 - 포트 상태 디버깅
  python run_model_server.py kill-ports 8080 8081       - 특정 포트 강제 정리

설정 관리:
  python run_model_server.py add-model                   - 기존 config에 새 모델 추가
  python run_model_server.py template                    - 다중 모델 템플릿 생성

사용 예시:
  python run_model_server.py start                       # 모델 선택 후 시작
  python run_model_server.py start qwen3-embedding       # 특정 모델로 시작
  python run_model_server.py start gpt-oss-20b --port 8085  # 특정 포트로 시작
  python run_model_server.py debug-ports                 # 포트 충돌 디버깅

포트 충돌 해결 기능:
  - 실제 포트 사용 상태 확인 (로컬 + EC2)
  - 자동 포트 할당 (8080부터 순차 검색)
  - 포트 사용 불가 시 랜덤 포트 할당
  - 고유한 세션 ID 생성 (타임스탬프 포함)
  
개선된 기능:
  - 포트 중복 사용 방지
  - 실시간 포트 상태 모니터링
  - 강제 포트 정리 도구
  - 자세한 디버깅 정보
        """)
        sys.exit(1)
    
    command = sys.argv[1].lower()
    
    # 템플릿 생성
    if command == 'template':
        ConfigManager.create_template()
        return
    
    # 모델 추가
    if command == 'add-model':
        ConfigManager.add_model_interactive()
        return
    
    # 포트 강제 정리
    if command == 'kill-ports':
        kill_remote_ports()
        return
    
    # 컨트롤러 초기화
    try:
        controller = MultiBuildServer()
    except Exception as e:
        print(f" 초기화 실패: {e}")
        sys.exit(1)
    
    # 포트 디버깅
    if command == 'debug-ports':
        controller.debug_ports()
        return
    
    # 명령어 실행
    if command == 'start':
        model_id = None
        preferred_port = None
        
        # 인자 파싱
        i = 2
        while i < len(sys.argv):
            if sys.argv[i] == '--port' and i + 1 < len(sys.argv):
                try:
                    preferred_port = int(sys.argv[i + 1])
                    i += 2
                except ValueError:
                    print(" 잘못된 포트 번호입니다.")
                    sys.exit(1)
            elif not model_id:
                model_id = sys.argv[i]
                i += 1
            else:
                i += 1
        
        success = controller.start_session(model_id, preferred_port)
        if success:
            print("\n 세션이 시작되었습니다!")
            print("   - 다른 모델로 추가 세션을 시작하려면: python run_model_server.py start [model_id]")
            print("   - 상태 확인: python run_model_server.py status")
            print("   - 세션 중지: python run_model_server.py stop-session <session_id>")
            print("   - 포트 디버깅: python run_model_server.py debug-ports")
            
            # 메인 스레드를 유지하여 로그 출력 계속
            try:
                while True:
                    time.sleep(1)
            except KeyboardInterrupt:
                controller.session_manager._graceful_shutdown()
        sys.exit(0 if success else 1)
    
    elif command == 'stop-session':
        if len(sys.argv) < 3:
            print(" 세션 ID를 입력해주세요: python run_model_server.py stop-session <session_id>")
            controller.show_status()  # 현재 세션들 표시
            sys.exit(1)
        session_id = sys.argv[2]
        success = controller.stop_session(session_id)
        sys.exit(0 if success else 1)
    
    elif command == 'stop-all':
        success = controller.stop_all_sessions()
        sys.exit(0 if success else 1)
    
    elif command == 'status':
        controller.show_status()
    
    elif command == 'models':
        controller.list_models()
    
    else:
        print(f" 알 수 없는 명령어: {command}")
        print("python run_model_server.py 를 실행하여 도움말을 확인하세요.")
        sys.exit(1)


if __name__ == "__main__":
    main()