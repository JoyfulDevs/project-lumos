#!/usr/bin/env python3

import json
import re
from typing import List, Dict, Any, Optional
from pathlib import Path


class JiraDescriptionProcessor:

    def __init__(self, include_empty: bool = False):

        self.include_empty = include_empty
        self.na_patterns = [
            'n/a', 'na', 'n.a', 'n.a.', 'not applicable', 'not available',
            '해당없음', '해당 없음', '없음', 'tbd', 'to be determined'
        ]
    
    def extract_jira_description(self, jira_data: Dict[str, Any]) -> str:
        try:
            return jira_data.get('fields', {}).get('description', '')
        except Exception as e:
            print(f"Description 추출 중 오류: {e}")
            return ""
    
    def is_meaningful_content(self, content: str) -> bool:
        if not content:
            return False
        
        lines = content.split('\n')
        content_lines = []
        
        for line in lines:
            stripped = line.strip()
            if not stripped.startswith('h4.'):
                content_lines.append(stripped)
        
        actual_content = '\n'.join(content_lines).strip()
        
        if not actual_content:
            return False
        
        cleaned_content = re.sub(r'[\r\n\t\s]+', ' ', actual_content).strip()
        cleaned_content = re.sub(r'^[-\s]*$', '', cleaned_content)
        
        if not cleaned_content:
            return False
        
        cleaned_lower = cleaned_content.lower().strip()
        
        if cleaned_lower in self.na_patterns:
            return False
        

        if cleaned_lower.startswith('-'):
            content_after_dash = cleaned_lower[1:].strip()
            if content_after_dash in self.na_patterns:
                return False
        

        if len(cleaned_content) <= 2:
            return False
        
        return True
    
    def split_by_h4_sections(self, description: str) -> List[Dict[str, str]]:
        """description을 h4 섹션별로 분할"""
        if not description.strip():
            return []
        

        if not re.search(r'h4\.\s+', description):
            return []
        

        sections = re.split(r'(h4\.\s+[^\r\n]+)', description)
        

        if sections and not sections[0].strip():
            sections = sections[1:]
        
        result = []
        

        for i in range(0, len(sections), 2):
            if i < len(sections):
                title = sections[i].strip() if i < len(sections) else ""
                content = sections[i + 1].strip() if i + 1 < len(sections) else ""
                
                if title.startswith('h4.'):

                    full_section = title
                    if content:
                        full_section += "\r\n" + content
                    
                    if self.include_empty or self.is_meaningful_content(full_section):
                        result.append({
                            "title": title,
                            "content": full_section
                        })
                    else:
                        print(f"  - 빈 내용으로 인해 제외된 섹션: {title}")
        
        return result
    
    def create_embedding_requests(self, jira_key: str, sections: List[Dict[str, str]]) -> List[Dict[str, Any]]:
        """섹션들을 임베딩 요청 형식으로 변환"""
        requests = []
        
        for idx, section in enumerate(sections, 1):
            request = {
                "custom_id": f"{jira_key}-content-{idx}",
                "method": "POST",
                "url": "/v1/embeddings",
                "body": {
                    "model": "text-embedding-3-small",
                    "input": section["content"],
                    "encoding_format": "float"
                }
            }
            requests.append(request)
        
        return requests
    
    def process_single_jira(self, jira_data: Dict[str, Any]) -> List[Dict[str, Any]]:
        """단일 JIRA 데이터를 처리하여 임베딩 요청들 생성"""
        try:
            jira_key = jira_data.get('key', 'UNKNOWN')
            description = self.extract_jira_description(jira_data)
            
            if not description.strip():
                print(f"Warning: {jira_key}에 description이 없습니다.")
                return []
            
            sections = self.split_by_h4_sections(description)
            
            if not sections:
                print(f"Warning: {jira_key}에서 h4 섹션을 찾을 수 없습니다.")
                request = {
                    "custom_id": f"{jira_key}-content-1",
                    "method": "POST",
                    "url": "/v1/embeddings",
                    "body": {
                        "model": "text-embedding-3-small",
                        "input": description,
                        "encoding_format": "float"
                    }
                }
                return [request]
            else:
                section_requests = self.create_embedding_requests(jira_key, sections)
                print(f"{jira_key}: {len(sections)}개 의미있는 섹션으로 분할되었습니다.")
                return section_requests
        
        except Exception as e:
            print(f"JIRA 데이터 처리 중 오류 ({jira_data.get('key', 'UNKNOWN')}): {e}")
            return []
    
    def process_jira_data_list(self, jira_data_list: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        all_requests = []
        
        for jira_data in jira_data_list:
            requests = self.process_single_jira(jira_data)
            all_requests.extend(requests)
        
        return all_requests


class JiraDataLoader:
    
    @staticmethod
    def load_from_file(file_path: str) -> List[Dict[str, Any]]:
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                data = json.load(f)
            
            if isinstance(data, list):
                return data
            elif isinstance(data, dict):
                return [data]
            else:
                print(f"Error: 지원하지 않는 데이터 형식입니다.")
                return []
        
        except Exception as e:
            print(f"파일 로드 중 오류: {e}")
            return []
    
    @staticmethod
    def get_sample_data() -> List[Dict[str, Any]]:
        return [
            {
                "key": "GS-10691",
                "fields": {
                    "description": "h4. Why will be improved (개선사유)\r\n- 일부 필드는 빈 문자열이 서버에서 파싱 에러를 발생시킨다.\r\n\r\nh4. What will be improved (개선목표)\r\n- 인코딩시 필드를 생략하도록 설정.\r\n\r\nh4. How do users configure and use it? (사용자가 어떻게 설정 및 사용 하는가?)\r\n- N/A\r\n\r\nh4. How is it implemented? (어떻게 구현 하는가?)\r\n- N/A\r\n\r\nh4. Migration (마이그레이션)\r\n- N/A\r\n\r\nh4. Test Case (테스트 방법)\r\n- N/A\r\n\r\nh4. Deployment impacts and procedures (배포 영향 및 절차)\r\n- N/A"
                }
            }
        ]


class EmbeddingRequestSaver:
    
    @staticmethod
    def save_requests(requests: List[Dict[str, Any]], output_path: str, format_type: str = 'jsonl') -> bool:
        try:
            output_file = Path(output_path)
            
            if format_type == 'jsonl':
                # JSONL 형식으로 저장 (각 줄마다 하나의 JSON 객체)
                with open(output_file, 'w', encoding='utf-8') as f:
                    for request in requests:
                        f.write(json.dumps(request, ensure_ascii=False) + '\n')
            else:
                # JSON 배열 형식으로 저장
                with open(output_file, 'w', encoding='utf-8') as f:
                    json.dump(requests, f, ensure_ascii=False, indent=2)
            
            print(f"결과가 {output_file}에 저장되었습니다. ({len(requests)}개 요청)")
            return True
        
        except Exception as e:
            print(f"파일 저장 중 오류: {e}")
            return False
    
    @staticmethod
    def preview_first_request(requests: List[Dict[str, Any]]) -> None:
        if requests:
            print("\n=== 첫 번째 요청 미리보기 ===")
            print(json.dumps(requests[0], ensure_ascii=False, indent=2))


class JiraEmbeddingProcessor:
    
    def __init__(self, include_empty: bool = False):
        self.processor = JiraDescriptionProcessor(include_empty)
        self.loader = JiraDataLoader()
        self.saver = EmbeddingRequestSaver()
    
    def process_from_file(self, input_path: str, output_path: str = 'embedding_requests.jsonl', 
                         format_type: str = 'jsonl') -> bool:
        print(f"JIRA 데이터 로딩중: {input_path}")
        
        jira_data = self.loader.load_from_file(input_path)
        if not jira_data:
            print("로드할 데이터가 없습니다.")
            return False
        
        return self._process_and_save(jira_data, output_path, format_type)
    
    def process_sample_data(self, output_path: str = 'embedding_requests.jsonl', 
                          format_type: str = 'jsonl') -> bool:

        print("샘플 데이터로 테스트 실행중...")
        
        jira_data = self.loader.get_sample_data()
        return self._process_and_save(jira_data, output_path, format_type)
    
    def process_data_list(self, jira_data: List[Dict[str, Any]], 
                         output_path: str = 'embedding_requests.jsonl', 
                         format_type: str = 'jsonl') -> bool:

        return self._process_and_save(jira_data, output_path, format_type)
    
    def _process_and_save(self, jira_data: List[Dict[str, Any]], 
                         output_path: str, format_type: str) -> bool:

        embedding_requests = self.processor.process_jira_data_list(jira_data)
        
        if not embedding_requests:
            print("처리할 요청이 없습니다.")
            return False
        
        success = self.saver.save_requests(embedding_requests, output_path, format_type)
        
        if success:
            self.saver.preview_first_request(embedding_requests)
        
        return success


# 편의 함수들
def process_jira_file(input_path: str, output_path: str = 'embedding_requests.jsonl', 
                     format_type: str = 'jsonl', include_empty: bool = False) -> bool:

    processor = JiraEmbeddingProcessor(include_empty)
    return processor.process_from_file(input_path, output_path, format_type)


def process_jira_data(jira_data: List[Dict[str, Any]], output_path: str = 'embedding_requests.jsonl', 
                     format_type: str = 'jsonl', include_empty: bool = False) -> bool:

    processor = JiraEmbeddingProcessor(include_empty)
    return processor.process_data_list(jira_data, output_path, format_type)