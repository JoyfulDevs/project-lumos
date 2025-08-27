#!/usr/bin/env python3
"""
통합 AI 빌드 서버 클래스
"""
from config_manager import ConfigManager
from ec2_manager import EC2Manager
from port_manager import PortManager
from session_manager import SessionManager


class MultiBuildServer:
    """통합 AI 빌드 서버"""
    
    def __init__(self, config_file='config.json'):
        # 설정 관리자 초기화
        self.config_manager = ConfigManager(config_file)
        config = self.config_manager.get_config()
        
        # 각 관리자 초기화
        self.ec2_manager = EC2Manager(config)
        self.port_manager = PortManager(config, self.ec2_manager)
        self.session_manager = SessionManager(config, self.ec2_manager, self.port_manager)
        
        print(f" 다중 AI 빌드 서버 준비 완료 (리전: {config['aws_region']})")
        print("자동 종료 기능 활성화됨")
        
        # 모델 정보 표시
        models = self.config_manager.get_available_models()
        print(f"사용 가능한 모델: {len(models)}")
    
    # SessionManager의 메서드들을 위임
    def start_session(self, model_id=None, preferred_port=None):
        """새 세션 시작"""
        return self.session_manager.start_session(model_id, preferred_port)
    
    def stop_session(self, session_id):
        """특정 세션 중지"""
        return self.session_manager.stop_session(session_id)
    
    def stop_all_sessions(self):
        """모든 세션 중지"""
        return self.session_manager.stop_all_sessions()
    
    def show_status(self):
        """상태 확인"""
        return self.session_manager.show_status()
    
    def list_models(self):
        """사용 가능한 모델 목록 표시"""
        return self.session_manager.list_models()
    
    # PortManager의 메서드들을 위임
    def debug_ports(self):
        """포트 상태 디버깅"""
        return self.port_manager.debug_ports(self.session_manager.active_sessions)
    
    def kill_remote_ports(self, ports_to_kill):
        """원격 포트 강제 종료"""
        return self.port_manager.kill_remote_ports(ports_to_kill)
    
    # ConfigManager의 메서드들을 위임
    def add_model_interactive(self):
        """대화형 모델 추가"""
        return self.config_manager.add_model_interactive()
    
    def create_template(self):
        """설정 템플릿 생성"""
        return self.config_manager.create_template()
    
    # EC2Manager의 메서드들을 위임  
    def get_instance_status(self):
        """EC2 인스턴스 상태 확인"""
        return self.ec2_manager.get_instance_status()
    
    def start_ec2(self):
        """EC2 인스턴스 시작"""
        return self.ec2_manager.start_instance()
    
    def stop_ec2(self):
        """EC2 인스턴스 중지"""
        return self.ec2_manager.stop_instance()