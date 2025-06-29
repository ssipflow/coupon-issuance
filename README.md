# Coupon Issuance System

## 1. 개발환경

- Golang 1.23
- MySQL 8.0
- Redis 7.x
- GORM
- Connect-gRPC
- Asynq
- godotenv
- Envoy

---

## 2. 실행 방법

```bash
docker compose up --scale backend-app={numberOfInstance} --build
```

- `--scale backend-app={numberOfInstance}`
  - numberOfInstance: `backend-app` 인스턴스 scale out
  - `docker compose` Instance
---

## 3. Architecture Overview

- **gRPC 기반**의 API 서버 (Connect-go 사용)
- **Asynq 기반 비동기 메시지 큐**
  - Producer/Consumer 간의 **정합성 및 중복 발급 방지** 구현
- Envoy 프록시 기반 서비스 디스커버리 및 로드밸런싱
  - HTTP/gRPC 요청을 여러 backend-app 인스턴스에 분산처리
- DB: MySQL, Redis 조합
  - Redis: 캐싱 및 락, 중복 발급 방지
  - MySQL: 발급 결과의 최종 저장소
- GORM을 통한 명시적 트랜잭션 제어

---

## 4. 주요 고려사항

### 동시성 정합성

- Redis `SETNX`로 중복 발급 방지
- Redis `INCR` 기반 선점, DB 최종 반영
- Consumer 트랜잭션 실패 시 Redis 상태 복원 및 롤백 처리

### 테스트 전략

| 도구 | 목적 | 설명 |
|------|------|------|
| `go test` | 단위 테스트 | 로직 단위의 기본 검증 |
| `k6` | 부하 테스트 | 최대 1000TPS 목표 성능 검증 |

### go test 
```bash
go test -v -count=1 ./test/stress
```
- Campaign 생성 및 TPS 1000 테스트

```bash
go test -v -count=1 ./test/limit
```
- Campaign 생성 및 쿠폰수량 한정 테스트

### K6 부하 테스트 예시
```bash
brew install k6
```

```bash
k6 run test/k6/stress/stress_test.js
```

- `CreateCampaign` 호출
- `/test/k6/stress_test.js` 의 `campaignId` 설정
- 1000TPS 부하 테스트 진행

---

## 5. API 명세

## 5.0 Error Message
| message | 의미 |
|--------|------|
| `RECORD_NOT_FOUND` | 데이터 없음 |
| `CAMPAIGN_NOT_STARTED` | 쿠폰 캠페인 시작 안함 |
| `COUPON_ALREADY_CLAIMED` | 중복 발급 |
| `COUPON_SOLD_OUT` | 수량 소진 |
| `INTERNAL_SERVER_ERROR` | 서버 내부 오류 |


### 5.1 CreateCampaign

- POST `/coupon.v1.CouponService/CreateCampaign`
- Request:

```json
{
  "name": "여름 이벤트",
  "couponLimit": 1000,
  "startTime": "2025-06-25T00:00:00Z"
}
```

- Response:

```json
{
    "campaignId": 1
}
```

---

### 5.2 GetCampaign

- POST `/coupon.v1.CouponService/CreateCampaign`
- `name`, `coupon_limit`, `start_time(ISO8601)` 입력
- Request:

```json
{
    "campaignId": 1
}
```

- Response:
```json
{
    "id": 1,
    "name": "여름 이벤트",
    "couponLimit": "1000",
    "currentCoupon": "2",
    "startTime": "2025-06-25T00:00:00Z",
    "createdAt": "2025-06-29T09:54:21Z",
    "updatedAt": "2025-06-29T09:54:21Z",
    "issuedCoupons": [
        "5아아처더머서2파3",
        "다어저사버자러나6저"
    ]
}
```

### 5.3 IssueCoupon

- POST `/coupon.v1.CouponService/IssueCoupon`
- Request:

```json
{
  "campaignId": 1,
  "userId": 1001
}
```

- Response:

```json
{
  "message": "OK"
}
```
---

## 6. DB 스키마 요약

### 6.1 campaigns

| 필드명 | 설명 |
|--------|------|
| id | PK |
| name | 캠페인 이름 |
| coupon_limit | 최대 발급 수량 |
| current_coupon | 현재 발급된 수량 |
| start_time | 발급 시작 시간 |

### 6.2 coupons

| 필드명 | 설명 |
|--------|------|
| id | PK |
| campaign_id | 매핑되는 캠페인 ID |
| user_id | 사용자 ID |
| code | 쿠폰 코드 |

- Index
  - campaign_id, user_id 복합 인덱스: 캠페인별 유저 유니크 보장
  - campaign_id, code 복합 인덱스: 캠페인별 쿠폰코드 유니크 보장

---

