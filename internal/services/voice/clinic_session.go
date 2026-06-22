package voice

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"hello/internal/platform/cachekit"
)

var clinicCache = cachekit.Default()

// clinicTurn 会话内单轮 Q&A（供多轮上下文，不含 thinking 全文）。
type clinicTurn struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// clinicSession Redis voice:clinic:session:{wxId} 结构；首问创建，12h 固定 TTL，后续提问不续期。
type clinicSession struct {
	DeviceNo        string       `json:"deviceNo"`
	FirstQuestionAt int64        `json:"firstQuestionAt"`
	Turns           []clinicTurn `json:"turns"`
}

func loadClinicSession(ctx context.Context, wxID int64) (clinicSession, bool, error) {
	key := cachekit.VoiceClinicSessionKey(wxID)
	v, ok, err := clinicCache.Get(ctx, key)
	if err != nil {
		return clinicSession{}, false, err
	}
	if !ok || v == "" {
		return clinicSession{}, false, nil
	}
	var sess clinicSession
	if err := json.Unmarshal([]byte(v), &sess); err != nil {
		return clinicSession{}, false, err
	}
	return sess, true, nil
}

// saveClinicSession 写回会话；已有 session 时 MUST NOT 刷新 EXPIRE（固定 TTL 非 sliding）。
func saveClinicSession(ctx context.Context, wxID int64, sess clinicSession, ttlSeconds int, isNew bool) error {
	key := cachekit.VoiceClinicSessionKey(wxID)
	raw, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	if isNew {
		return clinicCache.SetEX(ctx, key, string(raw), time.Duration(ttlSeconds)*time.Second)
	}
	// 非首问更新：保留原 TTL，不 sliding 续期（12h 固定窗口）。
	remain, err := clinicCache.TTL(ctx, key)
	if err != nil {
		return err
	}
	if remain > 0 {
		return clinicCache.SetEX(ctx, key, string(raw), time.Duration(remain)*time.Second)
	}
	return clinicCache.Set(ctx, key, string(raw))
}

func appendClinicTurn(ctx context.Context, wxID int64, cfg AIClinicConfig, deviceNo, question, answer string) error {
	sess, exists, err := loadClinicSession(ctx, wxID)
	if err != nil {
		return err
	}
	isNew := !exists
	if isNew {
		sess = clinicSession{
			DeviceNo:        strings.TrimSpace(deviceNo),
			FirstQuestionAt: time.Now().Unix(),
			Turns:           nil,
		}
	}
	sess.Turns = append(sess.Turns, clinicTurn{Question: question, Answer: answer})
	ttl := cfg.SessionTTLSeconds
	if ttl <= 0 {
		ttl = 12 * 3600
	}
	return saveClinicSession(ctx, wxID, sess, ttl, isNew)
}

func clinicSessionMessages(sess clinicSession) []map[string]string {
	out := make([]map[string]string, 0, len(sess.Turns)*2)
	for _, t := range sess.Turns {
		if q := strings.TrimSpace(t.Question); q != "" {
			out = append(out, map[string]string{"role": "user", "content": q})
		}
		if a := strings.TrimSpace(t.Answer); a != "" {
			out = append(out, map[string]string{"role": "assistant", "content": a})
		}
	}
	return out
}

// SessionSyncTurn auth_ok 后 session_sync 帧中的单轮 Q&A；不含 thinking（Redis 未持久化 reasoning）。
type SessionSyncTurn struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// SessionSyncPayload auth_ok 成功后立即下发的会话同步帧。
type SessionSyncPayload struct {
	Type      string            `json:"type"`
	Turns     []SessionSyncTurn `json:"turns"`
	ExpiresAt int64             `json:"expiresAt"`
}

// BuildSessionSync 读 Redis voice:clinic:session:{wxId}，过滤 question/answer 均非空的已完成轮次。
func BuildSessionSync(ctx context.Context, wxID int64, sessionTTLSeconds int) (SessionSyncPayload, error) {
	payload := SessionSyncPayload{
		Type:      "session_sync",
		Turns:     []SessionSyncTurn{},
		ExpiresAt: 0,
	}
	sess, exists, err := loadClinicSession(ctx, wxID)
	if err != nil {
		return payload, err
	}
	if !exists || sess.FirstQuestionAt <= 0 {
		return payload, nil
	}
	ttl := sessionTTLSeconds
	if ttl <= 0 {
		ttl = 12 * 3600
	}
	payload.ExpiresAt = sess.FirstQuestionAt + int64(ttl)
	for _, t := range sess.Turns {
		q := strings.TrimSpace(t.Question)
		a := strings.TrimSpace(t.Answer)
		if q == "" || a == "" {
			continue
		}
		payload.Turns = append(payload.Turns, SessionSyncTurn{Question: q, Answer: a})
	}
	return payload, nil
}
