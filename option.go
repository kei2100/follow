package follow

import (
	"time"

	"github.com/kei2100/follow/posfile"
)

type option struct {
	detectRotateDelay   time.Duration
	followRotate        bool
	positionFile        posfile.PositionFile
	watchRotateInterval time.Duration
}

// OptionFunc let you change follow.Reader behavior.
type OptionFunc func(o *option)

func (o *option) apply(opts ...OptionFunc) {
	o.followRotate = true
	o.watchRotateInterval = 200 * time.Millisecond
	o.detectRotateDelay = 5 * time.Second
	for _, fn := range opts {
		fn(o)
	}
}

// WithDetectRotateDelay let you change detectRotateDelay
func WithDetectRotateDelay(v time.Duration) OptionFunc {
	return func(o *option) {
		o.detectRotateDelay = v
	}
}

// WithFollowRotate let you change followRotate
func WithFollowRotate(follow bool) OptionFunc {
	return func(o *option) {
		o.followRotate = follow
	}
}

// WithPositionFile let you change positionFile
func WithPositionFile(positionFile posfile.PositionFile) OptionFunc {
	return func(o *option) {
		o.positionFile = positionFile
	}
}

// WithPositionFilePath let you change positionFile
func WithPositionFilePath(path string) (OptionFunc, error) {
	if path == "" {
		return WithPositionFile(nil), nil
	}
	pf, err := posfile.Open(path)
	if err != nil {
		return nil, err
	}
	return WithPositionFile(pf), nil
}

// WithWatchRotateInterval let you change watchRotateInterval
func WithWatchRotateInterval(v time.Duration) OptionFunc {
	return func(o *option) {
		o.watchRotateInterval = v
	}
}
