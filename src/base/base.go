package base

import (
	"context"

	log "github.com/sirupsen/logrus"
)

type Option func(*Base) error

type Base struct {
	ctx    context.Context
	cancel context.CancelFunc
	Log    log.FieldLogger
}

// Generate the Base of the shared module base
func NewBase(opts ...Option) (*Base, error) {
	contxt, cancel := context.WithCancel(context.Background())
	base := &Base{
		ctx:    contxt,
		cancel: cancel,
		Log:    NewDefaultLogger(),
	}
	// Iterate throung the options given
	for _, opt := range opts {
		err := opt(base)
		if err != nil {
			return nil, err
		}
	}
	return base, nil
}

// Set specific context to the base model
func WithContext(ctx context.Context) Option {
	return func(b *Base) error {
		b.ctx = ctx
		return nil
	}
}

// Set specific Logger for the Base model
func WithLogger(opts LogOpts) Option {
	return func(b *Base) error {
		logger := log.New()
		logger.SetFormatter(ParseLogFormatter(opts.Formatter))
		logger.SetOutput(ParseLogOutput(opts.Output))
		logger.SetLevel(ParseLogLevel(opts.Level))
		l := logger.WithField("module", opts.ModName)
		b.Log = l
		return nil
	}
}

// function that return the context of the Base module
func (b *Base) Ctx() context.Context {
	return b.ctx
}

// function that cancels the base of the project
func (b *Base) Cancel() {
	b.cancel()
	return
}
