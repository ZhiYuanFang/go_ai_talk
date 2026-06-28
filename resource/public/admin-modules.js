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
		id: 'app-status-admin',
		title: 'App 维护通知',
		externalUrl: 'https://status.pangbao.cuplay.top/admin',
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
