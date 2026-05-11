package entity

// AppVersion 对应 ai_voice_app.version 表：App 发版信息。
type AppVersion struct {
	Id            int64  `json:"id"            ` //
	LatestVersion string `json:"latestVersion" ` //
	ReleaseNotes  string `json:"releaseNotes"  ` //
	DownloadUrl   string `json:"downloadUrl"   ` //
	ForceUpdate   int    `json:"forceUpdate"   ` // 1 表示强制更新
	MinVersion    string `json:"minVersion"    ` // 可选：低于该版本则强制更新
	ReleaseDate   int64 `json:"releaseDate"    ` // 当前版本上线时间：Unix 时间戳（秒），与库列 release_date（bigint）一致
}
