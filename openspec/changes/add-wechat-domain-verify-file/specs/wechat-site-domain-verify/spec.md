## ADDED Requirements

### Requirement: gateway-app MUST 在根路径提供微信域名校验 TXT

`gateway-app-server` MUST 对 `GET` 与 `HEAD` 请求路径 `/90fafbe9bf8308ecbd2063d7b479a309.txt` 返回 HTTP 200，响应体为纯文本校验串 `41a8fe99b666f6646724f29ea87962e34c6df2b6`（不得包裹 HTML、JSON 或网关统一业务 envelope）。静态文件 MUST 存放于 `resource/public/90fafbe9bf8308ecbd2063d7b479a309.txt`，并由显式路由提供（与官网根路径静态页同一托管进程）。该路径 MUST 允许匿名访问（Bearer 豁免），且 MUST NOT 计入 App API usage 统计。

#### Scenario: 微信爬虫校验域名归属

- **WHEN** 未登录客户端请求 `GET /90fafbe9bf8308ecbd2063d7b479a309.txt`（公网形态为 `https://www.pangbao.cuplay.top/90fafbe9bf8308ecbd2063d7b479a309.txt`）
- **THEN** 系统 SHALL 返回 200，且响应体 SHALL 为 `41a8fe99b666f6646724f29ea87962e34c6df2b6`

#### Scenario: 校验路径不要求登录

- **WHEN** 请求未携带 App Bearer Token
- **THEN** 系统 SHALL 仍返回校验 TXT，MUST NOT 因鉴权失败而 401 或跳转登录页

#### Scenario: 校验路径不计入 usage

- **WHEN** 客户端访问 `/90fafbe9bf8308ecbd2063d7b479a309.txt`
- **THEN** 系统 SHALL NOT 将该请求记入 App API 使用统计
