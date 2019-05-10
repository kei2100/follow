package follow

import (
	"time"

	"github.com/kei2100/follow/posfile"
)

type option struct {
	rotatedFilePathPatterns []string
	positionFile            posfile.PositionFile
	optionFollowRotate
}

type optionFollowRotate struct {
	detectRotateDelay   time.Duration
	followRotate        bool
	watchRotateInterval time.Duration
}

// OptionFunc let you change follow.Reader behavior.
type OptionFunc func(o *option)

// Default values
const (
	DefaultFollowRotate      = true
	DefaultRotateInterval    = 100 * time.Millisecond
	DefaultDetectRotateDelay = 5 * time.Second
)

func (o *option) apply(opts ...OptionFunc) {
	o.followRotate = DefaultFollowRotate
	o.watchRotateInterval = DefaultRotateInterval
	o.detectRotateDelay = DefaultDetectRotateDelay
	for _, fn := range opts {
		fn(o)
	}
}

// WithRotatedFilePathPatterns let you change rotatedFilePathPatterns
func WithRotatedFilePathPatterns(globPatterns []string) OptionFunc {
	return func(o *option) {
		o.rotatedFilePathPatterns = globPatterns
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

// WithWatchRotateInterval let you change watchRotateInterval
func WithWatchRotateInterval(v time.Duration) OptionFunc {
	return func(o *option) {
		o.watchRotateInterval = v
	}
}
