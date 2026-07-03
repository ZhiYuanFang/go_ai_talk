// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgPostVote is the golang structure for table ucg_post_vote.
type UcgPostVote struct {
	Id         uint64 `json:"id"         ` //
	PostId     uint64 `json:"postId"     ` //
	VoterWxId  uint64 `json:"voterWxId"  ` //
	Side       string `json:"side"       ` // left|right
	CreatedAt  int64  `json:"createdAt"  ` //
	UpdatedAt  int64  `json:"updatedAt"  ` //
}
