## ADDED Requirements

### Requirement: 意图智能体门户与导航收口

Hub 运维导航 SHALL 以单一入口「意图智能体」暴露意图智能体相关工具，并 MUST 提供门户页列出子能力入口。下列模块 MUST NOT 再作为顶栏平级 `showInNav` 入口：意图向量缓存页、类型向量页、规则缺口页。上述子页的既有 URL MUST 仍可直接访问（深链）。门户 MUST 至少包含指向以下能力的入口：事件向量（原意图向量缓存页）、类型向量、查史播报向量、规则缺口。

#### Scenario: 顶栏仅见意图智能体

- **WHEN** 运维在 Hub 主页查看导航链接（`showInNav`）
- **THEN** 可见「意图智能体」入口，且 MUST NOT 再平级展示「意图向量缓存」「类型向量」「规则缺口」三个独立导航项

#### Scenario: 门户聚合四子入口

- **WHEN** 运维打开意图智能体门户页
- **THEN** 页面提供前往事件向量、类型向量、查史播报向量、规则缺口的链接或等价入口

#### Scenario: 旧深链仍可用

- **WHEN** 运维直接访问既有 `intent-vector-admin.html` / `type-vector-admin.html` / `rule-gaps-admin.html` 路径
- **THEN** 页面仍可加载（不因导航收口而删除路由）
