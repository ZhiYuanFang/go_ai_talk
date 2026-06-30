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
	"github.com/gogf/gf/v2/os/glog"
)

// simCommentExcludedMediaTypeVideo 与 ucg_post.media_type 视频值一致（ucg MediaTypeVideo=2）。
const simCommentExcludedMediaTypeVideo = 2

var videoPollMu sync.Mutex
var videoPostInFlight bool

// IsVideoPostInFlight 是否有进行中的 T4 视频流水线（submit + poll + 发帖）。
func IsVideoPostInFlight() bool {
	videoPollMu.Lock()
	defer videoPollMu.Unlock()
	return videoPostInFlight
}

func setVideoPostInFlight(v bool) {
	videoPollMu.Lock()
	videoPostInFlight = v
	videoPollMu.Unlock()
}

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
	_ = password
	var account, regPassword string
	var wxID int64
	committed := false
	defer func() {
		if !committed && wxID > 0 {
			rollbackSimRegistration(ctx, wxID)
		}
	}()
	failRegister := func(msg string) {
		RecordTaskRun(ctx, "register", false, msg)
	}
	for attempt := 0; attempt < 8; attempt++ {
		account, regPassword, err = GenerateRandomSimCredentials()
		if err != nil {
			failRegister(err.Error())
			return
		}
		wxID, err = simRegister(ctx, account, regPassword)
		if err == nil {
			break
		}
		if attempt == 7 {
			failRegister(err.Error())
			return
		}
	}
	sess, err := usernameLogin(ctx, account, regPassword)
	if err != nil {
		failRegister(err.Error())
		return
	}
	sys, user, err := LoadRenderedPrompt(ctx, "register_nickname", nil)
	if err != nil {
		failRegister(err.Error())
		return
	}
	msgs := []aimodel.Message{{Role: "user", Content: user}}
	if sys != "" {
		msgs = []aimodel.Message{{Role: "system", Content: sys}, {Role: "user", Content: user}}
	}
	resp, err := aimodel.Invoke(ctx, aimodel.LaneSimText, simChatRequest(simTempRegisterNickname, 64, msgs))
	if err != nil {
		failRegister(err.Error())
		return
	}
	nickname := strings.TrimSpace(resp.Content)
	if nickname == "" {
		failRegister("昵称为空")
		return
	}
	_, userAv, _ := LoadRenderedPrompt(ctx, "register_avatar", nil)
	imgRes, err := aimodel.GenerateImage(ctx, aimodel.LaneSimImageGen, userAv)
	if err != nil {
		failRegister(err.Error())
		return
	}
	if strings.TrimSpace(imgRes.URL) == "" {
		failRegister("头像 URL 为空")
		return
	}
	avatarKey, err := uploadImageFromURL(ctx, sess.AccessToken, imgRes.URL)
	if err != nil {
		failRegister(err.Error())
		return
	}
	if strings.TrimSpace(avatarKey) == "" {
		failRegister("avatarKey 为空")
		return
	}
	if err = appPut(ctx, sess.AccessToken, "/ucg/app/api/profile/me", g.Map{
		"nickname": nickname, "avatarKey": avatarKey,
	}, nil); err != nil {
		failRegister(err.Error())
		return
	}
	if err = InsertWxCredential(ctx, wxID, account, regPassword); err != nil {
		failRegister(err.Error())
		return
	}
	committed = true
	RecordTaskRun(ctx, "register", true, "")
	glog.Infof(ctx, "[simuser] registered account=%s wxId=%d", account, wxID)
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

