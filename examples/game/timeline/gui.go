package timeline

import (
	"github.com/faiface/pixel/imdraw"
)

type TimelineNode struct {
	X int
	Y int
}

type TimelineNodeConnection struct {
	StartID int
	EndID int
}

type Timeline struct {
	NodeSize uint
	X int
	Y int
}

func (t *Timeline) Draw(draw *imdraw.IMDraw) {
	draw.Ellipse()
}