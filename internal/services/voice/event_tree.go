package voice

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/os/glog"
)

// eventTreeIndex 由扁平事件列表构建的层级索引，供匹配与追问使用。
type eventTreeIndex struct {
	childrenByParent map[int64][]entity.Event
	depthByID        map[int64]int
}

func buildEventTreeIndex(events []entity.Event) eventTreeIndex {
	childrenByParent := make(map[int64][]entity.Event)
	byID := make(map[int64]entity.Event, len(events))
	for _, e := range events {
		byID[e.Id] = e
		pid := e.ParentId
		if pid < 0 {
			pid = 0
		}
		childrenByParent[pid] = append(childrenByParent[pid], e)
	}
	depthByID := make(map[int64]int, len(events))
	var depth func(id int64) int
	depth = func(id int64) int {
		if d, ok := depthByID[id]; ok {
			return d
		}
		e, ok := byID[id]
		if !ok || e.ParentId <= 0 {
			depthByID[id] = 0
			return 0
		}
		d := depth(e.ParentId) + 1
		depthByID[id] = d
		return d
	}
	for _, e := range events {
		depth(e.Id)
	}
	return eventTreeIndex{
		childrenByParent: childrenByParent,
		depthByID:        depthByID,
	}
}

func (idx eventTreeIndex) hasChildren(id int64) bool {
	return len(idx.childrenByParent[id]) > 0
}

func (idx eventTreeIndex) sortForMatch(candidates []entity.Event) {
	sort.Slice(candidates, func(i, j int) bool {
		di := idx.depthByID[candidates[i].Id]
		dj := idx.depthByID[candidates[j].Id]
		if di != dj {
			return di > dj
		}
		return len([]rune(candidates[i].Name)) > len([]rune(candidates[j].Name))
	})
}

// hasSignificantOverlap 判断两个文本是否有显著的交集（至少两个连续字符）。
func hasSignificantOverlap(text, keyword string) bool {
	textRunes := []rune(text)
	keywordRunes := []rune(keyword)
	if len(textRunes) < 2 || len(keywordRunes) < 2 {
		return false
	}
	for i := 0; i < len(textRunes)-1; i++ {
		for j := 0; j < len(keywordRunes)-1; j++ {
			if textRunes[i] == keywordRunes[j] && textRunes[i+1] == keywordRunes[j+1] {
				return true
			}
		}
	}
	return false
}

// extractEventFromCandidates 在候选集合中按深度与名称长度优先匹配事件。
func extractEventFromCandidates(ctx context.Context, normalizedTranscript string, candidates []entity.Event, idx eventTreeIndex) (entity.Event, string, bool) {
	if len(candidates) == 0 {
		return entity.Event{}, "", false
	}
	sorted := append([]entity.Event(nil), candidates...)
	idx.sortForMatch(sorted)
	for _, event := range sorted {
		if hasSignificantOverlap(normalizedTranscript, event.Name) {
			glog.Infof(ctx, "命中事件名: %s", event.Name)
			return event, event.Name, true
		}
		if event.ExtraNames != "" {
			for _, extraName := range strings.Split(event.ExtraNames, ",") {
				extraName = strings.TrimSpace(extraName)
				if extraName != "" && strings.Contains(normalizedTranscript, extraName) {
					glog.Infof(ctx, "命中额外名称: %s", extraName)
					return event, extraName, true
				}
			}
		}
	}
	return entity.Event{}, "", false
}

// formatChildDisambiguationReply 根据直接子节点名称生成追问话术。
func formatChildDisambiguationReply(parentName string, children []entity.Event) string {
	names := make([]string, 0, len(children))
	for _, c := range children {
		if n := strings.TrimSpace(c.Name); n != "" {
			names = append(names, n)
		}
	}
	if len(names) == 0 {
		return fmt.Sprintf("请具体说明%s的类型", strings.TrimSpace(parentName))
	}
	if len(names) == 1 {
		return fmt.Sprintf("%s是%s吗？", strings.TrimSpace(parentName), names[0])
	}
	if len(names) == 2 {
		return fmt.Sprintf("%s是%s还是%s？", strings.TrimSpace(parentName), names[0], names[1])
	}
	last := names[len(names)-1]
	prefix := strings.Join(names[:len(names)-1], "、")
	return fmt.Sprintf("%s是%s还是%s？", strings.TrimSpace(parentName), prefix, last)
}
