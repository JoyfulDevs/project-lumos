#!/usr/bin/env python3
"""
포트 관리 및 충돌 해결 모듈 - 모델 타입별 고정 포트 할당
"""
import socket
import subprocess
import random
import os


class PortManager:
    """포트 관리 및 충돌 해결"""
    
    def __init__(self, config, ec2_manager):
        self.config = config
        self.ec2_manager = ec2_manager
        
        # 모델 타입별 고정 포트 설정
        self.port_assignments = {
            'embedding': 8080,  # 임베딩 모델은 8080 고정
            'generation': 8081  # 생성 모델은 8081 고정
        }
    
    def is_local_port_in_use(self, port):
        """로컬에서 포트가 사용 중인지 확인"""
        try:
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
                sock.settimeout(1)
                result = sock.connect_ex(('localhost', port))
                return result == 0  # 연결 성공하면 포트 사용 중
        except:
            return False
    
    def check_remote_port_in_use(self, port):
        """EC2 인스턴스에서 특정 포트가 사용 중인지 확인"""
        state, public_ip = self.ec2_manager.get_instance_status()
        
        if state != 'running' or not public_ip:
            return False
        
        ssh_key = os.path.expanduser(self.config['ssh_key_path'])
        
        # netstat으로 포트 사용 여부 확인
        check_cmd = [
            'ssh', '-i', ssh_key,
            '-o', 'StrictHostKeyChecking=no',
            '-o', 'UserKnownHostsFile=/dev/null',
            '-o', 'ConnectTimeout=5',
            f"{self.config['ec2_user']}@{public_ip}",
            f"netstat -tlnp | grep ':{port}' || echo 'PORT_FREE'"
        ]
        
        try:
            result = subprocess.run(
                check_cmd, 
                capture_output=True, 
                text=True, 
                timeout=10,
                encoding='utf-8',  # UTF-8 인코딩 명시적 지정
                errors='replace'   # 디코딩 에러 시 치환 문자로 대체
            )
            
            # PORT_FREE가 출력되면 포트가 비어있음
            return 'PORT_FREE' not in result.stdout
            
        except (subprocess.TimeoutExpired, subprocess.CalledProcessError):
            # 에러 발생 시 안전하게 사용 중으로 간주
            print(f"포트 {port} 원격 확인 실패, 안전하게 사용중으로 간주")
            return True
    
    def get_assigned_port_for_model(self, model_info):
        """모델 정보에 따라 할당된 포트 반환"""
        is_embedding = model_info.get('embedding', True)
        model_type = 'embedding' if is_embedding else 'generation'
        return self.port_assignments[model_type]
    
    def check_port_availability(self, port, model_type, used_ports=None):
        """특정 포트의 사용 가능성 확인"""
        used_ports = used_ports or set()
        
        print(f"🔍 {model_type} 모델용 포트 {port} 사용 가능성 검사 중...")
        
        # 1. 메모리상의 세션에서 사용 중인지 확인
        if port in used_ports:
            print(f"    포트 {port}: 메모리상 세션에서 사용중")
            return False, "메모리상 세션에서 사용중"
        
        # 2. 로컬에서 포트 사용 가능한지 확인
        if self.is_local_port_in_use(port):
            print(f"    포트 {port}: 로컬에서 사용중")
            return False, "로컬에서 사용중"
        
        # 3. EC2에서 포트 사용 중인지 확인
        if self.check_remote_port_in_use(port):
            print(f"    포트 {port}: EC2에서 사용중")
            return False, "EC2에서 사용중"
        
        print(f"    포트 {port}: 사용 가능!")
        return True, "사용 가능"
    
    def get_available_port(self, model_info, preferred_port=None, used_ports=None):
        """모델 타입에 따른 고정 포트 할당 (포트 충돌 시 에러 반환)"""
        used_ports = used_ports or set()
        
        # 모델 타입 확인
        is_embedding = model_info.get('embedding', True)
        model_type = 'embedding' if is_embedding else 'generation'
        
        # 할당된 포트 가져오기
        assigned_port = self.get_assigned_port_for_model(model_info)
        
        # 사용자가 선호 포트를 지정한 경우 경고
        if preferred_port and preferred_port != assigned_port:
            print(f"⚠️ 경고: {model_type} 모델은 포트 {assigned_port}로 고정되어 있습니다.")
            print(f"    요청된 포트 {preferred_port}는 무시되고 {assigned_port}를 사용합니다.")
        
        # 포트 사용 가능성 확인
        is_available, reason = self.check_port_availability(assigned_port, model_type, used_ports)
        
        if is_available:
            print(f"✅ {model_type} 모델용 포트 {assigned_port} 할당 완료")
            return assigned_port
        else:
            # 포트가 사용 중인 경우 에러 발생
            raise Exception(
                f"❌ {model_type} 모델용 포트 {assigned_port}이 이미 사용 중입니다 ({reason})\n"
                f"해결 방법:\n"
                f"  1. 기존 {model_type} 모델 세션을 중지: python run_model_server.py stop-session <session_id>\n"
                f"  2. 포트 강제 정리: python run_model_server.py kill-ports {assigned_port}\n"
                f"  3. 포트 상태 확인: python run_model_server.py debug-ports"
            )
    
    def show_remote_ports(self, public_ip):
        """EC2에서 사용 중인 포트들 표시"""
        ssh_key = os.path.expanduser(self.config['ssh_key_path'])
        
        ports_cmd = [
            'ssh', '-i', ssh_key,
            '-o', 'StrictHostKeyChecking=no',
            '-o', 'UserKnownHostsFile=/dev/null',
            '-o', 'ConnectTimeout=5',
            f"{self.config['ec2_user']}@{public_ip}",
            "netstat -tlnp | grep ':80[0-9][0-9]' | awk '{print $4}' | cut -d: -f2 | sort -n"
        ]
        
        try:
            result = subprocess.run(
                ports_cmd, 
                capture_output=True, 
                text=True, 
                timeout=10,
                encoding='utf-8',  # UTF-8 인코딩 명시적 지정
                errors='replace'   # 디코딩 에러 시 치환 문자로 대체
            )
            
            if result.stdout.strip():
                used_ports = result.stdout.strip().split('\n')
                print(f"   🔌 EC2 사용중 포트: {', '.join(used_ports)}")
            else:
                print(f"   🔌 EC2 사용중 포트: 없음")
                
        except:
            print(f"   🔌 EC2 포트 확인 실패")
    
    def debug_ports(self, active_sessions):
        """포트 상태 디버깅"""
        print("\n🔍 포트 충돌 디버깅 - 고정 포트 할당 방식")
        print("-" * 50)
        
        # 모델 타입별 고정 포트 정보 표시
        print("📋 모델 타입별 고정 포트 할당:")
        for model_type, port in self.port_assignments.items():
            print(f"   {model_type} 모델: 포트 {port}")
        print()
        
        # 1. 할당된 포트들의 상태 확인
        print("🔍 할당된 포트 상태:")
        for model_type, port in self.port_assignments.items():
            try:
                with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
                    sock.settimeout(1)
                    result = sock.connect_ex(('localhost', port))
                    status = "🔴 사용중" if result == 0 else "🟢 사용가능"
                    print(f"   포트 {port} ({model_type}): {status}")
            except:
                print(f"   포트 {port} ({model_type}): ❓ 확인불가")
        
        # 2. 메모리상 활성 세션
        print(f"\n🔍 메모리상 활성 세션: {len(active_sessions)}")
        for session_id, session in active_sessions.items():
            model_type = 'embedding' if session.get('model_info', {}).get('embedding', True) else 'generation'
            print(f"   {session_id}: 포트 {session['port']} ({model_type})")
        
        # 3. EC2 원격 포트 확인
        state, public_ip = self.ec2_manager.get_instance_status()
        if state == 'running' and public_ip:
            print(f"\n🔌 EC2 원격 포트 상태 ({public_ip}):")
            self.show_remote_ports(public_ip)
        else:
            print(f"\n🔌 EC2 상태: {state} (포트 확인 불가)")
    
    def kill_remote_ports(self, ports_to_kill):
        """EC2에서 특정 포트 범위의 프로세스 강제 종료"""
        state, public_ip = self.ec2_manager.get_instance_status()
        
        if state != 'running' or not public_ip:
            print("❌ EC2가 실행 중이 아닙니다.")
            return False
        
        ssh_key = os.path.expanduser(self.config['ssh_key_path'])
        
        for port in ports_to_kill:
            print(f"🔫 포트 {port}에서 실행 중인 프로세스 종료 중...")
            
            kill_cmd = [
                'ssh', '-i', ssh_key,
                '-o', 'StrictHostKeyChecking=no',
                '-o', 'UserKnownHostsFile=/dev/null',
                f"{self.config['ec2_user']}@{public_ip}",
                f"pkill -f 'port {port}' || lsof -ti:{port} | xargs -r kill -9"
            ]
            
            try:
                result = subprocess.run(
                    kill_cmd, 
                    capture_output=True, 
                    text=True, 
                    timeout=10,
                    encoding='utf-8',  # UTF-8 인코딩 명시적 지정
                    errors='replace'   # 디코딩 에러 시 치환 문자로 대체
                )
                if result.returncode == 0:
                    print(f"    포트 {port} 정리 완료")
                else:
                    print(f"   포트 {port}: {result.stderr.strip() or '프로세스 없음'}")
            except:
                print(f"   포트 {port} 정리 실패")
        
        print("포트 정리 작업 완료")
        return True
    
    def show_port_assignment_info(self):
        """포트 할당 정보 표시"""
        print("\n📋 모델 타입별 포트 할당:")
        print("-" * 30)
        for model_type, port in self.port_assignments.items():
            print(f"   {model_type.capitalize()} 모델: 포트 {port}")
        print()