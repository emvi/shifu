package cms

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteRegular(t *testing.T) {
	r, err := NewRoute("/foo/Bar", Content{})
	assert.NoError(t, err)
	assert.True(t, r.isRaw)
	assert.Len(t, r.match, 0)
	assert.Equal(t, "/foo/Bar", r.raw)
}

func TestRouteMatcher(t *testing.T) {
	r, err := NewRoute("/match/{x:(this|that){1}}/route/{y}", Content{})
	assert.NoError(t, err)
	assert.False(t, r.isRaw)
	assert.Len(t, r.match, 5)
	assert.Equal(t, "/match/{x:(this|that){1}}/route/{y}", r.raw)
	assert.Equal(t, "", r.match[0].path)
	assert.Equal(t, "match", r.match[1].path)
	assert.Equal(t, "^(this|that){1}$", r.match[2].regex.String())
	assert.Equal(t, "x", r.match[2].variable)
	assert.Equal(t, "route", r.match[3].path)
	assert.Empty(t, r.match[4].path)
	assert.Nil(t, r.match[4].regex)
	assert.Equal(t, "y", r.match[4].variable)
}

func TestRouteMatcherError(t *testing.T) {
	r, err := NewRoute("/match/{x:(this|that){1}}/route/{y:[0-9]+", Content{})
	assert.Equal(t, fmt.Errorf(ErrMatcherBrackets, "/match/{x:(this|that){1}}/route/{y:[0-9]+").Error(), err.Error())
	assert.Nil(t, r)
	r, err = NewRoute("/match/{}/route/{y:[0-9]+}", Content{})
	assert.Equal(t, fmt.Errorf(ErrMatcherEmpty, "/match/{}/route/{y:[0-9]+}").Error(), err.Error())
	assert.Nil(t, r)
	r, err = NewRoute("/match/{ }/route/{y:[0-9]+}", Content{})
	assert.Equal(t, fmt.Errorf(ErrMatcherVariable, "/match/{ }/route/{y:[0-9]+}").Error(), err.Error())
	assert.Nil(t, r)
	r, err = NewRoute("/match/{x:**}/route/{y:[0-9]+}", Content{})
	assert.Equal(t, fmt.Errorf(ErrParsingMatcher, "/match/{x:**}/route/{y:[0-9]+}", "error parsing regexp: invalid nested repetition operator: `**`").Error(), err.Error())
	assert.Nil(t, r)
}

func TestRouteMatch(t *testing.T) {
	r, err := NewRoute("/match/{x:(this|that){1}}/route/{y:[0-9]+}", Content{})
	assert.NoError(t, err)
	_, match := r.Match("/")
	assert.False(t, match)
	_, match = r.Match("/match/a/route/123")
	assert.False(t, match)
	_, match = r.Match("/match/this/route/xyz")
	assert.False(t, match)
	_, match = r.Match("/match/this/route/123/")
	assert.False(t, match)
	_, match = r.Match("/match/this/route/123/x")
	assert.False(t, match)
	_, match = r.Match("/match/this/route/123x")
	assert.False(t, match)
	vars, match := r.Match("/match/this/route/123")
	assert.True(t, match)
	assert.Len(t, vars, 2)
	assert.Equal(t, vars["x"], "this")
	assert.Equal(t, vars["y"], "123")
	vars, match = r.Match("/match/that/route/321")
	assert.True(t, match)
	assert.Equal(t, vars["x"], "that")
	assert.Equal(t, vars["y"], "321")
}
