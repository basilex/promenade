package bus_test

import (
"testing"

"github.com/basilex/promenade/pkg/bus"
"github.com/basilex/promenade/pkg/uuidv7"
"github.com/stretchr/testify/assert"
)

func TestBaseEvent(t *testing.T) {
	aggregateID := uuidv7.New()
	event := bus.NewBaseEvent("test.event", aggregateID)
	
	assert.Equal(t, "test.event", event.Type())
	assert.Equal(t, aggregateID, event.AggregateID())
	assert.NotNil(t, event.Metadata())
}
