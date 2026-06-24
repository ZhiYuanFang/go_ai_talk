package ucg

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"

	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/util/guid"
)

// feedCursor Feed 分页 opaque cursor（冻结 session 与坐标上下文）。
type feedCursor struct {
	SessionID      string  `json:"sid"`
	Lat            float64 `json:"lat,omitempty"`
	Lng            float64 `json:"lng,omitempty"`
	HasViewer      bool    `json:"hv,omitempty"`
	RadiusIdx      int     `json:"ri"`
	GeoOffset      int     `json:"go"`
	ZsetOffset     int     `json:"zo"`
	LastFinalScore float64 `json:"lfs,omitempty"`
	LastPostID     uint64  `json:"lpid,omitempty"`
}

func newFeedSessionID() string {
	return guid.S()
}

func encodeFeedCursor(c feedCursor) string {
	raw, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeFeedCursor(raw string) (feedCursor, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return feedCursor{}, false
	}
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return feedCursor{}, false
	}
	var c feedCursor
	if err = json.Unmarshal(data, &c); err != nil {
		return feedCursor{}, false
	}
	if strings.TrimSpace(c.SessionID) == "" {
		return feedCursor{}, false
	}
	return c, true
}

func loadSessionSeen(ctx context.Context, sessionID string) (map[uint64]struct{}, error) {
	seen := make(map[uint64]struct{})
	if strings.TrimSpace(sessionID) == "" {
		return seen, nil
	}
	members, err := ucgCache.SetMembers(ctx, cachekit.UCGFeedSessionKey(sessionID))
	if err != nil {
		return nil, err
	}
	for _, m := range members {
		id, err := strconv.ParseUint(strings.TrimSpace(m), 10, 64)
		if err != nil || id == 0 {
			continue
		}
		seen[id] = struct{}{}
	}
	return seen, nil
}

func markSessionSeen(ctx context.Context, cfg FeedConfig, sessionID string, postIDs []uint64) error {
	if strings.TrimSpace(sessionID) == "" || len(postIDs) == 0 {
		return nil
	}
	members := make([]string, 0, len(postIDs))
	for _, id := range postIDs {
		if id == 0 {
			continue
		}
		members = append(members, strconv.FormatUint(id, 10))
	}
	if len(members) == 0 {
		return nil
	}
	return ucgCache.SetAddWithExpire(ctx, cachekit.UCGFeedSessionKey(sessionID), cfg.sessionTTL(), members...)
}

func cursorViewer(c feedCursor) (ViewerCoords, bool) {
	if !c.HasViewer {
		return ViewerCoords{}, false
	}
	return ValidViewerCoords(&c.Lat, &c.Lng)
}

func afterCursor(lastFinalScore float64, lastPostID uint64, finalScore float64, postID uint64) bool {
	if lastPostID == 0 && lastFinalScore == 0 {
		return true
	}
	if finalScore < lastFinalScore {
		return true
	}
	if finalScore > lastFinalScore {
		return false
	}
	return postID < lastPostID
}
