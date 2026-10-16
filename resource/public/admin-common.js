// 运维 Hub 共享：Admin JWT 登录与 Bearer fetch。
window.AdminCommon = (function () {
	const TOKEN_KEY = 'gatewayAdminToken';

	function getToken() {
		return sessionStorage.getItem(TOKEN_KEY) || '';
	}

	function setToken(token) {
		if (token) {
			sessionStorage.setItem(TOKEN_KEY, token);
		} else {
			sessionStorage.removeItem(TOKEN_KEY);
		}
	}

	function clearToken() {
		sessionStorage.removeItem(TOKEN_KEY);
	}

	function authHeaders(extra) {
		const headers = Object.assign({'Content-Type': 'application/json'}, extra || {});
		const token = getToken();
		if (token) {
			headers.Authorization = 'Bearer ' + token;
		}
		return headers;
	}

	async function login(username, password) {
		const res = await window.fetch('/device/admin/api/login', {
			method: 'POST',
			headers: {'Content-Type': 'application/json'},
			body: JSON.stringify({username: username, password: password})
		});
		const data = await res.json().catch(function () { return {}; });
		if (!res.ok || data.code !== 0) {
			throw new Error(data.message || ('HTTP ' + res.status));
		}
		setToken(data.data && data.data.accessToken);
		return data.data;
	}

	async function adminFetch(path, options) {
		options = options || {};
		const headers = authHeaders(options.headers);
		const res = await window.fetch(path, Object.assign({}, options, {headers: headers}));
		const data = await res.json().catch(function () { return {}; });
		if (!res.ok || data.code !== 0) {
			throw new Error(data.message || ('HTTP ' + res.status));
		}
		if (!data.data){
			return data
		}
		return data.data;
	}

	async function adminFetchForm(path, formData) {
		const headers = {};
		const token = getToken();
		if (token) {
			headers.Authorization = 'Bearer ' + token;
		}
		const res = await window.fetch(path, {method: 'POST', headers: headers, body: formData});
		const data = await res.json().catch(function () { return {}; });
		if (!res.ok || data.code !== 0) {
			throw new Error(data.message || ('HTTP ' + res.status));
		}
		if (!data.data){
			return data
		}
		return data.data;
	}

	function requireAdmin() {
		if (!getToken()) {
			window.location.href = '/device/admin';
			return false;
		}
		return true;
	}

	function hubUrl() {
		return '/device/admin';
	}

	return {
		getToken: getToken,
		setToken: setToken,
		clearToken: clearToken,
		login: login,
		adminFetch: adminFetch,
		adminFetchForm: adminFetchForm,
		requireAdmin: requireAdmin,
		hubUrl: hubUrl
	};
})();

// 飞轮页「复制提示词」：成功走剪贴板，失败则选中只读框便于手动复制。
document.addEventListener('click', function (ev) {
	var btn = ev.target && ev.target.closest ? ev.target.closest('.seed-copy-btn') : null;
	if (!btn) return;
	var id = btn.getAttribute('data-seed-target') || 'seedPrompt';
	var box = document.getElementById(id);
	if (!box) return;
	var text = box.value || '';
	function mark(ok) {
		var prev = btn.textContent;
		btn.textContent = ok ? '已复制' : '请手动复制';
		setTimeout(function () { btn.textContent = prev; }, 1200);
	}
	if (navigator.clipboard && navigator.clipboard.writeText) {
		navigator.clipboard.writeText(text).then(function () { mark(true); }).catch(function () {
			box.focus();
			box.select();
			mark(false);
		});
		return;
	}
	box.focus();
	box.select();
	mark(false);
});
