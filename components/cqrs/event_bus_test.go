package cqrs_test

import (
	"context"
	"testing"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEventBus(t *testing.T) {
	pub := newPublisherStub()
	generateTopic := func(commandName string) string {
		return ""
	}
	marshaler := cqrs.JSONMarshaler{}

	cb, err := cqrs.NewEventBus(pub, generateTopic, marshaler)
	assert.NotNil(t, cb)
	assert.NoError(t, err)

	cb, err = cqrs.NewEventBus(nil, generateTopic, marshaler)
	assert.Nil(t, cb)
	assert.Error(t, err)

	cb, err = cqrs.NewEventBus(pub, nil, marshaler)
	assert.Nil(t, cb)
	assert.Error(t, err)

	cb, err = cqrs.NewEventBus(pub, generateTopic, nil)
	assert.Nil(t, cb)
	assert.Error(t, err)
}

func TestEventBus_Send_ContextPropagation(t *testing.T) {
	publisher := newPublisherStub()

	eventBus, err := cqrs.NewEventBus(
		publisher,
		func(eventName string) string {
			return "whatever"
		},
		cqrs.JSONMarshaler{},
	)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), contextKey("key"), "value")

	err = eventBus.Publish(ctx, "message")
	require.NoError(t, err)

	assert.Equal(t, ctx, publisher.messages["whatever"][0].Context())
}

func TestEventBus_Send_topic_name(t *testing.T) {
	cb, err := cqrs.NewEventBus(
		assertPublishTopicPublisher{ExpectedTopic: "cqrs_test.TestEvent", T: t},
		func(commandName string) string {
			return commandName
		},
		cqrs.JSONMarshaler{},
	)
	require.NoError(t, err)

	err = cb.Publish(context.Background(), TestEvent{})
	require.NoError(t, err)
}

func TestEventBus_Send_OnPublish(t *testing.T) {
	publisher := newPublisherStub()

	eb, err := cqrs.NewEventBusWithConfig(
		publisher,
		cqrs.EventConfig{
			GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
				return "whatever", nil
			},
			Marshaler: cqrs.JSONMarshaler{},
			OnPublish: func(params cqrs.OnEventSendParams) error {
				params.Message.Metadata.Set("key", "value")
				return nil
			},
		},
	)
	require.NoError(t, err)

	err = eb.Publish(context.Background(), TestEvent{})
	require.NoError(t, err)

	assert.Equal(t, "value", publisher.messages["whatever"][0].Metadata.Get("key"))
}

func TestEventBus_Send_OnPublish_error(t *testing.T) {
	publisher := newPublisherStub()

	expectedErr := errors.New("some error")

	eb, err := cqrs.NewEventBusWithConfig(
		publisher,
		cqrs.EventConfig{
			GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
				return "whatever", nil
			},
			Marshaler: cqrs.JSONMarshaler{},
			OnPublish: func(params cqrs.OnEventSendParams) error {
				return expectedErr
			},
		},
	)
	require.NoError(t, err)

	err = eb.Publish(context.Background(), TestEvent{})
	require.EqualError(t, err, "cannot execute OnPublish: some error")
}
