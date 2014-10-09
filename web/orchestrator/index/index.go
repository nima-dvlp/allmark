// Copyright 2014 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package index

import (
	"github.com/andreaskoch/allmark2/common/logger"
	"github.com/andreaskoch/allmark2/common/route"
	"github.com/andreaskoch/allmark2/dataaccess"
)

func New(logger logger.Logger, items []*dataaccess.Item) *Index {
	index := &Index{
		logger: logger,

		itemList: make([]*dataaccess.Item, 0),
		routeMap: make(map[string]*dataaccess.Item),
		itemTree: newItemTree(logger),
	}

	for _, item := range items {
		index.add(item)
	}

	return index
}

type Index struct {
	logger logger.Logger

	// indizes
	itemList []*dataaccess.Item
	routeMap map[string]*dataaccess.Item // route -> item,
	itemTree *ItemTree
}

func (index *Index) IsMatch(r route.Route) (item *dataaccess.Item, isMatch bool) {

	// check for a direct match
	if item, isMatch = index.routeMap[route.ToKey(r)]; isMatch {
		return item, isMatch
	}

	// no match
	return nil, false
}

func (index *Index) IsFileMatch(r route.Route) (*dataaccess.File, bool) {

	var parent *dataaccess.Item
	parentRoute := r
	for parentRoute.Level() >= 0 {

		parent, _ = index.IsMatch(parentRoute)
		if parent == nil {

			// next level
			newParentRoute, exists := parentRoute.Parent()
			if !exists {
				break
			}

			parentRoute = newParentRoute
			continue
		}

		// found a non-virtual parent
		break

	}

	// abort if there is no non-virtual parent
	if parent == nil {
		index.logger.Warn("No file found for route %q", r)
		return nil, false
	}

	// check if the parent has a file with the supplied route
	if file := parent.GetFile(r); file != nil {
		return file, true
	}

	// file not found
	return nil, false
}

func (index *Index) GetParent(childRoute route.Route) *dataaccess.Item {

	if childRoute.IsEmpty() {
		return nil
	}

	// abort if the supplied route is already a root
	if childRoute.Level() == 0 {
		return nil
	}

	// get the parent route
	parentRoute, exists := childRoute.Parent()
	if !exists {
		return nil
	}

	item, isMatch := index.IsMatch(parentRoute)
	if !isMatch {
		return nil
	}

	return item
}

func (index *Index) Root() *dataaccess.Item {
	return index.itemTree.Root()
}

func (index *Index) Size() int {
	return len(index.itemList)
}

// Get all childs that match the given expression
func (index *Index) GetAllChilds(route route.Route, expression func(item *dataaccess.Item) bool) []*dataaccess.Item {

	childs := make([]*dataaccess.Item, 0)

	// get all direct childs of the supplied route
	directChilds := index.GetDirectChilds(route)

	for _, child := range directChilds {

		// evaluate expression
		if !expression(child) {
			continue
		}

		// append child
		childs = append(childs, child)

		// recurse
		childs = append(childs, index.GetAllChilds(child.Route(), expression)...)

	}

	// sort the items by ascending by route
	dataaccess.SortItemBy(sortItemsByRoute).Sort(childs)

	return childs
}

func (index *Index) GetDirectChilds(route route.Route) []*dataaccess.Item {
	// get all mathching childs
	childs := index.itemTree.GetChildItems(route)

	// sort the items by ascending by route
	dataaccess.SortItemBy(sortItemsByRoute).Sort(childs)

	return childs
}

func (index *Index) add(item *dataaccess.Item) {

	// abort if item is invalid
	if item == nil {
		index.logger.Warn("Cannot add an invalid item to the index.")
		return
	}

	// the the item to the indizes
	index.itemList = append(index.itemList, item)
	index.routeMap[route.ToKey(item.Route())] = item
	index.itemTree.Insert(item)
}

// sort the items by name
func sortItemsByRoute(item1, item2 *dataaccess.Item) bool {

	// ascending by route
	return item1.Route().Value() > item2.Route().Value()

}