// RunPostVideoSubmitTask T4：提交视频生成并内联轮询 async-result 直至发帖或失败。
func RunPostVideoSubmitTask(ctx context.Context, password string, flags RuntimeFlags) {
	if !taskEnabled(ctx) {
		return
	}
	if IsVideoPostInFlight() {
		RecordTaskRun(ctx, "post_video_submit", true, "video poll in progress")
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
	jobID, err := InsertVideoJob(ctx, sess.WxID, textResp.Content, taskID)
	if err != nil {
		RecordTaskRun(ctx, "post_video_submit", false, err.Error())
		return
	}
	job := videoJobRow{Id: jobID, WxId: sess.WxID, Content: textResp.Content, TaskId: taskID}
	setVideoPostInFlight(true)
	if isManualRun(ctx) {
		runVideoPollPipeline(ctx, sess, job, flags)
		return
	}
	go runVideoPollPipeline(context.Background(), sess, job, flags)
}

// runVideoPollPipeline 轮询视频生成结果并发帖；结束时写 sim_task_run 并清除全局单飞标志。
func runVideoPollPipeline(ctx context.Context, sess loginSession, job videoJobRow, flags RuntimeFlags) {
	interval := flags.PostVideoPollInterval
	if interval <= 0 {
		interval = 2 * time.Minute
	}
	maxWait := flags.PostVideoPollMaxWait
	if maxWait <= 0 {
		maxWait = 30 * time.Minute
	}
	deadline := time.Now().Add(maxWait)
	var success bool
	var errMsg string
	defer func() {
		setVideoPostInFlight(false)
		if success {
			RecordTaskRun(ctx, "post_video_submit", true, "")
		} else {
			RecordTaskRun(ctx, "post_video_submit", false, errMsg)
		}
	}()
	for time.Now().Before(deadline) {
		res, pErr := aimodel.PollVideoGeneration(ctx, job.TaskId)
		if pErr != nil {
			time.Sleep(interval)
			continue
		}
		switch res.Status {
		case aimodel.VideoStatusFailed:
			_ = UpdateVideoJobStatus(ctx, job.Id, "skipped")
			errMsg = "video generation failed"
			return
		case aimodel.VideoStatusSuccess:
			vKey, uErr := uploadVideoFromURL(ctx, sess.AccessToken, res.VideoURL)
			if uErr != nil {
				_ = UpdateVideoJobStatus(ctx, job.Id, "skipped")
				errMsg = uErr.Error()
				return
			}
			if err := appPost(ctx, sess.AccessToken, "/ucg/app/api/posts", g.Map{
				"content": job.Content, "mediaType": 2, "submit": true,
				"media": []g.Map{{"objectKey": vKey, "mediaKind": 2, "sortOrder": 0}},
			}, nil); err != nil {
				_ = UpdateVideoJobStatus(ctx, job.Id, "skipped")
				errMsg = err.Error()
				return
			}
			_ = UpdateVideoJobStatus(ctx, job.Id, "done")
			success = true
			return
		default:
			_ = UpdateVideoJobStatus(ctx, job.Id, "processing")
			time.Sleep(interval)
		}
	}
	_ = UpdateVideoJobStatus(ctx, job.Id, "skipped")
	errMsg = "poll timeout"
}

// RunChatScanTask T5：全量 sim ids → ucg 未读抽样 → 单次 LLM 回复。
func RunChatScanTask(ctx context.Context, password string, flags RuntimeFlags) {
	_ = password
	_ = flags
	if !taskEnabled(ctx) {
		return
	}
	simWxIds, err := listAllSimWxIds(ctx)
	if err != nil {
		RecordTaskRun(ctx, "chat_scan", false, err.Error())
		return
	}
	if len(simWxIds) == 0 {
		RecordTaskRun(ctx, "chat_scan", false, "无模拟用户")
		return
	}
	sample, err := sampleSimUnreadChat(ctx, simWxIds)
	if err != nil {
		RecordTaskRun(ctx, "chat_scan", false, err.Error())
		return
	}
	if !sample.Found || sample.ConversationId == 0 || sample.SimWxId <= 0 {
		RecordTaskRun(ctx, "chat_scan", false, "无未读会话")
		return
	}
	history := buildChatHistoryFromSample(sample.Messages)
	_, user, _ := LoadRenderedPrompt(ctx, "chat_reply", map[string]string{"chat_history": history})
	resp, err := aimodel.Invoke(ctx, aimodel.LaneSimText, simChatRequest(simTempChatReply, 512, []aimodel.Message{{Role: "user", Content: user}}))
	if err != nil {
		RecordTaskRun(ctx, "chat_scan", false, err.Error())
		return
	}
	if err = sendInternalChat(ctx, sample.SimWxId, sample.ConversationId, strings.TrimSpace(resp.Content)); err != nil {
		RecordTaskRun(ctx, "chat_scan", false, err.Error())
		return
	}
	RecordTaskRun(ctx, "chat_scan", true, "")
}

// RunFollowTask T6：sim 关注发过帖的真人 author。
func RunFollowTask(ctx context.Context, password string) {
	if !taskEnabled(ctx) {
		return
	}
	simWxIds, err := listAllSimWxIds(ctx)
	if err != nil {
		RecordTaskRun(ctx, "follow", false, err.Error())
		return
	}
	follower, err := pickRandomSimWx(ctx)
	if err != nil {
		RecordTaskRun(ctx, "follow", false, err.Error())
		return
	}
	var targetWxId int64
	for i := 0; i < simFollowRandomMaxTry; i++ {
		authorWxId, aErr := sampleRandomRealAuthor(ctx, simWxIds)
		if aErr != nil {
			RecordTaskRun(ctx, "follow", false, aErr.Error())
			return
		}
		if authorWxId != follower.WxId {
			targetWxId = authorWxId
			break
		}
	}
	if targetWxId <= 0 {
		RecordTaskRun(ctx, "follow", false, "无可用关注目标")
		return
	}
	sess, err := usernameLogin(ctx, follower.Account, password)
	if err != nil {
		RecordTaskRun(ctx, "follow", false, err.Error())
		return
	}
	path := fmt.Sprintf("/ucg/app/api/follow/%d", targetWxId)
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

func buildChatHistoryFromSample(msgs []struct {
	SenderWxId int64  `json:"senderWxId"`
	Content    string `json:"content"`
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
