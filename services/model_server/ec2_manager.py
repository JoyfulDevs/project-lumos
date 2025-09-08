#!/usr/bin/env python3
"""
EC2 인스턴스 관리 모듈
"""
import boto3
import time
from botocore.exceptions import ClientError


class EC2Manager:
    """EC2 인스턴스 관리"""
    
    def __init__(self, config):
        self.config = config
        
        # AWS 클라이언트 초기화
        self.ec2_client = boto3.client(
            'ec2',
            aws_access_key_id=config['aws_access_key'],
            aws_secret_access_key=config['aws_secret_key'],
            region_name=config['aws_region']
        )
    
    def get_instance_status(self):
        """인스턴스 상태 확인"""
        try:
            response = self.ec2_client.describe_instances(
                InstanceIds=[self.config['instance_id']]
            )
            instance = response['Reservations'][0]['Instances'][0]
            return instance['State']['Name'], instance.get('PublicIpAddress')
        except ClientError as e:
            print(f" 인스턴스 상태 확인 실패: {e}")
            return None, None
    
    def start_instance(self):
        """EC2 인스턴스 시작"""
        state, _ = self.get_instance_status()
        
        if state == 'running':
            print("✅ EC2가 이미 실행 중입니다!")
            return True
        elif state == 'pending':
            print("⏳ EC2가 시작 중입니다...")
        else:
            print("🔌 EC2 인스턴스 시작 중...")
            try:
                self.ec2_client.start_instances(
                    InstanceIds=[self.config['instance_id']]
                )
            except ClientError as e:
                print(f" EC2 시작 실패: {e}")
                return False
        
        return self._wait_for_ready()
    
    def _wait_for_ready(self):
        """EC2 준비 대기"""
        print("⏳ EC2 부팅 및 SSH 준비 대기 중...")
        
        for i in range(30):
            state, public_ip = self.get_instance_status()
            
            if state == 'running' and public_ip:
                print(f"🟢 EC2 실행 완료! (IP: {public_ip})")
                time.sleep(30)  # SSH 서비스 준비 시간
                return True
            elif state in ['terminated', 'terminating']:
                print(f" EC2가 종료 상태입니다: {state}")
                return False
            
            time.sleep(10)
        
        print("⏰ EC2 시작 시간 초과")
        return False
    
    def stop_instance(self):
        """EC2만 중지"""
        try:
            state, _ = self.get_instance_status()
            if state == 'running':
                self.ec2_client.stop_instances(
                    InstanceIds=[self.config['instance_id']]
                )
                print("✅ EC2 중지 명령 전송 완료")
            else:
                print(f"ℹ️ EC2는 이미 {state} 상태입니다.")
        except Exception as e:
            print(f"⚠️ EC2 중지 중 오류: {e}")
    
    def stop_instance_with_confirmation(self, timeout=5, auto_stop=False):
        """사용자 확인 후 EC2 중지"""
        if auto_stop:
            print("자동으로 EC2를 중지합니다...")
            self.stop_instance()
            return
            
        print("EC2 인스턴스를 중지할까요?")
        print("   y: EC2 중지")
        print("   n: EC2 실행 상태로 유지")
        print(f"   ({timeout}초 후 자동으로 EC2 중지)")
        
        try:
            import sys
            
            if sys.platform.startswith('win'):
                # Windows
                import msvcrt
                import time
                
                start_time = time.time()
                choice = None
                
                while time.time() - start_time < timeout:
                    if msvcrt.kbhit():
                        choice = msvcrt.getch().decode('utf-8').lower()
                        break
                    time.sleep(0.1)
                
                if choice == 'n':
                    print("✅ EC2는 실행 상태로 유지됩니다.")
                    return
            else:
                # Linux/Mac
                import select
                ready, _, _ = select.select([sys.stdin], [], [], timeout)
                if ready:
                    choice = sys.stdin.readline().strip().lower()
                    if choice == 'n':
                        print("✅ EC2는 실행 상태로 유지됩니다.")
                        return
            
            print(f"\n시간 초과 - EC2 자동 중지 중...")
            self.stop_instance()
            
        except Exception as e:
            print(f"\n입력 처리 오류 ({e}) - 안전을 위해 EC2를 중지합니다...")
            self.stop_instance()
    
    def get_status_emoji(self, state):
        """상태 이모지"""
        return {
            'running': '🟢', 
            'stopped': '🔴', 
            'pending': '🟡', 
            'stopping': '🟠'
        }.get(state, '⚪')