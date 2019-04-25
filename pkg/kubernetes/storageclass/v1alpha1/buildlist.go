/*
Copyright 2019 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/pkg/errors"
	storagev1 "k8s.io/api/storage/v1"
)

// ListBuilder enables building an instance of StorageClassList
type ListBuilder struct {
	list    *StorageClassList
	filters predicateList
	errs    []error
}

// NewListBuilder returns a instance of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &StorageClassList{items: []*StorageClass{}}}
}

// Build returns the final instance of patch
func (b *ListBuilder) Build() (*StorageClassList, []error) {
	if len(b.errs) > 0 {
		return nil, b.errs
	}
	return b.list, nil
}

// ListBuilderForAPIList builds the ListBuilder object based on SC API list
func ListBuilderForAPIList(scl *storagev1.StorageClassList) *ListBuilder {
	b := &ListBuilder{list: &StorageClassList{}}
	if scl == nil {
		b.errs = append(b.errs, errors.New("failed to build pvc list: missing api list"))
		return b
	}
	for _, sc := range scl.Items {
		sc := sc
		b.list.items = append(b.list.items, &StorageClass{object: &sc})
	}
	return b
}

// ListBuilderForObjects returns a instance of ListBuilder from SC instances
func ListBuilderForObjects(scs ...*StorageClass) *ListBuilder {
	b := &ListBuilder{list: &StorageClassList{}}
	if scs == nil {
		b.errs = append(b.errs, errors.New("failed to build pvc list: missing objects"))
		return b
	}
	for _, sc := range scs {
		sc := sc
		b.list.items = append(b.list.items, sc)
	}
	return b
}

// List returns the list of StorageClass instances that was built by this builder
func (b *ListBuilder) List() *StorageClassList {
	if b.filters == nil && len(b.filters) == 0 {
		return b.list
	}
	filtered := &StorageClassList{}
	for _, sc := range b.list.items {
		if b.filters.all(sc) {
			sc := sc // Pin it
			filtered.items = append(filtered.items, sc)
		}
	}
	return filtered
}

// WithFilter add filters on which the StorageClass has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}

// APIList builds core API PVC list using listbuilder
func (b *ListBuilder) APIList() (*storagev1.StorageClassList, []error) {
	l, errs := b.Build()
	if len(errs) > 0 {
		return nil, errs
	}
	if l == nil {
		errs := append(errs, errors.New("failed to build pvc list: object list nil"))
		return nil, errs
	}
	return l.ToAPIList(), nil
}
