package domain

import "context"

type UserWatcher interface {
	WatchUsers(ctx context.Context) <-chan *UserEvent
}
