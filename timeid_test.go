package event_test

import (
	"testing"

	"github.com/pkartner/event"

	"github.com/stretchr/testify/assert"
)

func TestIDPart(t *testing.T) {
	id := event.GenerateTimeID(5657657654,1)
	idPart := id.IDPart()

	assert.EqualValues(t, 1, idPart)
}