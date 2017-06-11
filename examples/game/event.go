package main

import (
	"github.com/pkartner/event"
)

const (
	NextTurnEventType = "next_turn"
	SetPolicyEventType = "set_policy"
)

type NextTurnEvent struct {
	event.BaseTimelineEvent
}

// Type TODO
func (e *NextTurnEvent) Type() event.EventType {
    return NextTurnEventType
}

func NextTurn() *NextTurnEvent{
	e := NextTurnEvent{}

	return &e
}

type SetPolicyEvent struct {
	event.BaseTimelineEvent
	Policy string
	State bool
}

func (e *SetPolicyEvent) Type() event.EventType {
	return SetPolicyEventType
}

func SetPolicy() *SetPolicyEvent {
	e := SetPolicyEvent{}

	return &e
}


