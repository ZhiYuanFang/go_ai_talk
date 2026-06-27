package simuser

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"hello/internal/services/aimodel"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/glog"
)

// simCommentExcludedMediaTypeVideo 与 ucg_post.media_type 视频值一致（ucg MediaTypeVideo=2）。
const simCommentExcludedMediaTypeVideo = 2

var ephemeralMu sync.Mutex
var ephemeralActive = map[string]struct{}{}

// RunRegisterTask T1：注册模拟用户。
func RunRegisterTask(ctx context.Context, password string) {
	cfg, err := GetConfig(ctx)
	if err != nil || (!cfg.Enabled && !isManualRun(ctx)) {
		return
	}
	total, err := countSimUsers(ctx)
	if err != nil {
		RecordTaskRun(ctx, "register", false, err.Error())
		return
	}
	if total >= cfg.MaxSimUsers {
		RecordTaskRun(ctx, "register", true, "已达上限")
		return
	}
	account, err := NextAccountName(ctx)
	if err != nil {
		RecordTaskRun(ctx, "register", false, err.Error())
		return
	}
	if password == "" {
		password = "123456"
	}
	wxID, err := simRegister(ctx, account, password)
	if err != nil {
		RecordTaskRun(ctx, "register", false, err.Error())
		return
	}
	sess, err := usernameLogin(ctx, account, password)
	if err != nil {
		RecordTaskRun(ctx, "register", false, err.Error())
		return
	}
	_ = wxID
	sys, user, err := LoadRenderedPrompt(ctx, "register_nickname", nil)
	if err != nil {
		RecordTaskRun(ctx, "register", false, err.Error())
		return
	}
	msgs := []aimodel.Message{{Role: "user", Content: user}}
	if sys != "" {
		msgs = []aimodel.Message{{Role: "system", Content: sys}, {Role: "user", Content: user}}
	}
	resp, err := aimodel.Invoke(ctx, aimodel.LaneSimText, simChatRequest(simTempRegisterNickname, 64, msgs))
	if err != nil {
		RecordTaskRun(ctx, "register", false, err.Error())
		return
	}
	nickname := strings.TrimSpace(resp.Content)
	_, userAv, _ := LoadRenderedPrompt(ctx, "register_avatar", nil)
	imgRes, err := aimodel.GenerateImage(ctx, aimodel.LaneSimImageGen, userAv)
	if err != nil {
		RecordTaskRun(ctx, "register", false, err.Error())
		return
	}
	avatarKey, err := uploadImageFromURL(ctx, sess.AccessToken, imgRes.URL)
	if err != nil {
		RecordTaskRun(ctx, "register", false, err.Error())
		return
	}
	if err = appPut(ctx, sess.AccessToken, "/ucg/app/api/profile/me", g.Map{
		"nickname": nickname, "avatarKey": avatarKey,
	}, nil); err != nil {
		RecordTaskRun(ctx, "register", false, err.Error())
		return
	}
	RecordTaskRun(ctx, "register", true, "")
	glog.Infof(ctx, "[simuser] registered account=%s wxId=%d", account, sess.WxID)
}

