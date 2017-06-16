package main

import (
	"fmt"

	"github.com/pkartner/event"
)

type ValueMap map[string]float64
type WeightMap map[string]map[string]float64

type BranchStore struct {
	BranchID event.ID
	Values ValueMap
	Weights WeightMap
	ActivePolicies map[string]struct{}
	Turn uint64
}

// GetBranchID TODO
func(s *BranchStore) GetBranchID() event.ID{
    return s.BranchID
}

// SetBranchID TODO
func(s *BranchStore) SetBranchID(id event.ID) {
    s.BranchID = id
}

func NewBranchStore() event.BranchStore {
    return &BranchStore{}
}

func GetBranchStore(s *event.Store) *BranchStore {
    store, ok := s.Attributes.(*BranchStore)
    if !ok {
        panic(fmt.Errorf("Store not of right type"))
    }
    return store
} 

type Values struct {
	Values map[string]struct {
		AffectedBy []struct{
			ValueName string `json:"value_name"`
			Weight float64 `json:"weight"`
		} `json:"affected_by"`
	} `json:"values"`

}

type Policies struct {
	Policies map[string]struct {
		Name string `json:"name"`
		FlatAmountPerTurn []struct {
			ValueName string `json:"value_name"`
			Amount float64 `json:"amount"`
		} `json:"flat_amount_per_turn"`
		WeightIncrease []struct {
			DestValueName string `json:"dest_value_name"`
			SourceValueName string `json:"source_value_name"`
			Weight float64 `json:"weight"`
		}
	} `json:"policies"`
}

type Game struct {
	GameData GameData
	EventStore event.EventStore
	Dispatcher *event.TimelineDispatcher
}

type GameData struct {
	Values Values
	Policies Policies
}

func (g *Game) Restore() {
	event.RestoreEvents(g.EventStore, g.Dispatcher)
}

func CalculateWeightMap(policies Policies, values Values, activatedPolicies map[string]struct{}) WeightMap {
	weightMap := WeightMap{}
	for k, v := range values.Values {
		weightMap[k] = map[string]float64{}
		for _, v2 := range v.AffectedBy {
			weightMap[k][v2.ValueName] = v2.Weight
		}
	}
	for k, _ := range activatedPolicies {
		for _, v := range policies.Policies[k].WeightIncrease {
			value := weightMap[v.DestValueName][v.SourceValueName]
			weightMap[v.DestValueName][v.SourceValueName] = value + v.Weight
		}
	}

	return weightMap
}

func RecountValues(values ValueMap, weights WeightMap) ValueMap{
	newValues := ValueMap{}
	for k, _ := range values {
		newValues[k] = values[k] + CalculateAddedValue(k, values, weights)
	}
	return newValues
}

func CalculateAddedValue(key string, values ValueMap, weights WeightMap) float64 {
	weightsForValue := weights[key]
	adjustment := 0.0
	for k, v := range weightsForValue {
		adjustment += values[k] * v
	}
	return adjustment
}