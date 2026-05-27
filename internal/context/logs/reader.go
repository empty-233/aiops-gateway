package logs

import "context"

type Reader interface {
	Query(ctx context.Context, source Source, options *Options) (string, error)
}