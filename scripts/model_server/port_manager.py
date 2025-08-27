#!/usr/bin/env python3
"""
포트 관리 및 충돌 해결 모듈
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
    
    def get_available_port(self, preferred_port=None, used_ports=None):
        """실제로 사용 가능한 포트 찾기 (로컬 + 원격 확인)"""
        base_port = preferred_port or self.config.get('base_port', 8080)
        used_ports = used_ports or set()
        
        max_attempts = 100  # 무한루프 방지
        attempts = 0
        
        # 1. 선호 포트부터 순차적으로 시도
        port = base_port
        while attempts < max_attempts:
            print(f"🔍 포트 {port} 사용 가능성 검사 중...")
            
            # 1. 메모리상의 세션에서 사용 중인지 확인
            if port in used_ports:
                print(f"    포트 {port}: 메모리상 세션에서 사용중")
                port += 1
                attempts += 1
                continue
            
            # 2. 로컬에서 포트 사용 가능한지 확인
            if self.is_local_port_in_use(port):
                print(f"    포트 {port}: 로컬에서 사용중")
                port += 1
                attempts += 1
                continue
            
            # 3. EC2에서 포트 사용 중인지 확인
            if self.check_remote_port_in_use(port):
                print(f"    포트 {port}: EC2에서 사용중")
                port += 1
                attempts += 1
                continue
            
            # 모든 검사를 통과한 포트 반환
            print(f"   포트 {port}: 사용 가능!")
            return port
        
        # 순차 검색 실패 시 랜덤 포트 시도
        print("⚠️ 순차 포트 검색 실패, 랜덤 포트 시도...")
        for _ in range(20):
            random_port = random.randint(8100, 8999)
            if (random_port not in used_ports and 
                not self.is_local_port_in_use(random_port) and 
                not self.check_remote_port_in_use(random_port)):
                print(f" 랜덤 포트 {random_port} 사용 가능!")
                return random_port
        
        # 모든 시도 실패
        raise Exception(f"사용 가능한 포트를 찾을 수 없습니다 (시도한 범위: {base_port}~{port}, 랜덤 포트도 실패)")
    
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
        print("\n🔍 포트 충돌 디버깅")
        print("-" * 50)
        
        # 1. 로컬 포트 8080-8090 확인
        print("🔍 로컬 포트 상태:")
        for port in range(8080, 8091):
            try:
                with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
                    sock.settimeout(1)
                    result = sock.connect_ex(('localhost', port))
                    status = "🔴 사용중" if result == 0 else "🟢 사용가능"
                    print(f"   포트 {port}: {status}")
            except:
                print(f"   포트 {port}: ❓ 확인불가")
        
        # 2. 메모리상 활성 세션
        print(f"\n🔍 메모리상 활성 세션: {len(active_sessions)}")
        for session_id, session in active_sessions.items():
            print(f"   {session_id}: 포트 {session['port']}")
        
        # 3. EC2 원격 포트 확인
        state, public_ip = self.ec2_manager.get_instance_status()
        if state == 'running' and public_ip:
            print(f"\n EC2 원격 포트 상태 ({public_ip}):")
            self.show_remote_ports(public_ip)
        else:
            print(f"\n EC2 상태: {state} (포트 확인 불가)")
    
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