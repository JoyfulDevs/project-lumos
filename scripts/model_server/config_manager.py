#!/usr/bin/env python3
"""
설정 관리 모듈
"""
import os
import json
import sys


class ConfigManager:
    """설정 파일 로드 및 관리"""
    
    def __init__(self, config_file='config.json'):
        self.config_file = config_file
        self.config = self._load_config()
    
    def _load_config(self):
        """설정 파일 로드"""
        if not os.path.exists(self.config_file):
            print(f" {self.config_file} 파일을 찾을 수 없습니다.")
            print("💡 관리자에게 설정 파일을 요청하세요.")
            sys.exit(1)
        
        try:
            with open(self.config_file, 'r', encoding='utf-8') as f:
                config = json.load(f)
            
            # 필수 항목 체크
            required_keys = ['aws_access_key', 'aws_secret_key', 'instance_id', 'ssh_key_path']
            for key in required_keys:
                if key not in config or not config[key]:
                    print(f" 설정 파일에 {key}가 누락되었습니다.")
                    sys.exit(1)
            
            # 기존 설정 형식 호환성 유지
            config.setdefault('aws_region', 'us-west-2')
            config.setdefault('ec2_user', 'ubuntu')
            
            # 기존 단일 모델 설정을 다중 모델 형식으로 변환
            if 'models' not in config:
                # 기존 설정에서 기본 모델 생성
                default_model = {
                    'name': 'Default Model',
                    'path': config.get('model_path', ''),
                    'gpu_layers': config.get('gpu_layers', 32),
                    'threads': config.get('threads', 4),
                    'embedding': True  
                }
                
                config['models'] = {
                    'default': default_model
                }

            if 'base_port' not in config:
                config['base_port'] = config.get('server_port', 8080)
            
            return config
            
        except json.JSONDecodeError as e:
            print(f" 설정 파일 형식 오류: {e}")
            sys.exit(1)
        except Exception as e:
            print(f" 설정 파일 로드 실패: {e}")
            sys.exit(1)
    
    def get_config(self):
        """전체 설정 반환"""
        return self.config
    
    def get_available_models(self):
        """사용 가능한 모델 목록 반환"""
        return self.config.get('models', {})
    
    def add_model(self, model_id, model_info):
        """새 모델 추가"""
        if 'models' not in self.config:
            self.config['models'] = {}
        
        self.config['models'][model_id] = model_info
        self._save_config()
    
    def _save_config(self):
        """설정 파일 저장"""
        try:
            with open(self.config_file, 'w', encoding='utf-8') as f:
                json.dump(self.config, f, indent=2, ensure_ascii=False)
        except Exception as e:
            print(f" 설정 파일 저장 실패: {e}")
            raise
    
    @staticmethod
    def add_model_interactive():
        """대화형으로 새 모델 추가"""
        config_file = 'config.json'
        
        if not os.path.exists(config_file):
            print(" config.json 파일을 찾을 수 없습니다.")
            return
        
        try:
            config_manager = ConfigManager(config_file)
            
            print("\n 새 모델 추가")
            print("-" * 30)
            
            model_id = input("모델 ID (예: gpt-oss-20b): ").strip()
            if not model_id:
                print(" 모델 ID가 필요합니다.")
                return
            
            model_name = input("모델 이름: ").strip() or model_id
            model_path = input("모델 파일 경로: ").strip()
            
            if not model_path:
                print(" 모델 파일 경로가 필요합니다.")
                return
            
            gpu_layers = input("GPU 레이어 수 [32]: ").strip()
            gpu_layers = int(gpu_layers) if gpu_layers else 32
            
            threads = input("스레드 수 [4]: ").strip()
            threads = int(threads) if threads else 4
            
            is_embedding = input("임베딩 모델인가요? (y/n) [n]: ").strip().lower()
            embedding = is_embedding == 'y'
            
            # 새 모델 추가
            model_info = {
                'name': model_name,
                'path': model_path,
                'gpu_layers': gpu_layers,
                'threads': threads,
                'embedding': embedding
            }
            
            config_manager.add_model(model_id, model_info)
            
            print(f"\n 모델 '{model_id}' 추가 완료!")
            print("config.json 파일이 업데이트되었습니다.")
            
        except Exception as e:
            print(f" 모델 추가 실패: {e}")
    
    @staticmethod
    def create_template():
        """설정 템플릿 생성"""
        template = {
            "_comment": "AI 빌드 서버 설정 파일 - 다중 모델 지원",
            
            "aws_access_key": "YOUR_AWS_ACCESS_KEY",
            "aws_secret_key": "YOUR_AWS_SECRET_KEY",
            "aws_region": "us-east-1",
            "instance_id": "i-1234567890abcdef0",
            
            "ssh_key_path": "./keys/server-key.pem",
            "ec2_user": "ubuntu",
            
            "base_port": 8080,
            "server_work_dir": "/home/ubuntu/llama.cpp",
            
            "models": {
                "qwen3-embedding": {
                    "name": "Qwen3 Embedding 0.6B",
                    "path": "/home/ubuntu/llama.cpp/models/qwen3-embedding-0.6b/Qwen3-Embedding-0.6B-Q8_0.gguf",
                    "gpu_layers": 32,
                    "threads": 4,
                    "embedding": True
                },
                "gpt-oss-20b": {
                    "name": "GPT OSS 20B",
                    "path": "/home/ubuntu/llama.cpp/models/gpt-oss-20b/gpt-oss-20b-f16.gguf",
                    "gpu_layers": 40,
                    "threads": 4,
                    "embedding": False
                }
            },
            
            # 기존 설정 호환성을 위해 유지 (deprecated)
            "server_port": 8080,
            "model_path": "/home/ubuntu/llama.cpp/models/qwen3-embedding-0.6b/Qwen3-Embedding-0.6B-Q8_0.gguf",
            "gpu_layers": 32,
            "threads": 4
        }
        
        with open('config.json.template', 'w', encoding='utf-8') as f:
            json.dump(template, f, indent=2, ensure_ascii=False)
        
        print("다중 모델 지원 설정 템플릿 생성: config.json.template")
        print("기존 config.json과 호환되며, models 섹션으로 확장 가능")