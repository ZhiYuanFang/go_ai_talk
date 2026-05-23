## ADDED Requirements

### Requirement: App 网关按主机白名单回显 CORS Origin

`gateway-app-server` 对浏览器跨域请求 SHALL 在响应中包含 CORS 头。当且仅当请求头 `Origin` 解析成功且其主机（不含端口比较 IP 字面量，含端口时取 hostname）等于 `192.168.0.131` 或 `120.55.50.105`，且 scheme 为 `http` 或 `https` 时，SHALL 将 `Access-Control-Allow-Origin` 设为该 `Origin` 的完整原始值（回显），从而允许该主机上任意端口的 Web 来源。

#### Scenario: 匹配内网 IP 任意端口

- **WHEN** 请求包含 `Origin: http://192.168.0.131:5173` 且请求到达 `gateway-app-server` 的 HTTP 处理链
- **THEN** 响应包含 `Access-Control-Allow-Origin: http://192.168.0.131:5173`

#### Scenario: 匹配公网 IP 任意端口

- **WHEN** 请求包含 `Origin: https://120.55.50.105:8443` 且请求到达 `gateway-app-server` 的 HTTP 处理链
- **THEN** 响应包含 `Access-Control-Allow-Origin: https://120.55.50.105:8443`

#### Scenario: 非白名单主机不回显

- **WHEN** 请求包含 `Origin: https://evil.example` 且请求到达 `gateway-app-server` 的 HTTP 处理链
- **THEN** 响应 SHALL NOT 设置 `Access-Control-Allow-Origin`（或不得回显该 Origin）

### Requirement: CORS 方法与请求头

`gateway-app-server` SHALL 在 CORS 响应中声明允许方法包含 `GET`、`POST`、`OPTIONS`，并 SHALL 在 `Access-Control-Allow-Headers`（或对预检的等价响应）中允许 `Content-Type` 与 `Authorization`，以满足常见 JSON 与 Bearer 联调。

#### Scenario: 预检请求获得方法与头

- **WHEN** 浏览器发送 `OPTIONS` 预检，且 `Origin` 通过主机白名单校验，且带有 `Access-Control-Request-Method: POST` 与 `Access-Control-Request-Headers: content-type, authorization`
- **THEN** 响应状态码为成功（2xx），且包含允许上述方法与头的 CORS 响应头（具体头名大小写遵循实现，语义须满足浏览器识别）

### Requirement: 预检不破坏既有鉴权豁免

对 `OPTIONS` 请求的 Bearer 豁免行为 SHALL 保持与变更前一致：预检请求 MUST NOT 因缺少 Bearer 被拒绝（例如 401）。

#### Scenario: OPTIONS 无 Bearer 仍成功预检

- **WHEN** `OPTIONS` 请求指向需鉴权的 API 路径，且无 `Authorization` 头，但 `Origin` 通过白名单
- **THEN** 响应 SHALL NOT 仅因缺少 Bearer 而返回 401（允许返回 204 或其它 2xx 并完成 CORS 头）
