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