// RunCommentTask T2：评论广场帖（ucg internal random 模式全库抽样，略偏新帖）。
func RunCommentTask(ctx context.Context, password string) {
	if !taskEnabled(ctx) {
		return
	}
	sess, _, err := randomSimSession(ctx, password)
	if err != nil {
		RecordTaskRun(ctx, "comment", false, err.Error())
		return
	}
	var sample struct {
		List []struct {
			PostId         uint64 `json:"postId"`
			Content        string `json:"content"`
			MediaType      int    `json:"mediaType"`
			CoverObjectKey string `json:"coverObjectKey"`
			CoverCdnUrl    string `json:"coverCdnUrl"`
		} `json:"list"`
	}
	if err = ucgInternalPost(ctx, "/ucg/internal/api/posts/sample", g.Map{
		"mode":              "random",
		"excludeMediaTypes": []int{simCommentExcludedMediaTypeVideo},
	}, &sample); err != nil || len(sample.List) == 0 {
		RecordTaskRun(ctx, "comment", false, "无已发布帖")
		return
	}
	post := sample.List[0]
	if post.MediaType == simCommentExcludedMediaTypeVideo {
		glog.Warningf(ctx, "[simuser] comment postId=%d 仍为视频帖，跳过", post.PostId)
		RecordTaskRun(ctx, "comment", false, "跳过视频帖")
		return
	}
	_, user, _ := LoadRenderedPrompt(ctx, "comment", map[string]string{"post_content": post.Content})
	coverCdnURL := strings.TrimSpace(post.CoverCdnUrl)
	if post.MediaType != 0 && coverCdnURL == "" {
		glog.Warningf(ctx, "[simuser] comment postId=%d mediaType=%d 无 coverCdnUrl，降级纯文本 simVision", post.PostId, post.MediaType)
	}
	resp, err := aimodel.Invoke(ctx, aimodel.LaneSimVision, simChatRequest(simTempComment, 256, []aimodel.Message{{
		Role:    "user",
		Content: commentVisionUserContent(coverCdnURL, user),
	}}))
	if err != nil {
		RecordTaskRun(ctx, "comment", false, err.Error())
		return
	}
	path := fmt.Sprintf("/ucg/app/api/posts/%d/comments", post.PostId)
	// 评论帖子
	if err = appPost(ctx, sess.AccessToken, path, g.Map{"content": strings.TrimSpace(resp.Content)}, nil); err != nil {
		RecordTaskRun(ctx, "comment", false, err.Error())
		return
	}
	RecordTaskRun(ctx, "comment", true, "")
}

// RunPostImageTask T3：图文动态。
func RunPostImageTask(ctx context.Context, password string) {
	if !taskEnabled(ctx) {
		return
	}
	sess, _, err := randomSimSession(ctx, password)
	if err != nil {
		RecordTaskRun(ctx, "post_image", false, err.Error())
		return
	}
	topic := randomTopic()
	_, user, _ := LoadRenderedPrompt(ctx, "post_image_text", map[string]string{"topic": topic})
	textResp, err := aimodel.Invoke(ctx, aimodel.LaneSimText, simChatRequest(simTempPostImageText, 512, []aimodel.Message{{Role: "user", Content: user}}))
	if err != nil {
		RecordTaskRun(ctx, "post_image", false, err.Error())
		return
	}
	imgRes, err := aimodel.GenerateImage(ctx, aimodel.LaneSimImageGen, textResp.Content)
	if err != nil {
		RecordTaskRun(ctx, "post_image", false, err.Error())
		return
	}
	key, err := uploadImageFromURL(ctx, sess.AccessToken, imgRes.URL)
	if err != nil {
		RecordTaskRun(ctx, "post_image", false, err.Error())
		return
	}
	if err = appPost(ctx, sess.AccessToken, "/ucg/app/api/posts", g.Map{
		"content": textResp.Content, "mediaType": 1, "submit": true,
		"media": []g.Map{{"objectKey": key, "mediaKind": 1, "sortOrder": 0}},
	}, nil); err != nil {
		RecordTaskRun(ctx, "post_image", false, err.Error())
		return
	}
	RecordTaskRun(ctx, "post_image", true, "")
}

// RunPostVideoSubmitTask T4：提交视频生成。
func RunPostVideoSubmitTask(ctx context.Context, password string) {
	if !taskEnabled(ctx) {
		return
	}
	sess, _, err := randomSimSession(ctx, password)
	if err != nil {
		RecordTaskRun(ctx, "post_video_submit", false, err.Error())
		return
	}
	ok, err := HasPendingVideoJob(ctx, sess.WxID)
	if err != nil || ok {
		RecordTaskRun(ctx, "post_video_submit", true, "已有 pending job")
		return
	}
	topic := randomTopic()
	_, user, _ := LoadRenderedPrompt(ctx, "post_video_text", map[string]string{"topic": topic})
	textResp, err := aimodel.Invoke(ctx, aimodel.LaneSimText, simChatRequest(simTempPostVideoText, 256, []aimodel.Message{{Role: "user", Content: user}}))
	if err != nil {
		RecordTaskRun(ctx, "post_video_submit", false, err.Error())
		return
	}
	taskID, err := aimodel.SubmitVideoGeneration(ctx, aimodel.LaneSimVideoGen, textResp.Content)
	if err != nil {
		RecordTaskRun(ctx, "post_video_submit", false, err.Error())
		return
	}
	if err = InsertVideoJob(ctx, sess.WxID, textResp.Content, taskID); err != nil {
		RecordTaskRun(ctx, "post_video_submit", false, err.Error())
		return
	}
	RecordTaskRun(ctx, "post_video_submit", true, "")
}

