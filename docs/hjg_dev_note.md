## 21_11_29
### 21_11_29 routing package golila/mux -> fiber/v2 마이그레이션 적용
 - main.go & room.go & user.go & route.go : fiber/v2로 교체
 
## 21_11_30
### 21_11_30 redis 추가 및 redis 연동 추가 & RMQ 큐 생성 연동 완료
 - redis 추가 및 redis 연동 추가 & RMQ 큐 생성 연동 완료

## 21_12_03
### 21_12_03 배포 직전 커밋
 - 배포 직전 커밋

## 22_01_19
### 22_01_19 아키텍쳐 수정중 커밋

## 22_02_10
 ### 22_02_10 생략된 로직 추가
 - rabbitmq_service.go : private된 메서드 인터페이스에 추가후 public으로 수정
 - rabbitmq_handler.go : 방생성 이후 생략된 consume 로직 추가