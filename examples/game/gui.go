package main

import (
	"github.com/faiface/pixel"
	txt "github.com/faiface/pixel/text"
	"github.com/faiface/pixel/imdraw"
)

type GuiPolicy struct {
	Text txt.Text
	Background imdraw.IMDraw

}

func NewGuiPolicy(name string) *GuiPolicy {
	policy := GuiPolicy{}
	policy.Text.Write([]byte(name))

	return &policy
}

func (p *GuiPolicy) Draw(t pixel.Target, m pixel.Matrix) {
	p.Text.Draw(t, m)
}