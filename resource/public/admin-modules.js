// Web 管理模块登记（唯一入口：App 网关同源）。
window.ADMIN_MODULES = [
	{
		id: 'version-admin',
		title: 'App 版本管理',
		pagePath: '/device/app/version-admin.html',
		showInNav: true
	},
	{
		id: 'ucg-admin',
		title: 'UCG 管理',
		pagePath: '/device/admin/ucg-admin.html',
		showInNav: true
	},
	{
		id: 'ai-model-admin',
		title: 'AI 模型与并发',
		pagePath: '/device/admin/ai-model-admin.html',
		showInNav: true
	},
	{
		id: 'voice-admin',
		title: 'Voice AI 额度',
		pagePath: '/device/admin/voice-admin.html',
		showInNav: true
	},
	{
		id: 'intent-agent-admin',
		title: '意图智能体',
		pagePath: '/device/admin/intent-agent-admin.html',
		showInNav: true
	},
	{
		id: 'intent-vector-admin',
		title: '事件向量',
		pagePath: '/device/admin/intent-vector-admin.html',
		showInNav: false
	},
	{
		id: 'type-vector-admin',
		title: '类型向量',
		pagePath: '/device/admin/type-vector-admin.html',
		showInNav: false
	},
	{
		id: 'op-vector-admin',
		title: 'OP 轨飞轮',
		pagePath: '/device/admin/op-vector-admin.html',
		showInNav: false
	},
	{
		id: 'qty-vector-admin',
		title: '数量飞轮',
		pagePath: '/device/admin/qty-vector-admin.html',
		showInNav: false
	},
	{
		id: 'read-source-vector-admin',
		title: 'R·数据源飞轮',
		pagePath: '/device/admin/read-source-vector-admin.html',
		showInNav: false
	},
	{
		id: 'history-speak-vector-admin',
		title: '查史播报向量',
		pagePath: '/device/admin/history-speak-vector-admin.html',
		showInNav: false
	},
	{
		id: 'rule-gaps-admin',
		title: '规则缺口',
		pagePath: '/device/admin/rule-gaps-admin.html',
		showInNav: false
	},
	{
		id: 'sim-admin',
		title: '模拟用户管理',
		pagePath: '/device/admin/sim-admin.html',
		showInNav: true
	},
	{
		id: 'cash-vip-admin',
		title: 'VIP 权益',
		pagePath: '/device/admin/cash-vip-admin.html',
		showInNav: false
	},
	{
		id: 'cash-feature-admin',
		title: '开通功能管理',
		pagePath: '/device/admin/cash-feature-admin.html',
		showInNav: true
	},
	{
		id: 'cash-xiaozhi-mcp-admin',
		title: '小智 MCP 开通',
		pagePath: '/device/admin/cash-xiaozhi-mcp-admin.html',
		showInNav: true
	},
	{
		id: 'cash-feeding-eligibility-admin',
		title: '喂养资格门槛',
		pagePath: '/device/admin/cash-feeding-eligibility-admin.html',
		showInNav: true
	},
	{
		id: 'push-device-admin',
		title: '推送设备',
		pagePath: '/device/admin/push-device-admin.html',
		showInNav: true
	},
	{
		id: 'app-status-admin',
		title: 'App 维护通知',
		externalUrl: 'https://notify.cuplay.top/admin',
		openInNewTab: true,
		showInNav: true
	},
	{
		id: 'swagger',
		title: 'API 接口调试',
		pagePath: '/swagger',
		showInNav: true
	},
	{
		id: 'api-usage-stats',
		title: '功能使用统计',
		pagePath: '/device/admin/api-usage-stats',
		showInNav: false,
		showInDeviceRecord: true
	},
	{
		id: 'client-usage-stats',
		title: '客户端使用统计',
		pagePath: '/device/admin/client-usage-stats',
		showInNav: false,
		showInDeviceRecord: true
	}
];
