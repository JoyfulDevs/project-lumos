package retry_test

import (
	"context"
	"errors"
	"testing"

	"github.com/joyfuldevs/project-lumos/pkg/retry"
)

func TestDo(t *testing.T) {
	testCases := []struct {
		desc      string
		fn        func(ctx context.Context) error
		opts      []retry.Option
		wantErr   bool
		callCount int
	}{
		{
			desc: "success on first try",
			fn: func(ctx context.Context) error {
				return nil
			},
			opts:      nil,
			wantErr:   false,
			callCount: 1,
		},
		{
			desc: "success after 2 retries",
			fn: func() func(ctx context.Context) error {
				callCount := 0
				return func(ctx context.Context) error {
					callCount++
					if callCount < 3 {
						return errors.New("temporary error")
					}
					return nil
				}
			}(),
			opts: []retry.Option{
				retry.WithMaxRetries(5),
				retry.WithBackoff(0),
				retry.WithMaxBackoff(0),
			},
			wantErr:   false,
			callCount: 3,
		},
		{
			desc: "fail after max retries",
			fn: func(ctx context.Context) error {
				return errors.New("persistent error")
			},
			opts: []retry.Option{
				retry.WithMaxRetries(2),
				retry.WithBackoff(0),
				retry.WithMaxBackoff(0),
			},
			wantErr:   true,
			callCount: 3,
		},
		{
			desc: "non-retryable error",
			fn: func(ctx context.Context) error {
				return errors.New("non-retryable error")
			},
			opts: []retry.Option{
				retry.WithMaxRetries(3),
				retry.WithRetryable(func(err error) bool {
					return err.Error() != "non-retryable error"
				}),
			},
			wantErr:   true,
			callCount: 1,
		},
		{
			desc: "retryable error",
			fn: func() func(ctx context.Context) error {
				callCount := 0
				return func(ctx context.Context) error {
					callCount++
					if callCount < 3 {
						return errors.New("retryable error")
					}
					return nil
				}
			}(),
			opts: []retry.Option{
				retry.WithMaxRetries(5),
				retry.WithBackoff(0),
				retry.WithMaxBackoff(0),
				retry.WithRetryable(func(err error) bool {
					return err.Error() == "retryable error"
				}),
			},
			wantErr:   false,
			callCount: 3,
		},
		{
			desc: "non-retryable error after retries",
			fn: func() func(ctx context.Context) error {
				callCount := 0
				return func(ctx context.Context) error {
					callCount++
					if callCount < 2 {
						return errors.New("retryable error")
					} else {
						return errors.New("non-retryable error")
					}
				}
			}(),
			opts: []retry.Option{
				retry.WithMaxRetries(5),
				retry.WithBackoff(0),
				retry.WithMaxBackoff(0),
				retry.WithRetryable(func(err error) bool {
					return err.Error() == "retryable error"
				}),
			},
			wantErr:   true,
			callCount: 2,
		},
		{
			desc: "nil retryable function",
			fn: func(ctx context.Context) error {
				return errors.New("some error")
			},
			opts: []retry.Option{
				retry.WithMaxRetries(1),
				retry.WithBackoff(0),
				retry.WithMaxBackoff(0),
				retry.WithRetryable(nil),
			},
			wantErr:   true,
			callCount: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()

			callCount := 0
			wrappedFn := func(ctx context.Context) error {
				callCount++
				return tc.fn(ctx)
			}

			err := retry.Do(ctx, wrappedFn, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("expected error = %v, got %v", tc.wantErr, err != nil)
			}
			if callCount != tc.callCount {
				t.Errorf("expected call count = %d, got %d", tc.callCount, callCount)
			}
		})
	}
}

func TestDoWithData(t *testing.T) {
	testCases := []struct {
		desc      string
		fn        func(ctx context.Context) (string, error)
		opts      []retry.Option
		wantErr   bool
		wantData  string
		callCount int
	}{
		{
			desc: "success on first try",
			fn: func(ctx context.Context) (string, error) {
				return "success", nil
			},
			opts:      nil,
			wantErr:   false,
			wantData:  "success",
			callCount: 1,
		},
		{
			desc: "success after 2 retries",
			fn: func() func(ctx context.Context) (string, error) {
				callCount := 0
				return func(ctx context.Context) (string, error) {
					callCount++
					if callCount < 3 {
						return "", errors.New("temporary error")
					}
					return "success after retries", nil
				}
			}(),
			opts: []retry.Option{
				retry.WithMaxRetries(5),
				retry.WithBackoff(0),
				retry.WithMaxBackoff(0),
			},
			wantErr:   false,
			wantData:  "success after retries",
			callCount: 3,
		},
		{
			desc: "fail after max retries",
			fn: func(ctx context.Context) (string, error) {
				return "partial data", errors.New("persistent error")
			},
			opts: []retry.Option{
				retry.WithMaxRetries(2),
				retry.WithBackoff(0),
				retry.WithMaxBackoff(0),
			},
			wantErr:   true,
			wantData:  "partial data",
			callCount: 3,
		},
		{
			desc: "non-retryable error",
			fn: func(ctx context.Context) (string, error) {
				return "error data", errors.New("non-retryable error")
			},
			opts: []retry.Option{
				retry.WithMaxRetries(3),
				retry.WithRetryable(func(err error) bool {
					return err.Error() != "non-retryable error"
				}),
			},
			wantErr:   true,
			wantData:  "error data",
			callCount: 1,
		},
		{
			desc: "retryable error",
			fn: func() func(ctx context.Context) (string, error) {
				callCount := 0
				return func(ctx context.Context) (string, error) {
					callCount++
					if callCount < 3 {
						return "", errors.New("retryable error")
					}
					return "final success", nil
				}
			}(),
			opts: []retry.Option{
				retry.WithMaxRetries(5),
				retry.WithBackoff(0),
				retry.WithMaxBackoff(0),
				retry.WithRetryable(func(err error) bool {
					return err.Error() == "retryable error"
				}),
			},
			wantErr:   false,
			wantData:  "final success",
			callCount: 3,
		},
		{
			desc: "non-retryable error after retries",
			fn: func() func(ctx context.Context) (string, error) {
				callCount := 0
				return func(ctx context.Context) (string, error) {
					callCount++
					if callCount < 2 {
						return "", errors.New("retryable error")
					} else {
						return "last attempt data", errors.New("non-retryable error")
					}
				}
			}(),
			opts: []retry.Option{
				retry.WithMaxRetries(5),
				retry.WithBackoff(0),
				retry.WithMaxBackoff(0),
				retry.WithRetryable(func(err error) bool {
					return err.Error() == "retryable error"
				}),
			},
			wantErr:   true,
			wantData:  "last attempt data",
			callCount: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()

			callCount := 0
			wrappedFn := func(ctx context.Context) (string, error) {
				callCount++
				return tc.fn(ctx)
			}

			result, err := retry.DoWithData(ctx, wrappedFn, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("expected error = %v, got %v", tc.wantErr, err != nil)
			}
			if result != tc.wantData {
				t.Errorf("expected result = %s, got %s", tc.wantData, result)
			}
			if callCount != tc.callCount {
				t.Errorf("expected call count = %d, got %d", tc.callCount, callCount)
			}
		})
	}
}
