package scheduler

import "context"

type jobKey struct{}

func WithJobName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, jobKey{}, name)
}

func JobNameFrom(ctx context.Context) string {
	return ctx.Value(jobKey{}).(string)
}
