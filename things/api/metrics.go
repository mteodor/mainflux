// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// +build !test

package api

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/things"
)

var _ things.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     things.Service
}

// MetricsMiddleware instruments core service by tracking request count and
// latency.
func MetricsMiddleware(svc things.Service, counter metrics.Counter, latency metrics.Histogram) things.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (ms *metricsMiddleware) CreateThings(ctx context.Context, token string, ths ...things.Thing) (saved []things.Thing, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "create_things").Add(1)
		ms.latency.With("method", "create_things").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.CreateThings(ctx, token, ths...)
}

func (ms *metricsMiddleware) UpdateThing(ctx context.Context, token string, thing things.Thing) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_thing").Add(1)
		ms.latency.With("method", "update_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.UpdateThing(ctx, token, thing)
}

func (ms *metricsMiddleware) UpdateKey(ctx context.Context, token, id, key string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_key").Add(1)
		ms.latency.With("method", "update_key").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.UpdateKey(ctx, token, id, key)
}

func (ms *metricsMiddleware) ViewThing(ctx context.Context, token, id string) (things.Thing, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_thing").Add(1)
		ms.latency.With("method", "view_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ViewThing(ctx, token, id)
}

func (ms *metricsMiddleware) ListThings(ctx context.Context, token string, offset, limit uint64, name string, metadata things.Metadata) (things.Page, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_things").Add(1)
		ms.latency.With("method", "list_things").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ListThings(ctx, token, offset, limit, name, metadata)
}

func (ms *metricsMiddleware) ListThingsByChannel(ctx context.Context, token, id string, offset, limit uint64, connected bool) (things.Page, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_things_by_channel").Add(1)
		ms.latency.With("method", "list_things_by_channel").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ListThingsByChannel(ctx, token, id, offset, limit, connected)
}

func (ms *metricsMiddleware) RemoveThing(ctx context.Context, token, id string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "remove_thing").Add(1)
		ms.latency.With("method", "remove_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.RemoveThing(ctx, token, id)
}

func (ms *metricsMiddleware) CreateChannels(ctx context.Context, token string, channels ...things.Channel) (saved []things.Channel, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "create_channels").Add(1)
		ms.latency.With("method", "create_channels").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.CreateChannels(ctx, token, channels...)
}

func (ms *metricsMiddleware) UpdateChannel(ctx context.Context, token string, channel things.Channel) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_channel").Add(1)
		ms.latency.With("method", "update_channel").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.UpdateChannel(ctx, token, channel)
}

func (ms *metricsMiddleware) ViewChannel(ctx context.Context, token, id string) (things.Channel, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_channel").Add(1)
		ms.latency.With("method", "view_channel").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ViewChannel(ctx, token, id)
}

func (ms *metricsMiddleware) ListChannels(ctx context.Context, token string, offset, limit uint64, name string, metadata things.Metadata) (things.ChannelsPage, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_channels").Add(1)
		ms.latency.With("method", "list_channels").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ListChannels(ctx, token, offset, limit, name, metadata)
}

func (ms *metricsMiddleware) ListChannelsByThing(ctx context.Context, token, id string, offset, limit uint64, connected bool) (things.ChannelsPage, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_channels_by_thing").Add(1)
		ms.latency.With("method", "list_channels_by_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ListChannelsByThing(ctx, token, id, offset, limit, connected)
}

func (ms *metricsMiddleware) RemoveChannel(ctx context.Context, token, id string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "remove_channel").Add(1)
		ms.latency.With("method", "remove_channel").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.RemoveChannel(ctx, token, id)
}

func (ms *metricsMiddleware) Connect(ctx context.Context, token string, chIDs, thIDs []string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "connect").Add(1)
		ms.latency.With("method", "connect").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Connect(ctx, token, chIDs, thIDs)
}

func (ms *metricsMiddleware) Disconnect(ctx context.Context, token, chanID, thingID string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "disconnect").Add(1)
		ms.latency.With("method", "disconnect").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Disconnect(ctx, token, chanID, thingID)
}

func (ms *metricsMiddleware) CanAccessByKey(ctx context.Context, id, key string) (string, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "can_access_by_key").Add(1)
		ms.latency.With("method", "can_access_by_key").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.CanAccessByKey(ctx, id, key)
}

func (ms *metricsMiddleware) CanAccessByID(ctx context.Context, chanID, thingID string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "can_access_by_id").Add(1)
		ms.latency.With("method", "can_access_by_id").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.CanAccessByID(ctx, chanID, thingID)
}

func (ms *metricsMiddleware) Identify(ctx context.Context, key string) (string, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "identify").Add(1)
		ms.latency.With("method", "identify").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Identify(ctx, key)
}

func (ms *metricsMiddleware) CreateGroup(ctx context.Context, token string, group things.Group) (g things.Group, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "create_group").Add(1)
		ms.latency.With("method", "create_group").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.CreateGroup(ctx, token, group)
}

func (ms *metricsMiddleware) UpdateGroup(ctx context.Context, token string, group things.Group) (g things.Group, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_group").Add(1)
		ms.latency.With("method", "update_group").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateGroup(ctx, token, group)
}

func (ms *metricsMiddleware) RemoveGroup(ctx context.Context, token string, id string) (err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "remove_group").Add(1)
		ms.latency.With("method", "remove_group").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.RemoveGroup(ctx, token, id)
}

func (ms *metricsMiddleware) Group(ctx context.Context, token, id string) (g things.Group, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "group").Add(1)
		ms.latency.With("method", "group").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Group(ctx, token, id)
}

func (ms *metricsMiddleware) Groups(ctx context.Context, token, parentID string, offset, limit uint64, meta things.Metadata) (gp things.GroupPage, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "groups").Add(1)
		ms.latency.With("method", "groups").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Groups(ctx, token, parentID, offset, limit, meta)
}

func (ms *metricsMiddleware) Members(ctx context.Context, token, groupID string, offset, limit uint64, meta things.Metadata) (gp things.MemberPage, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "members").Add(1)
		ms.latency.With("method", "members").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Members(ctx, token, groupID, offset, limit, meta)
}

func (ms *metricsMiddleware) Memberships(ctx context.Context, token, groupID string, offset, limit uint64, meta things.Metadata) (gp things.GroupPage, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "memberships").Add(1)
		ms.latency.With("method", "membeships").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Memberships(ctx, token, groupID, offset, limit, meta)
}

func (ms *metricsMiddleware) Assign(ctx context.Context, token, memberID, groupID string) (err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "assign").Add(1)
		ms.latency.With("method", "assign").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Assign(ctx, token, memberID, groupID)
}

func (ms *metricsMiddleware) Unassign(ctx context.Context, token, memberID, groupID string) (err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "unassign").Add(1)
		ms.latency.With("method", "unassign").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Unassign(ctx, token, memberID, groupID)
}
