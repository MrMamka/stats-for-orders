package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArraysToDepthOrder(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		in := [][]float64{{1.1, 11}, {2.2, 22}, {3.3, 33}}
		expected := []DepthOrder{
			{Price: 1.1, BaseQty: 11},
			{Price: 2.2, BaseQty: 22},
			{Price: 3.3, BaseQty: 33},
		}
		out, err := arraysToDepthOrder(in)

		assert.NoError(t, err)
		assert.Equal(t, expected, out)
	})

	t.Run("Error", func(t *testing.T) {
		in := [][]float64{{1.1, 11}, {2.2, 22, 5}, {3.3, 33}}
		_, err := arraysToDepthOrder(in)
		assert.Error(t, err)
	})
}

func TestDepthOrderToArrays(t *testing.T) {
	in := []DepthOrder{
		{Price: 1.1, BaseQty: 11},
		{Price: 2.2, BaseQty: 22},
		{Price: 3.3, BaseQty: 33},
	}
	expected := [][]float64{{1.1, 11}, {2.2, 22}, {3.3, 33}}
	out := depthOrderToArrays(in)

	assert.Equal(t, expected, out)
}
