package ucg

import (
	"context"
	"strconv"

	"hello/internal/platform/cachekit"
)

// BackfillPublishedPostRedis 运维 backfill：单帖 ZADD/GEO/snapshot。
func BackfillPublishedPostRedis(ctx context.Context, postID uint64) error {
	return syncPublishedPostRedis(ctx, postID)
}

// BackfillPostVoteCountsRedis 运维 backfill：单帖投票 Hash + PostSnapshot 票数。
func BackfillPostVoteCountsRedis(ctx context.Context, postID uint64) error {
	if postID == 0 {
		return nil
	}
	if _, err := backfillVoteCountsFromMySQL(ctx, postID); err != nil {
		return err
	}
	refreshPostSnapshotFromDB(ctx, postID)
	return nil
}

// BackfillUserLikedPosts 运维 backfill：重建用户 liked SET（覆盖式 SADD）。
func BackfillUserLikedPosts(ctx context.Context, wxID int64, postIDs []uint64) error {
	if wxID <= 0 || len(postIDs) == 0 {
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
	return ucgCache.SetAdd(ctx, cachekit.UCGUserLikedPostsKey(wxID), members...)
}
