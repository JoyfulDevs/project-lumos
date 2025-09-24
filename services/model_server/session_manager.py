#!/usr/bin/env python3
"""
세션 관리 모듈 - 모델 타입별 고정 포트 할당
"""
import subprocess
import threading
import time
import signal
import atexit
import sys


class SessionManager:
    """AI 모델 세션 관리"""
    
    def __init__(self, config, ec2_manager, port_manager):
        self.config = config
        self.ec2_manager = ec2_manager
        self.port_manager = port_manager
        
        # 세션 관리
        self.active_sessions = {}  
        self.auto_shutdown = True
        
        # 종료 시그널 핸들러 등록 (메인 스레드에서만)
        try:
            signal.signal(signal.SIGINT, self._signal_handler)
            signal.signal(signal.SIGTERM, self._signal_handler)
            atexit.register(self._emergency_shutdown)
        except Exception as e:
            print(f"시그널 핸들러 등록 실패: {e}")
    
    def _signal_handler(self, signum, frame):
        """시그널 핸들러"""
        print(f"\n🛑 종료 시그널 감지 (시그널: {signum})")
        self._graceful_shutdown()
        sys.exit(0)
    
    def _emergency_shutdown(self):
        """응급 종료"""
        if self.auto_shutdown:
            self._cleanup_all_sessions()
    
    def _graceful_shutdown(self):
        """정상적인 종료 프로세스"""
        print("🧹 모든 세션 정리 중...")
        self._cleanup_all_sessions()
        
        # EC2Manager의 확인 기능 사용
        if self.auto_shutdown:
            # 활성 세션이 있었다면 자동으로 중지, 없었다면 사용자에게 물어보기
            auto_stop = len(self.active_sessions) > 0
            self.ec2_manager.stop_instance_with_confirmation(timeout=5, auto_stop=auto_stop)
        else:
            print("✅ EC2는 실행 상태로 유지됩니다.")
    
    def _cleanup_all_sessions(self):
        """모든 세션 정리"""
        for session_id in list(self.active_sessions.keys()):
            self._cleanup_session(session_id)
    
    def _cleanup_session(self, session_id):
        """특정 세션 정리"""
        if session_id in self.active_sessions:
            session = self.active_sessions[session_id]
            process = session.get('process')
            
            if process and process.poll() is None:
                print(f"세션 {session_id} 종료 중...")
                process.terminate()
                try:
                    process.wait(timeout=3)
                except subprocess.TimeoutExpired:
                    process.kill()
            
            del self.active_sessions[session_id]
    
    def _generate_unique_session_id(self, model_id, port):
        """고유한 세션 ID 생성"""
        timestamp = int(time.time() * 1000) % 100000  # 마지막 5자리
        return f"{model_id}_{port}_{timestamp}"
    
    def _check_model_type_conflict(self, model_info):
        """같은 타입의 모델이 이미 실행 중인지 확인"""
        is_embedding = model_info.get('embedding', True)
        model_type = 'embedding' if is_embedding else 'generation'
        
        for session_id, session in self.active_sessions.items():
            session_model_info = session.get('model_info', {})
            session_is_embedding = session_model_info.get('embedding', True)
            session_model_type = 'embedding' if session_is_embedding else 'generation'
            
            if session_model_type == model_type:
                return session_id, session
        
        return None, None
    
    def start_session(self, model_id=None, preferred_port=None):
        """새 세션 시작"""
        # EC2 시작
        if not self.ec2_manager.start_instance():
            return False
        
        # 모델 선택
        if not model_id:
            model_id = self._select_model()
            if not model_id:
                return False
        
        from config_manager import ConfigManager
        config_manager = ConfigManager()
        models = config_manager.get_available_models()
        
        if model_id not in models:
            print(f"❌ 모델을 찾을 수 없습니다: {model_id}")
            print("사용 가능한 모델:")
            for mid in models.keys():
                print(f"   - {mid}")
            return False
        
        model_info = models[model_id]
        is_embedding = model_info.get('embedding', True)
        model_type = 'embedding' if is_embedding else 'generation'
        
        # 포트 할당 정보 표시
        self.port_manager.show_port_assignment_info()
        
        # 같은 타입의 모델이 이미 실행 중인지 확인
        existing_session_id, existing_session = self._check_model_type_conflict(model_info)
        if existing_session_id:
            print(f"⚠️ 경고: {model_type} 모델이 이미 실행 중입니다!")
            print(f"   기존 세션: {existing_session_id}")
            print(f"   기존 모델: {existing_session.get('model_name', 'Unknown')}")
            print(f"   포트: {existing_session['port']}")
            print()
            
            choice = input("기존 세션을 중지하고 새 모델을 시작할까요? (y/n): ").strip().lower()
            if choice == 'y':
                print(f"🔄 기존 {model_type} 모델 세션 중지 중...")
                self._cleanup_session(existing_session_id)
                time.sleep(2)  # 포트 해제 대기
            else:
                print("❌ 세션 시작을 취소했습니다.")
                return False
        
        # 사용 가능한 포트 찾기 (모델 타입별 고정 포트)
        try:
            used_ports = {session['port'] for session in self.active_sessions.values()}
            port = self.port_manager.get_available_port(model_info, preferred_port, used_ports)
            print(f"✅ 할당된 포트: {port} ({model_type} 모델용)")
        except Exception as e:
            print(f"❌ 포트 할당 실패: {e}")
            return False
        
        # 고유한 세션 ID 생성
        session_id = self._generate_unique_session_id(model_id, port)
        print(f"🆔 생성된 세션 ID: {session_id}")
        
        # 세션 시작
        return self._run_model_server(session_id, model_id, port, model_info)
    
    def _select_model(self):
        """모델 선택 인터페이스"""
        from config_manager import ConfigManager
        config_manager = ConfigManager()
        models = config_manager.get_available_models()
        
        if not models:
            print("등록된 모델이 없습니다.")
            return None
        
        if len(models) == 1:
            model_id = list(models.keys())[0]
            print(f"자동 선택: {model_id}")
            return model_id
        
        print("\n사용 가능한 모델:")
        print("-" * 50)
        model_list = list(models.items())
        
        for i, (model_id, model_info) in enumerate(model_list, 1):
            model_type = "임베딩" if model_info.get('embedding', True) else "생성"
            assigned_port = self.port_manager.get_assigned_port_for_model(model_info)
            name = model_info.get('name', model_id)
            print(f"{i}. {model_id}")
            print(f"   이름: {name}")
            print(f"   타입: 🔹 {model_type} 모델 (포트 {assigned_port})")
            print()
        
        try:
            choice = int(input(f"선택 (1-{len(model_list)}): "))
            if 1 <= choice <= len(model_list):
                return model_list[choice - 1][0]
            else:
                print("❌ 잘못된 선택입니다.")
                return None
        except ValueError:
            print("❌ 숫자를 입력해주세요.")
            return None
    
    def _run_model_server(self, session_id, model_id, port, model_info):
        """모델 서버 실행"""
        import os
        
        state, public_ip = self.ec2_manager.get_instance_status()
        
        if state != 'running' or not public_ip:
            print("❌ EC2가 준비되지 않았습니다.")
            return False
        
        # SSH 키 파일 확인
        ssh_key = os.path.expanduser(self.config['ssh_key_path'])
        if not os.path.exists(ssh_key):
            print(f"❌ SSH 키 파일을 찾을 수 없습니다: {ssh_key}")
            print("💡 관리자에게 올바른 SSH 키를 요청하세요.")
            return False
        
        if not sys.platform.startswith("win"):
            os.chmod(ssh_key, 0o600)
        
        model_type = "임베딩" if model_info.get('embedding', True) else "생성"
        
        print(f"🚀 세션 시작: {session_id}")
        print(f"📦 모델: {model_info.get('name', model_id)} ({model_type})")
        print(f"🌐 주소: http://{public_ip}:{port}")
        print(f"🛑 개별 중지: python run_model_server.py stop-session {session_id}")
        print("-" * 50)
        
        # 서버 명령어 구성
        work_dir = self.config.get('server_work_dir', '/home/ubuntu/llama.cpp')
        
        server_command = f"""
cd {work_dir} && \\
./build/bin/llama-server \\
  -m {model_info['path']} \\
  --host 0.0.0.0 \\
  --port {port} \\
  --n-gpu-layers {model_info.get('gpu_layers', 32)} \\
  --threads {model_info.get('threads', 4)}"""
        
        if model_info.get('embedding', True):
            server_command += " --embedding"
        
        # 실행 전 마지막 포트 확인
        if self.port_manager.check_remote_port_in_use(port):
            print(f"⚠️ 경고: 포트 {port}가 이미 사용 중일 수 있습니다!")
            print("강제로 진행하려면 Enter, 취소하려면 Ctrl+C...")
            try:
                input()
            except KeyboardInterrupt:
                print("\n❌ 사용자가 취소했습니다.")
                return False
        
        ssh_cmd = [
            'ssh', '-i', ssh_key,
            '-o', 'StrictHostKeyChecking=no',
            '-o', 'UserKnownHostsFile=/dev/null',
            f"{self.config['ec2_user']}@{public_ip}",
            server_command
        ]
        
        try:
            print(f"🔧 SSH 명령 실행: {' '.join(ssh_cmd[:6])}...")
            process = subprocess.Popen(
                ssh_cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                text=True,
                bufsize=1,
                encoding='utf-8',  # UTF-8 인코딩 명시적 지정
                errors='replace'   # 디코딩 에러 시 치환 문자로 대체
            )
            
            # 세션 등록
            self.active_sessions[session_id] = {
                'process': process,
                'model_id': model_id,
                'model_name': model_info.get('name', model_id),
                'model_info': model_info,  # 모델 정보 추가
                'port': port,
                'public_ip': public_ip,
                'start_time': time.time(),
                'model_type': model_type
            }
            
            # 별도 스레드에서 로그 출력 (인코딩 에러 방지)
            def log_output():
                try:
                    for line in iter(process.stdout.readline, ''):
                        if line:
                            # 출력 가능한 문자만 유지, 문제 문자는 치환
                            safe_line = line.encode('utf-8', errors='replace').decode('utf-8')
                            print(f"[{session_id}] {safe_line.rstrip()}")
                        if process.poll() is not None:
                            break
                except UnicodeDecodeError as e:
                    print(f"[{session_id}] ⚠️ 인코딩 에러: {e}")
                except Exception as e:
                    print(f"[{session_id}] ❌ 로그 출력 에러: {e}")
            
            threading.Thread(target=log_output, daemon=True).start()
            
            print(f"✅ 세션 {session_id} 시작됨")
            return True
            
        except Exception as e:
            print(f"❌ 서버 실행 실패: {e}")
            # 실패한 세션 정리
            if session_id in self.active_sessions:
                del self.active_sessions[session_id]
            return False
    
    def stop_session(self, session_id):
        """특정 세션 중지"""
        if session_id not in self.active_sessions:
            print(f"❌ 세션을 찾을 수 없습니다: {session_id}")
            if self.active_sessions:
                print("실행 중인 세션:")
                for sid in self.active_sessions.keys():
                    print(f"  - {sid}")
            return False
        
        print(f"🛑 세션 중지 중: {session_id}")
        self._cleanup_session(session_id)
        print(f"✅ 세션 {session_id} 중지됨")
        return True
    
    def stop_all_sessions(self):
        """모든 세션 중지"""
        if not self.active_sessions:
            print("ℹ️ 실행 중인 세션이 없습니다.")
            return True
        
        print("🛑 모든 세션 중지 중...")
        self._cleanup_all_sessions()
        print("✅ 모든 세션 중지됨")
        return True
    
    def show_status(self):
        """상태 확인"""
        state, public_ip = self.ec2_manager.get_instance_status()
        
        print(f"\n🖥️ 다중 AI 빌드 서버 상태")
        print(f"    인스턴스: {self.config['instance_id']}")
        print(f"    상태: {self.ec2_manager.get_status_emoji(state)} {state}")
        
        # 포트 할당 정보 표시
        print(f"\n📋 모델 타입별 포트 할당:")
        for model_type, port in self.port_manager.port_assignments.items():
            print(f"    {model_type.capitalize()} 모델: 포트 {port}")
        
        if state == 'running' and public_ip:
            print(f"\n🌐 IP 주소: {public_ip}")
            
            # EC2에서 실제 사용 중인 포트들 확인
            self.port_manager.show_remote_ports(public_ip)
        
        print(f"\n🔥 활성 세션: {len(self.active_sessions)}")
        
        if self.active_sessions:
            print("-" * 70)
            for session_id, session in self.active_sessions.items():
                status = "🟢 실행중" if session['process'].poll() is None else "🔴 중지됨"
                runtime = int(time.time() - session.get('start_time', time.time()))
                model_type = session.get('model_type', 'Unknown')
                
                print(f"   {session_id}")
                print(f"     모델: {session['model_name']} ({model_type})")
                print(f"     주소: http://{session['public_ip']}:{session['port']}")
                print(f"     상태: {status}")
                print(f"     실행시간: {runtime//60}분 {runtime%60}초")
                
                # 실제 포트 사용 여부 확인
                if self.port_manager.check_remote_port_in_use(session['port']):
                    print(f"     포트: 🔴 {session['port']} (실제 사용중)")
                else:
                    print(f"     포트: 🟡 {session['port']} (비활성)")
                print()
        else:
            print("   (실행 중인 세션 없음)")
        
        print()
    
    def list_models(self):
        """사용 가능한 모델 목록 표시"""
        from config_manager import ConfigManager
        config_manager = ConfigManager()
        models = config_manager.get_available_models()
        
        print("\n📦 사용 가능한 모델 목록:")
        print("-" * 50)
        
        if not models:
            print("❌ 등록된 모델이 없습니다.")
            return
        
        # 포트 할당 정보도 함께 표시
        print("📋 모델 타입별 포트 할당:")
        for model_type, port in self.port_manager.port_assignments.items():
            print(f"   {model_type.capitalize()} 모델: 포트 {port}")
        print()
        
        for i, (model_id, model_info) in enumerate(models.items(), 1):
            model_type = "임베딩" if model_info.get('embedding', True) else "생성"
            assigned_port = self.port_manager.get_assigned_port_for_model(model_info)
            print(f"{i}. {model_id}")
            print(f"   이름: {model_info.get('name', model_id)}")
            print(f"   타입: 🔹 {model_type} 모델 (포트 {assigned_port})")
            print(f"   GPU 레이어: {model_info.get('gpu_layers', 32)}")
            print(f"   경로: {model_info.get('path', 'N/A')}")
            print()