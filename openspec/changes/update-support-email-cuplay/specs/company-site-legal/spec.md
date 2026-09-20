## MODIFIED Requirements

### Requirement: 官网页脚 MUST 展示公司主体与备案信息

`resource/public/pangbao-home.html` 页脚 MUST 展示运营主体中文全称「杭州彩逗科技有限公司」、英文名「Hangzhou Caidou Technology Co., Ltd.」、ICP 备案号「浙ICP备2024101366号」（可点击链至工信部备案查询站）、客服邮箱 `pangbao@cuplay.top`，以及隐私政策与用户协议链接。页脚 MUST NOT 再展示「官网事件 logo 与 Android 下载链路均来自系统权威数据」类运维旁白。

#### Scenario: 用户查看官网底部

- **WHEN** 用户打开官网首页并滚动至页脚
- **THEN** 页面 SHALL 可见公司中英文全称、浙ICP备2024101366号与客服邮箱 `pangbao@cuplay.top`，且可进入隐私政策与用户协议