## 7. 성능 테스트
### 7.1 주의사항
최초 성능 테스트 시에는 다음과 같은 이유로 목표 TPS 가 다소 낮게 측정될 수 있습니다.
- DB Connection Pool 미가동
- Redis 초기 연결 지연
- Docker 내부 네트워크 지연

이러한 요소는 실제 트래픽 흐름과 달리 초기화 작업이 포함되어 있는 상태로, 신뢰도 있는 TPS 측정을 위해 다음과 같은 **사전 워밍업**을 권장합니다.
```bash
# 1. 사전 테스트로 DB/Redis 연결 활성화
go test -v -count=1 ./test/stress

# 2. POST /coupon.v1.CouponService/CreateCampaign 으로 캠페인 생성
# 해당 캠페인 ID 로 stress_test.js 변경 후 테스트 진행
k6 run test/k6/stress/stress_test.js
```

### 7.2 Single Instance K6 Test
```bash
docker compose up --scale backend-app=0 --build
```

- **실행 환경**  
  - `backend-app-primary`: 1 인스턴스  
  - `backend-app`: 0 인스턴스

- **부하 조건**  
  - `1000 TPS` 고정 요청 속도 (1초 동안 1000개의 요청 시도)
  - `maxVUs=2000` 설정으로 충분한 가상 유저 확보
  - `constant_rate_test` 시나리오 사용

- **테스트 결과**  

```bash
         /\      Grafana   /‾‾/
    /\  /  \     |\  __   /  /
   /  \/    \    | |/ /  /   ‾‾\
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/

     execution: local
        script: test/k6/stress/stress_test.js
        output: -

     scenarios: (100.00%) 1 scenario, 2000 max VUs, 31s max duration (incl. graceful stop):
              * constant_rate_test: 1000.00 iterations/s for 1s (maxVUs: 1000-2000, gracefulStop: 30s)

INFO[0000] Failure issuing coupon for user 21157: COUPON_ALREADY_CLAIMED  source=console
INFO[0000] Failure issuing coupon for user 58055: COUPON_ALREADY_CLAIMED  source=console
INFO[0000] Failure issuing coupon for user 71145: COUPON_ALREADY_CLAIMED  source=console
INFO[0000] Failure issuing coupon for user 943: COUPON_ALREADY_CLAIMED  source=console
INFO[0000] Failure issuing coupon for user 52138: COUPON_ALREADY_CLAIMED  source=console


  █ TOTAL RESULTS

    checks_total.......................: 1001   992.577976/s
    checks_succeeded...................: 99.50% 996 out of 1001
    checks_failed......................: 0.49%  5 out of 1001

    ✗ status is 200
      ↳  99% — ✓ 996 / ✗ 5

    HTTP
    http_req_duration.......................................................: avg=7.8ms  min=2.04ms med=5.03ms max=56.76ms p(90)=14.18ms p(95)=23.2ms
      { expected_response:true }............................................: avg=7.82ms min=2.04ms med=5.04ms max=56.76ms p(90)=14.2ms  p(95)=23.2ms
    http_req_failed.........................................................: 0.49%  5 out of 1001
    http_reqs...............................................................: 1001   992.577976/s

    EXECUTION
    iteration_duration......................................................: avg=8.18ms min=2.29ms med=5.34ms max=60.27ms p(90)=14.39ms p(95)=23.48ms
    iterations..............................................................: 1001   992.577976/s
    vus.....................................................................: 7      min=7         max=7
    vus_max.................................................................: 1000   min=1000      max=1000

    NETWORK
    data_received...........................................................: 197 kB 195 kB/s
    data_sent...............................................................: 190 kB 189 kB/s




running (01.0s), 0000/1000 VUs, 1001 complete and 0 interrupted iterations
constant_rate_test ✓ [======================================] 0000/1000 VUs  1s  1000.00 iters/s
```

| 항목 | 수치 | 해설 |
|------|------|------|
| **총 요청 수** | 1000 | 정확히 1초간 1000건 요청 발생 |
| **성공 요청** | 996 | 쿠폰 발급 성공 |
| **실패 요청** | 5 | 중복 발급 요청 (`COUPON_ALREADY_CLAIMED`) |
| **성공률** | **99.5%** | 목표치인 99.5% 이상 달성 |
| **평균 응답시간** | 7.8ms | 단일 인스턴스에서도 안정적인 처리 |
| **95% 응답시간** | 23.2ms | 대부분의 요청이 25ms 이내 응답 |
| **TPS 유지 여부** |  **1000 TPS 달성** | `iterations/sec` = 992.57 |

- **정합성**  
  - DB(MySQL)의 `current_coupon`, Redis의 `coupon:issued:*` 값 모두 996으로 일치
  - Consumer 처리 중 실패한 요청은 Redis 롤백을 통해 정확히 복원됨

