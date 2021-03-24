package cron

import "context"

type dataKey struct{}

func WithData(ctx context.Context, data []byte) context.Context {
	return context.WithValue(ctx, dataKey{}, data)
}

func DataFrom(ctx context.Context) []byte {
	return ctx.Value(dataKey{}).([]byte)
}