// RunVideoPollTask P1：轮询视频任务；有 pending job 时返回 true 以缩短下一间隔。
func RunVideoPollTask(ctx context.Context, password string) bool {
	jobs, err := ListPendingVideoJobs(ctx)
	if err != nil || len(jobs) == 0 {
		return false
	}
	for _, job := range jobs {
		res, pErr := aimodel.PollVideoGeneration(ctx, job.TaskId)
		if pErr != nil {
			continue
		}
		switch res.Status {
		case aimodel.VideoStatusFailed:
			_ = UpdateVideoJobStatus(ctx, job.Id, "skipped")
		case aimodel.VideoStatusSuccess:
			account := accountForWx(ctx, job.WxId)
			if account == "" {
				_ = UpdateVideoJobStatus(ctx, job.Id, "skipped")
				continue
			}
			if password == "" {
				password = "123456"
			}
			sess, lErr := usernameLogin(ctx, account, password)
			if lErr != nil {
				_ = UpdateVideoJobStatus(ctx, job.Id, "skipped")
				continue
			}
			vKey, uErr := uploadVideoFromURL(ctx, sess.AccessToken, res.VideoURL)
			if uErr != nil {
				_ = UpdateVideoJobStatus(ctx, job.Id, "skipped")
				continue
			}
			if err = appPost(ctx, sess.AccessToken, "/ucg/app/api/posts", g.Map{
				"content": job.Content, "mediaType": 2, "submit": true,
				"media": []g.Map{{"objectKey": vKey, "mediaKind": 2, "sortOrder": 0}},
			}, nil); err != nil {
				_ = UpdateVideoJobStatus(ctx, job.Id, "skipped")
				continue
			}
			_ = UpdateVideoJobStatus(ctx, job.Id, "done")
		default:
			_ = UpdateVideoJobStatus(ctx, job.Id, "processing")
		}
	}
	return true
}

// RunChatScanTask T5：聊天巡检，真人未读触发 E1。
func RunChatScanTask(ctx context.Context, password string, flags RuntimeFlags) {
	if !taskEnabled(ctx) {
		return
	}
	sess, account, err := randomSimSession(ctx, password)
	if err != nil {
		RecordTaskRun(ctx, "chat_scan", false, err.Error())
		return
	}
	_ = account
	var conv struct {
		List []struct {
			Id          uint64 `json:"id"`
			PeerWxId    uint64 `json:"peerWxId"`
			UnreadCount int    `json:"unreadCount"`
		} `json:"list"`
	}
	if err = appGet(ctx, sess.AccessToken, "/ucg/app/api/conversations?page=1&pageSize=50", &conv); err != nil {
		RecordTaskRun(ctx, "chat_scan", false, err.Error())
		return
	}
	for _, c := range conv.List {
		if c.UnreadCount <= 0 {
			continue
		}
		if isPeerSimulated(ctx, int64(c.PeerWxId)) {
			continue
		}
		spawnEphemeralChat(ctx, sess, password, c.Id, int64(c.PeerWxId), flags)
	}
	RecordTaskRun(ctx, "chat_scan", true, "")
}

