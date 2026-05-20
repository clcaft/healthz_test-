package logger

import "context"

type trackingInfoCtxKey struct{}

func GetCtxWithTrackingInfo(ctx context.Context, trackingInfo interface{}) context.Context {
	return context.WithValue(ctx, trackingInfoCtxKey{}, trackingInfo)
}

func GetTrackingInfoFromCtx(ctx context.Context) interface{} {
	return ctx.Value(trackingInfoCtxKey{})
}
