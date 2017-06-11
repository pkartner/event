package main

import (
	"io/ioutil"
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"

	"github.com/pkartner/event"
	
	//"github.com/faiface/pixel/imdraw"
)



func LoadData(policies *Policies, values *Values) {
	data, err := ioutil.ReadFile("policies.json")
	if err != nil {
		panic (err)
	}	
	if err := json.Unmarshal(data, &policies); err != nil {
		panic(err)
	}
	data, err = ioutil.ReadFile("values.json")
	if err != nil {
		panic (err)
	}	
	if err := json.Unmarshal(data, &values); err != nil {
		panic(err)
	}
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title: "Timeline",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.Clear(colornames.Skyblue)

	// Load Data
	policies := Policies{}
	values := Values{}
	LoadData(&policies, &values)
	gameData := GameData{
		Policies: policies,
		Values: values,
	}

	// Create GUI
	guiPolicies := []*GuiPolicy{}
	for _, v := range policies.Policies {
		guiPolicy := NewGuiPolicy(v.Name)
		guiPolicies = append(guiPolicies, guiPolicy)
	}

	// Create Game object
	game := NewGame("time", &gameData)
	_ = game
	for !win.Closed() {
		m := pixel.IM.Moved(pixel.V(32, 32))
		for _, v := range guiPolicies {
			v.Draw(win, m)
		}
		win.Update()
	}
}

func NewGame(fileName string, GameData *GameData) *Game {
	game := Game{}
	// Setup Event stuff
    db, err := bolt.Open(fileName+".db", 0600, nil)
    if nil != err {
        panic(err)
    }
    eventStore := event.NewBoltEventStore(db)
    timeStore := event.NewTimelineStore(NewBranchStore, event.Reloader{
        EventStore: eventStore,
    }, nil)
    dispatcher := event.NewTimelineDispatcher(timeStore)
    dispatcher.SetMiddleware(
        event.EventStoreMiddleware(eventStore),
    )
    dispatcher.Dispatcher.Register(&event.WindbackEvent{}, dispatcher.WindbackHandler)
    dispatcher.Dispatcher.Register(&event.NewBranchEvent{}, dispatcher.NewBranchHandler)
    dispatcher.Register(&NextTurnEvent{}, game.NextTurnHandler)
    dispatcher.Register(&SetPolicyEvent{}, game.SetPolicyHandler)

	return &game
}

func main() {
	pixelgl.Run(run)
}