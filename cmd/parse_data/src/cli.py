#!/usr/bin/env python3

import argparse
from jira_processor import JiraEmbeddingProcessor


def main():
    """CLI 메인 함수"""
    parser = argparse.ArgumentParser(
        description='JIRA 데이터를 h4 섹션별로 나누어 임베딩 요청 형식으로 전처리 (빈 내용/N/A 제외)'
    )
    
    parser.add_argument(
        'input_file', 
        nargs='?', 
        help='입력 JSON 파일 경로'
    )
    
    parser.add_argument(
        '-o', '--output', 
        default='embedding_requests.jsonl',
        help='출력 파일 경로 (기본값: embedding_requests.jsonl)'
    )
    
    parser.add_argument(
        '-f', '--format', 
        choices=['jsonl', 'json'], 
        default='jsonl', 
        help='출력 형식 (기본값: jsonl)'
    )
    
    parser.add_argument(
        '--test', 
        action='store_true', 
        help='샘플 데이터로 테스트 실행'
    )
    
    parser.add_argument(
        '--include-empty', 
        action='store_true', 
        help='N/A나 빈 내용도 포함 (기본값: 제외)'
    )
    
    args = parser.parse_args()
    
    # JiraEmbeddingProcessor 인스턴스 생성
    processor = JiraEmbeddingProcessor(include_empty=args.include_empty)
    
    # 실행 모드 결정 및 처리
    if args.test or not args.input_file:
        # 테스트 모드
        success = processor.process_sample_data(args.output, args.format)
    else:
        # 파일 처리 모드
        success = processor.process_from_file(args.input_file, args.output, args.format)
    
    if not success:
        exit(1)


if __name__ == "__main__":
    main()