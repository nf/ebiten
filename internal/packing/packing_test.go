// Copyright 2018 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package packing_test

import (
	"testing"

	. "github.com/hajimehoshi/ebiten/internal/packing"
)

func TestPage(t *testing.T) {
	type Rect struct {
		X      int
		Y      int
		Width  int
		Height int
	}

	type Op struct {
		Width      int
		Height     int
		FreeNodeID int
	}

	cases := []struct {
		In  []Op
		Out []*Rect
	}{
		{
			In: []Op{
				{100, 100, -1},
				{100, 100, -1},
				{100, 100, -1},
				{100, 100, -1},
				{100, 100, -1},
				{100, 100, -1},
				{0, 0, 1},
				{0, 0, 3},
				{0, 0, 5},
				{0, 0, 0},
				{0, 0, 2},
				{0, 0, 4},
				{1024, 1024, -1},
			},
			Out: []*Rect{
				{0, 0, 100, 100},
				{0, 100, 100, 100},
				{0, 200, 100, 100},
				{0, 300, 100, 100},
				{0, 400, 100, 100},
				{0, 500, 100, 100},
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				{0, 0, 1024, 1024},
			},
		},
		{
			In: []Op{
				{100, 200, -1},
				{1024, 1024, -1},
				{100, 200, -1},
				{50, 50, -1},
				{200, 200, -1},
				{1024, 1024, -1},
				{500, 500, -1},
				{600, 600, -1},
				{100, 100, -1},
				{0, 0, 2},
				{100, 200, -1},
			},
			Out: []*Rect{
				{0, 0, 100, 200},
				nil,
				{0, 200, 100, 200},
				{0, 400, 50, 50},
				{100, 0, 200, 200},
				nil,
				{100, 200, 500, 500},
				nil,
				{0, 450, 100, 100},
				nil,
				{0, 200, 100, 200},
			},
		},
		{
			In: []Op{
				{256, 256, -1},
				{256, 256, -1},
				{256, 256, -1},
				{256, 256, -1},

				{256, 256, -1},
				{256, 256, -1},
				{256, 256, -1},
				{256, 256, -1},

				{256, 256, -1},
				{256, 256, -1},
				{256, 256, -1},
				{256, 256, -1},

				{256, 256, -1},
				{256, 256, -1},
				{256, 256, -1},
				{256, 256, -1},

				{256, 256, -1},
			},
			Out: []*Rect{
				{0, 0, 256, 256},
				{0, 256, 256, 256},
				{0, 512, 256, 256},
				{0, 768, 256, 256},

				{256, 0, 256, 256},
				{512, 0, 256, 256},
				{768, 0, 256, 256},
				{256, 256, 256, 256},

				{256, 512, 256, 256},
				{256, 768, 256, 256},
				{512, 256, 256, 256},
				{768, 256, 256, 256},

				{512, 512, 256, 256},
				{512, 768, 256, 256},
				{768, 512, 256, 256},
				{768, 768, 256, 256},

				nil,
			},
		},
		{
			In: []Op{
				{300, 300, -1},
				{300, 300, -1},
				{300, 300, -1},
				{300, 300, -1},
				{300, 300, -1},
				{300, 300, -1},
				{300, 300, -1},
				{300, 300, -1},
				{300, 300, -1},
				{300, 300, -1},
			},
			Out: []*Rect{
				{0, 0, 300, 300},
				{0, 300, 300, 300},
				{0, 600, 300, 300},
				{300, 0, 300, 300},
				{600, 0, 300, 300},
				{300, 300, 300, 300},
				{300, 600, 300, 300},
				{600, 300, 300, 300},
				{600, 600, 300, 300},
				nil,
			},
		},
	}

	for caseIndex, c := range cases {
		p := &Page{}
		nodes := []*Node{}
		for _, in := range c.In {
			if in.FreeNodeID == -1 {
				n := p.Alloc(in.Width, in.Height)
				nodes = append(nodes, n)
			} else {
				p.Free(nodes[in.FreeNodeID])
				nodes = append(nodes, nil)
			}
		}
		for i, out := range c.Out {
			if nodes[i] == nil {
				if out != nil {
					t.Errorf("(%d) nodes[%d]: should be nil but %v", caseIndex, i, out)
				}
				continue
			}
			x, y, width, height := nodes[i].Region()
			got := Rect{x, y, width, height}
			if out == nil {
				t.Errorf("(%d) nodes[%d]: got: %v, want: %v", caseIndex, i, got, nil)
				continue
			}
			want := *out
			if got != want {
				t.Errorf("(%d) nodes[%d]: got: %v, want: %v", caseIndex, i, got, want)
			}
		}
	}
}

func TestExtend(t *testing.T) {
	p := &Page{}
	s := p.Size()
	p.Alloc(s/2, s/2)
	// Now Extend extends the page by exntending the existing edge nodes.
	p.Extend()
	if p.Alloc(s*3/2, s*2) == nil {
		t.Fail()
	}
	if p.Alloc(s/2, s*3/2) == nil {
		t.Fail()
	}
	if p.Alloc(1, 1) != nil {
		t.Fail()
	}
}

func TestExtendAddingNewNodes(t *testing.T) {
	p := &Page{}
	s := p.Size()
	p.Alloc(s/2, s/2)
	n1 := p.Alloc(s/2, s/2)
	n2 := p.Alloc(s/2, s/2)
	p.Alloc(s/2, s/2)
	p.Free(n1)
	p.Free(n2)
	// There is an allocation at lower left.
	// In this case, Extend doesn't extend the existing edge nodes, but
	// instead adds new nodes.
	p.Extend()
	if p.Alloc(s*3/2, s*2) != nil {
		t.Fail()
	}
	if p.Alloc(s, s*2) == nil {
		t.Fail()
	}
	if p.Alloc(s, s) == nil {
		t.Fail()
	}
	if p.Alloc(s, s) != nil {
		t.Fail()
	}
}
