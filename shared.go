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

package ebiten

import (
	"github.com/hajimehoshi/ebiten/internal/graphics"
	"github.com/hajimehoshi/ebiten/internal/opengl"
	"github.com/hajimehoshi/ebiten/internal/packing"
	"github.com/hajimehoshi/ebiten/internal/restorable"
	"github.com/hajimehoshi/ebiten/internal/sync"
)

type sharedImage struct {
	restorable *restorable.Image
	page       packing.Page
}

var (
	theSharedImages = []*sharedImage{}
)

type sharedImagePart struct {
	sharedImage *sharedImage
	node        *packing.Node
}

func (s *sharedImagePart) image() *restorable.Image {
	return s.sharedImage.restorable
}

func (s *sharedImagePart) region() (x, y, width, height int) {
	return s.node.Region()
}

func (s *sharedImagePart) Dispose() {
	s.sharedImage.page.Free(s.node)
	if s.sharedImage.page.IsEmpty() {
		s.sharedImage.restorable.Dispose()
		s.sharedImage.restorable = nil
		index := -1
		for i, sh := range theSharedImages {
			if sh == s.sharedImage {
				index = i
				break
			}
		}
		if index == -1 {
			panic("not reached")
		}
		theSharedImages = append(theSharedImages[:index], theSharedImages[index+1:]...)
	}
}

var sharedImageLock sync.Mutex

func newSharedImagePart(width, height int) *sharedImagePart {
	sharedImageLock.Lock()
	sharedImageLock.Unlock()

	if width > packing.MaxSize || height > packing.MaxSize {
		return nil
	}
	for _, s := range theSharedImages {
		for {
			if n := s.page.Alloc(width, height); n != nil {
				return &sharedImagePart{
					sharedImage: s,
					node:        n,
				}
			}
			if !s.page.Extend() {
				break
			}
			newSharedImage := restorable.NewImage(s.page.Size(), s.page.Size(), false)
			newSharedImage.DrawImage(s.restorable, 0, 0, s.page.Size(), s.page.Size(), nil, nil, opengl.CompositeModeCopy, graphics.FilterNearest)
			s.restorable.Dispose()

			s.restorable = newSharedImage
		}
	}

	s := &sharedImage{}
	var n *packing.Node
	for {
		n = s.page.Alloc(width, height)
		if n != nil {
			break
		}
		if !s.page.Extend() {
			break
		}
	}
	if n == nil {
		panic("not reached")
	}
	s.restorable = restorable.NewImage(s.page.Size(), s.page.Size(), false)
	theSharedImages = append(theSharedImages, s)

	return &sharedImagePart{
		sharedImage: s,
		node:        n,
	}
}
