package gql

import (
	"context"

	"github.com/interline-io/transitland-server/model"
)

// SEGMENTS

type segmentResolver struct{ *Resolver }

func (r *segmentResolver) SegmentPatterns(ctx context.Context, obj *model.Segment) ([]*model.SegmentPattern, error) {
	return For(ctx).SegmentPatternsBySegmentID.Load(ctx, model.SegmentPatternParam{SegmentID: obj.ID})()
}

// SEGMENT PATTERNS

type segmentPatternResolver struct{ *Resolver }

func (r *segmentPatternResolver) Segment(ctx context.Context, obj *model.SegmentPattern) (*model.Segment, error) {
	return For(ctx).SegmentsByID.Load(ctx, obj.SegmentID)()
}