func spawnEphemeralChat(ctx context.Context, sess loginSession, password string, convID uint64, peerWxID int64, flags RuntimeFlags) {
	key := fmt.Sprintf("%d:%d", sess.WxID, peerWxID)
	ephemeralMu.Lock()
	if _, ok := ephemeralActive[key]; ok {
		ephemeralMu.Unlock()
		return
	}
	ephemeralActive[key] = struct{}{}
	ephemeralMu.Unlock()
	loop := flags.EphemeralChatLoop
	if loop <= 0 {
		loop = 5 * time.Minute
	}
	window := flags.EphemeralChatWindow
	if window <= 0 {
		window = 15 * time.Minute
	}
	go func() {
		defer func() {
			ephemeralMu.Lock()
			delete(ephemeralActive, key)
			ephemeralMu.Unlock()
		}()
		deadline := time.Now().Add(window)
		for time.Now().Before(deadline) {
			time.Sleep(loop)
			var conv struct {
				List []struct {
					Id          uint64 `json:"id"`
					UnreadCount int    `json:"unreadCount"`
				} `json:"list"`
			}
			if err := appGet(ctx, sess.AccessToken, "/ucg/app/api/conversations?page=1&pageSize=50", &conv); err != nil {
				continue
			}
			unread := 0
			for _, c := range conv.List {
				if c.Id == convID {
					unread = c.UnreadCount
					break
				}
			}
			if unread <= 0 {
				continue
			}
			var msgs struct {
				List []struct {
					Content    string `json:"content"`
					SenderWxId int64  `json:"senderWxId"`
				} `json:"list"`
			}
			path := fmt.Sprintf("/ucg/app/api/conversations/%d/messages?page=1&pageSize=20", convID)
			if err := appGet(ctx, sess.AccessToken, path, &msgs); err != nil {
				continue
			}
			history := buildChatHistory(msgs.List)
			_, user, _ := LoadRenderedPrompt(ctx, "chat_reply", map[string]string{"chat_history": history})
			resp, err := aimodel.Invoke(ctx, aimodel.LaneSimText, simChatRequest(simTempChatReply, 512, []aimodel.Message{{Role: "user", Content: user}}))
			if err != nil {
				continue
			}
			_ = sendInternalChat(ctx, sess.WxID, convID, strings.TrimSpace(resp.Content))
		}
	}()
}

// RunFollowTask T6：sim 关注 sim。
func RunFollowTask(ctx context.Context, password string) {
	if !taskEnabled(ctx) {
		return
	}
	follower, target, err := pickTwoDistinctSimWx(ctx)
	if err != nil {
		RecordTaskRun(ctx, "follow", false, err.Error())
		return
	}
	sess, err := usernameLogin(ctx, follower.Account, password)
	if err != nil {
		RecordTaskRun(ctx, "follow", false, err.Error())
		return
	}
	path := fmt.Sprintf("/ucg/app/api/follow/%d", target.WxId)
	if err = appPost(ctx, sess.AccessToken, path, g.Map{}, nil); err != nil {
		RecordTaskRun(ctx, "follow", false, err.Error())
		return
	}
	RecordTaskRun(ctx, "follow", true, "")
}

func taskEnabled(ctx context.Context) bool {
	if isManualRun(ctx) {
		return true
	}
	cfg, err := GetConfig(ctx)
	return err == nil && cfg.Enabled
}

func buildChatHistory(msgs []struct {
	Content    string `json:"content"`
	SenderWxId int64  `json:"senderWxId"`
}) string {
	var b strings.Builder
	for _, m := range msgs {
		b.WriteString(fmt.Sprintf("%d: %s\n", m.SenderWxId, m.Content))
	}
	return b.String()
}

func randomTopic() string {
	topics := []string{"宝宝辅食", "亲子阅读", "产后恢复", "婴儿睡眠", "早教游戏"}
	return topics[rand.Intn(len(topics))]
}

func isPeerSimulated(ctx context.Context, wxID int64) bool {
	base := deviceBase(ctx)
	resp, err := gclient.New().
		SetHeader("X-Device-Gateway-Internal-Secret", internalSecret(ctx)).
		ContentJson().Post(ctx, base+"/device/internal/api/ucg/wx/batch", g.Map{"wxIds": []int64{wxID}})
	if err != nil {
		return false
	}
	var data struct {
		List []struct {
			IsSimulated bool `json:"isSimulated"`
		} `json:"list"`
	}
	if parseEnvelope(resp.ReadAllString(), &data) != nil || len(data.List) == 0 {
		return false
	}
	return data.List[0].IsSimulated
}

func accountForWx(ctx context.Context, wxID int64) string {
	base := deviceBase(ctx)
	resp, err := gclient.New().
		SetHeader("X-Device-Gateway-Internal-Secret", internalSecret(ctx)).
		Get(ctx, base+"/device/internal/api/sim/wx/list?page=1&pageSize=200")
	if err != nil {
		return ""
	}
	var data struct {
		List []struct {
			WxId    int64  `json:"wxId"`
			Account string `json:"account"`
		} `json:"list"`
	}
	if parseEnvelope(resp.ReadAllString(), &data) != nil {
		return ""
	}
	for _, item := range data.List {
		if item.WxId == wxID {
			return item.Account
		}
	}
	return ""
}
