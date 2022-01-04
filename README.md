# chat_server_api 디테일

![e24d9680-e94d-11e9-977a-54046597cf22.png](/chat_server_api%20디테일%20bfb7b12835be4b89b484333f06e9b38c/e24d9680-e94d-11e9-977a-54046597cf22.png)

## RabbitMQ + Golang 채팅 서버

- 전체 네트워크 구상도
    
    ![KakaoTalk_Image_2022-01-02-15-25-39.jpeg](/chat_server_api%20디테일%20bfb7b12835be4b89b484333f06e9b38c/KakaoTalk_Image_2022-01-02-15-25-39.jpeg)
    
- chat_server
    
    ![KakaoTalk_Photo_2022-01-02-15-26-09.jpeg](/chat_server_api%20디테일%20bfb7b12835be4b89b484333f06e9b38c/KakaoTalk_Photo_2022-01-02-15-26-09.jpeg)
    
    - **단일 고루틴 역할**
        1. 각 채팅방 메시지큐 sub
        2. 각 채팅방에 pub이 들어오면 채팅방 유저 메시지큐에 라우팅키로 한번에 pub
        3. pub 이후 메시지의 종류에 따라 역할 수행
            1. 메시지 삭제
            2. 유저 방 탈퇴
            3. 유저 방 초대
- **rabbitmq 메시지큐 구상도**
    - **status 구성 확인 방법**
        
        > [rmq-admin.go-talk.kr](http://rmq-admin.go-talk.kr)
        > 
        > - 아디 비번
        >     - ~~보안상 비밀...~~
        
        ![KakaoTalk_Image_2022-01-02-15-35-27.jpeg](/chat_server_api%20디테일%20bfb7b12835be4b89b484333f06e9b38c/KakaoTalk_Image_2022-01-02-15-35-27.jpeg)
        
    - **메시지큐 역할**
        1. 검은 메시지 큐는 각 채팅방의 pub하는 send 전용 메시지큐.
        2. 파란 메시지 큐는 각 채팅방 유저에게 sub되고 있는 receive 전용 메시지 큐, 유저가 방을 sub을 시작하는 순간 생성된다.
- **ERD**
    - chat_server_db
        
        ![스크린샷_2022-01-02_오후_2.59.15.png](/chat_server_api%20디테일%20bfb7b12835be4b89b484333f06e9b38c/스크린샷_2022-01-02_오후_2.59.15.png)
        
        각 상태 테이블을 넣은 이유는 로그의 목적이 제일 크다.
        
        쿼리 조합에 따라 해볼수 있는 것이 많을 것 같아서 따로 분리를 해두었다.
        
    - chat_log_db
        
        ![스크린샷_2022-01-02_오후_2.59.15.png](/chat_server_api%20디테일%20bfb7b12835be4b89b484333f06e9b38c/스크린샷_2022-01-02_오후_2.59.15.png)
        
- **chat_server 중요 spec**
    - go 1.16.6
    - fiber
    - gorm [매핑 목적]
    - amqp [rabbitmq sub 목적]
    - redigo [레디스 접속 목적]
    - crypto [유저 비번 암호화]
    - postgres [db 드라이버]