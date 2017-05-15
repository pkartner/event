package main

import (
	"log"

	"engo.io/engo"
	"engo.io/ecs"
	"engo.io/engo/common"
)

type Scene struct {}

func (*Scene) Type() string {return "TimeGame"}

func (*Scene) Preload() {}

func (*Scene) Setup(world *ecs.World) {
	world.AddSystem(&common.RenderSystem{})
}

type Node struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

func main() {
	opts := engo.RunOptions{
		Title: "TimeGame",
		Width: 400,
		Height: 400,
	}
	engo.Run(opts, &Scene{})

	node := Node{BasicEntity: ecs.NewBasic()}
	node.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{10, 10},
		Width: 303,
		Height: 641,
	}

	texture, err := common.LoadedSprite("textures/city.png")
	if err != nil {
		log.Println("Unable to load texture: " + err.Error())
	}

	node.RenderComponent = common.RenderComponent{
		Drawable: texture,
		Scale: engo.Point{1, 1},
	}

}