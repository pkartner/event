package main

import (
	"fmt"

	"github.com/pkartner/event"
)

func EventCastFailError(expected, actual string) error {
	return fmt.Errorf("Event not right type expected: %s, actual: %s", expected, actual)
}

func (g *Game) SetPolicyHandler(e event.Event, s *event.Store) {
	event, ok := e.(*SetPolicyEvent)
	if !ok {
		panic(EventCastFailError(SetPolicyEventType, e.Type().String()))
	}
	store := GetBranchStore(s)
	if !event.State {
		delete(store.ActivePolicies, event.Policy)
		return
	}
	store.ActivePolicies[event.Policy] = struct{}{}
	
}

func (g *Game) NextTurnHandler(e event.Event, s*event.Store) {
	store := GetBranchStore(s)

	store.Weights = CalculateWeightMap(g.GameData.Policies, g.GameData.Values, store.ActivePolicies)
	store.Values = RecountValues(store.Values, store.Weights)

	store.Turn++
}