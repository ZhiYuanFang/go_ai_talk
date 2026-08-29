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
		id: 'sim-admin',
		title: '模拟用户管理',
		pagePath: '/device/admin/sim-admin.html',
		showInNav: true
	},
	{
		id: 'cash-vip-admin',
		title: 'VIP 权益',
		pagePath: '/device/admin/cash-vip-admin.html',
		showInNav: true
	},
	{
		id: 'cash-feature-admin',
		title: '开通功能管理',
		pagePath: '/device/admin/cash-feature-admin.html',
		showInNav: true
	},
	{
		id: 'cash-invite-code-admin',
		title: '邀请码管理',
		pagePath: '/device/admin/cash-invite-code-admin.html',
		showInNav: true
	},
	{
		id: 'cash-feeding-eligibility-admin',
		title: '喂养资格门槛',
		pagePath: '/device/admin/cash-feeding-eligibility-admin.html',
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
	}
];