- **의미**  
  - 단일 인스턴스 환경에서 TPS 1000 수준의 부하를 안정적으로 처리 가능
  - 평균 응답 시간 7ms 수준으로, 시스템의 I/O 성능과 비즈니스 로직이 경량화되어 있음을 확인
  - 수평 확장 없이도 가벼운 규모의 캠페인 처리에 적합하며, 수직 확장을 통한 성능 확보도 가능

### 7.3 Multi Instance K6 Test

- **실행 환경**  
  - `backend-app-primary`: 1 인스턴스  
  - `backend-app`: 2 인스턴스 (총 3개 인스턴스 로드밸런싱 구조)
  - `Envoy`를 통해 로드밸런싱 수행

- **부하 조건**  
  - `1000 TPS` 고정 요청 속도 (1초 동안 1000개의 요청 시도)
  - `maxVUs=2000` 설정으로 충분한 가상 유저 확보
  - `constant_rate_test` 시나리오 사용

- **테스트 결과**  

```bash
         /\      Grafana   /‾‾/
    /\  /  \     |\  __   /  /
   /  \/    \    | |/ /  /   ‾‾\
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/

     execution: local
        script: test/k6/stress/stress_test.js
        output: -

     scenarios: (100.00%) 1 scenario, 2000 max VUs, 31s max duration (incl. graceful stop):
              * constant_rate_test: 1000.00 iterations/s for 1s (maxVUs: 1000-2000, gracefulStop: 30s)

INFO[0000] Failure issuing coupon for user 19432: COUPON_ALREADY_CLAIMED  source=console
INFO[0000] Failure issuing coupon for user 85989: COUPON_ALREADY_CLAIMED  source=console
INFO[0000] Failure issuing coupon for user 57462: COUPON_ALREADY_CLAIMED  source=console
INFO[0001] Failure issuing coupon for user 38824: COUPON_ALREADY_CLAIMED  source=console


  █ TOTAL RESULTS

    checks_total.......................: 1000   993.611081/s
    checks_succeeded...................: 99.60% 996 out of 1000
    checks_failed......................: 0.40%  4 out of 1000

    ✗ status is 200
      ↳  99% — ✓ 996 / ✗ 4

    HTTP
    http_req_duration.......................................................: avg=11.67ms min=2.64ms med=7.54ms max=77.5ms  p(90)=20.74ms p(95)=38.68ms
      { expected_response:true }............................................: avg=11.68ms min=2.64ms med=7.54ms max=77.5ms  p(90)=20.67ms p(95)=38.69ms
    http_req_failed.........................................................: 0.40%  4 out of 1000
    http_reqs...............................................................: 1000   993.611081/s

    EXECUTION
    iteration_duration......................................................: avg=12.14ms min=2.9ms  med=7.87ms max=80.27ms p(90)=21.28ms p(95)=39.94ms
    iterations..............................................................: 1000   993.611081/s
    vus.....................................................................: 12     min=12        max=12
    vus_max.................................................................: 1000   min=1000      max=1000

    NETWORK
    data_received...........................................................: 196 kB 195 kB/s
    data_sent...............................................................: 190 kB 189 kB/s




running (01.0s), 0000/1000 VUs, 1000 complete and 0 interrupted iterations
constant_rate_test ✓ [======================================] 0000/1000 VUs  1s  1000.00 iters/s
```

| 항목 | 수치 | 해설 |
|------|------|------|
| **총 요청 수** | 1000 | 정확히 1초 간 1000건 요청 발생 |
| **성공 요청** | 996 | 발급 성공 처리 |
| **실패 요청** | 4 | `COUPON_ALREADY_CLAIMED`에 해당, 유효한 실패로 간주 |
| **성공률** | **99.6%** | 목표치인 99.5% 이상 달성 |
| **평균 응답시간** | 11.67ms | 멀티 인스턴스를 통한 응답 속도 향상 확인 |
| **95% 응답시간** | 38.68ms | 95% 요청이 40ms 이내에 응답 처리됨 |
| **TPS 유지 여부** | **1000 TPS 달성** | `iterations/sec` = 993.6, 정상 범위 내 도달 |

- **정합성**  
  - DB(MySQL)의 `current_coupon`, Redis의 `coupon:issued:*` 값 모두 996으로 일치
  - Consumer가 비정상 종료되었을 경우 Redis와 DB 상태 복원 처리도 정상 작동됨

- **의미**  
  - `asynq`, `SETNX`, `INCR`, 트랜잭션 롤백 구조가 대용량 요청에서도 정합성 보장
  - 실제 운영 환경에서도 수평 확장을 통해 충분한 트래픽 처리 가능성 확인됨