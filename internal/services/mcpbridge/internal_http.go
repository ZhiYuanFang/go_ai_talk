package mcpbridge

// internal_http.go：xiaozhi-mcp-service 内部 HTTP（Upsert/Remove/health）。
//
// 仅内网可达；鉴权 DEVICE_GATEWAY_INTERNAL_SECRET。
// 响应形态与 GoFrame MiddlewareHandlerResponse 对齐：{code,message,data}。

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"hello/internal/platform/httpmeta"

	"github.com/gogf/gf/v2/os/glog"
)

// InternalHTTP 承载 Manager 的内部控制面。
type InternalHTTP struct {
	Manager *Manager
}

type jsonEnv struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Handler 返回 mux；供 ListenAndServe 使用。
func (h *InternalHTTP) Handler() http.Handler {
	mux := http.NewServeMux()
	// health 供编排探针，不要求内部密钥；Upsert/Remove 必须鉴权。
	mux.HandleFunc("/xiaozhi-mcp/internal/api/health", h.handleHealth)
	mux.HandleFunc("/xiaozhi-mcp/internal/api/bindings/upsert", h.withSecret(h.handleUpsert))
	mux.HandleFunc("/xiaozhi-mcp/internal/api/bindings/remove", h.withSecret(h.handleRemove))
	return mux
}

func (h *InternalHTTP) withSecret(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		secret := strings.TrimSpace(r.Header.Get(httpmeta.HeaderDeviceGatewayInternalSecret))
		if secret == "" {
			secret = strings.TrimSpace(r.Header.Get(httpmeta.HeaderGatewayInternalSecretLegacy))
		}
		if !httpmeta.ValidateInternalSecret(secret) {
			writeJSON(w, http.StatusUnauthorized, jsonEnv{Code: 401, Message: "unauthorized"})
			return
		}
		next(w, r)
	}
}

func (h *InternalHTTP) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, jsonEnv{Code: 405, Message: "method not allowed"})
		return
	}
	n := 0
	if h.Manager != nil {
		n = h.Manager.ActiveCount()
	}
	writeJSON(w, http.StatusOK, jsonEnv{Code: 0, Message: "OK", Data: map[string]interface{}{
		"activeBridges": n,
	}})
}

func (h *InternalHTTP) handleUpsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, jsonEnv{Code: 405, Message: "method not allowed"})
		return
	}
	var body struct {
		Id       int64  `json:"id"`
		McpToken string `json:"mcpToken"`
		DeviceNo string `json:"deviceNo"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, jsonEnv{Code: 400, Message: "invalid json"})
		return
	}
	if h.Manager == nil {
		writeJSON(w, http.StatusOK, jsonEnv{Code: 50, Message: "manager not ready"})
		return
	}
	h.Manager.Upsert(r.Context(), BindingSpec{
		Id:       body.Id,
		McpToken: body.McpToken,
		DeviceNo: body.DeviceNo,
	})
	writeJSON(w, http.StatusOK, jsonEnv{Code: 0, Message: "OK"})
}

func (h *InternalHTTP) handleRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, jsonEnv{Code: 405, Message: "method not allowed"})
		return
	}
	var body struct {
		Id       int64  `json:"id"`
		McpToken string `json:"mcpToken"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, jsonEnv{Code: 400, Message: "invalid json"})
		return
	}
	if h.Manager == nil {
		writeJSON(w, http.StatusOK, jsonEnv{Code: 50, Message: "manager not ready"})
		return
	}
	h.Manager.Remove(r.Context(), body.Id, body.McpToken)
	writeJSON(w, http.StatusOK, jsonEnv{Code: 0, Message: "OK"})
}

func readJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dst)
}

func writeJSON(w http.ResponseWriter, status int, env jsonEnv) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}

// RunInternalHTTP 在 addr 上阻塞监听；ctx 取消时 Shutdown。
func RunInternalHTTP(ctx context.Context, addr string, h *InternalHTTP) error {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		addr = ":9809"
	}
	srv := &http.Server{
		Addr:              addr,
		Handler:           h.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		glog.Infof(ctx, "[xiaozhi-mcp] internal HTTP listen %s", addr)
		errCh <- srv.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
