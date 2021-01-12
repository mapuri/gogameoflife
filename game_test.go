package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStep(t *testing.T) {
	tests := map[string]struct {
		activeCells []Cell
		exptdCells  map[Cell]struct{}
	}{
		"only-one-active-cell-dies": {
			activeCells: []Cell{
				{Row: 0, Col: 1},
			},
			exptdCells: map[Cell]struct{}{},
		},
		"only-two-active-cell-die": {
			activeCells: []Cell{
				{Row: 0, Col: 0},
				{Row: 0, Col: 1},
			},
			exptdCells: map[Cell]struct{}{},
		},
		"three-adjcent-cells-alternate-cycle": {
			activeCells: []Cell{
				{Row: 0, Col: 0},
				{Row: 0, Col: 1},
				{Row: 0, Col: 2},
			},
			exptdCells: map[Cell]struct{}{
				{Row: 1, Col: 1}:  {},
				{Row: 0, Col: 1}:  {},
				{Row: -1, Col: 1}: {},
			},
		},
		"three-cells-generate--new-cell": {
			activeCells: []Cell{
				{Row: 0, Col: 0},
				{Row: 0, Col: 1},
				{Row: -1, Col: 0},
			},
			exptdCells: map[Cell]struct{}{
				{Row: 0, Col: 0}:  {},
				{Row: 0, Col: 1}:  {},
				{Row: -1, Col: 0}: {},
				{Row: -1, Col: 1}: {},
			},
		},
		"four-cells-never-changing-config": {
			activeCells: []Cell{
				{Row: 0, Col: 0},
				{Row: 0, Col: 2},
				{Row: 1, Col: 1},
				{Row: -1, Col: 1},
			},
			exptdCells: map[Cell]struct{}{
				{Row: 0, Col: 0}:  {},
				{Row: 0, Col: 2}:  {},
				{Row: 1, Col: 1}:  {},
				{Row: -1, Col: 1}: {},
			},
		},
	}
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			b := NewBoard(test.activeCells)
			b.Step()
			assert.Equal(t, test.exptdCells, b.ActiveCells)
		})
	}
}
