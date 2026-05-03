package workeroutbox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"hello/internal/services/contracts"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

const workerOutboxRowPending = "pending"

// RegisterOutboxEnqueueHandler 注册 domain_outbox 入队 HTTP 处理器（与 worker 健康检查共用 mux）。
func RegisterOutboxEnqueueHandler(mux *http.ServeMux) {
	mux.HandleFunc("/worker/internal/api/outbox/enqueue", handleOutboxEnqueue)
}

func handleOutboxEnqueue(w http.ResponseWriter, r *http.Request) {
	ctx := gctx.New()
	if r.Method != http.MethodPost {
		writeOutboxEnqueueResponse(w, http.StatusMethodNotAllowed, 1, "method not allowed", nil)
		return
	}
	var body struct {
		RoutingKey string                 `json:"routingKey"`
		Payload    map[string]interface{} `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeOutboxEnqueueResponse(w, http.StatusBadRequest, 1, "invalid json", nil)
		return
	}
	rkStr := strings.TrimSpace(body.RoutingKey)
	routingKey, ok := contracts.ParseRoutingKey(ctx, rkStr, "worker_outbox_enqueue")
	if !ok || !routingKey.IsValid() {
		writeOutboxEnqueueResponse(w, http.StatusBadRequest, 1, "invalid routingKey", nil)
		return
	}
	if body.Payload == nil {
		body.Payload = map[string]interface{}{}
	}
	eventID := fmtOutboxEventID(body.Payload)
	payloadJSON, err := json.Marshal(body.Payload)
	if err != nil {
		writeOutboxEnqueueResponse(w, http.StatusBadRequest, 1, err.Error(), nil)
		return
	}
	_, err = g.DB("outbox").Model("domain_outbox").Data(g.Map{
		"event_id":    eventID,
		"event_type":  routingKey.String(),
		"routing_key": routingKey.String(),
		"payload":     string(payloadJSON),
		"status":      workerOutboxRowPending,
		"attempts":    0,
		"last_error":  "",
	}).Insert()
	if err != nil {
		writeOutboxEnqueueResponse(w, http.StatusInternalServerError, 1, err.Error(), nil)
		return
	}
	writeOutboxEnqueueResponse(w, http.StatusOK, 0, "", nil)
}

func fmtOutboxEventID(payload map[string]interface{}) string {
	if v, ok := payload["event_id"]; ok {
		if s, ok := v.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				return s
			}
		}
	}
	return fmt.Sprintf("outbox-%d", time.Now().UnixNano())
}

func writeOutboxEnqueueResponse(w http.ResponseWriter, httpStatus, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(g.Map{
		"code":    code,
		"message": message,
		"data":    data,
	})
}
