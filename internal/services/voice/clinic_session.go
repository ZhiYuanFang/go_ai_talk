package voice

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

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

func clinicSessionRedisKey(wxID int64) string {
	return fmt.Sprintf("%s%d", clinicSessionKeyPrefix, wxID)
}

func loadClinicSession(ctx context.Context, wxID int64) (clinicSession, bool, error) {
	key := clinicSessionRedisKey(wxID)
	v, err := g.Redis().Do(ctx, "GET", key)
	if err != nil {
		return clinicSession{}, false, err
	}
	if v.IsNil() || v.IsEmpty() {
		return clinicSession{}, false, nil
	}
	var sess clinicSession
	if err := json.Unmarshal(v.Bytes(), &sess); err != nil {
		return clinicSession{}, false, err
	}
	return sess, true, nil
}

// saveClinicSession 写回会话；已有 session 时 MUST NOT 刷新 EXPIRE（固定 TTL 非 sliding）。
func saveClinicSession(ctx context.Context, wxID int64, sess clinicSession, ttlSeconds int, isNew bool) error {
	key := clinicSessionRedisKey(wxID)
	raw, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	if isNew {
		_, err = g.Redis().Do(ctx, "SET", key, string(raw), "EX", ttlSeconds)
		return err
	}
	// 非首问更新：保留原 TTL，不 sliding 续期（12h 固定窗口）。
	ttlVal, _ := g.Redis().Do(ctx, "TTL", key)
	remain := ttlVal.Int()
	if remain > 0 {
		_, err = g.Redis().Do(ctx, "SET", key, string(raw), "EX", remain)
	} else {
		_, err = g.Redis().Do(ctx, "SET", key, string(raw))
	}
	return err
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
