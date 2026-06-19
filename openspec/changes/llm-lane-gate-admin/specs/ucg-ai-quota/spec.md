## MODIFIED Requirements

### Requirement: ucg polish SHALL pre-check and consume quota locally

`POST /ucg/app/api/posts/polish` MUST 在调用上游 LLM（`LanePolish`）前于 **ucg-service 进程内**执行 polish check；若 `allowed=false` MUST 返回 code **40302** 与 message **「本月额度已用完」** 且 MUST NOT 调用上游。check 通过后 MUST 经 polish lane 闸门；队列满 MUST 返回 **50301**。上游成功返回有效正文后 MUST 于本进程 consume。参数错误、未配置 AI、上游失败、50301 MUST NOT 调用 consume。

#### Scenario: 额度用尽

- **WHEN** 用户润笔 check 得到 used=5、limit=5
- **THEN** API SHALL 返回 40302 与「本月额度已用完」且 SHALL NOT 请求上游

#### Scenario: 上游失败不扣减

- **WHEN** check 通过但上游返回 5xx
- **THEN** 系统 SHALL NOT 调用 consume 且 used SHALL 不变

#### Scenario: 队列满

- **WHEN** check 通过但 polish 闸门队列满
- **THEN** API SHALL 返回 50301 且 MUST NOT consume
