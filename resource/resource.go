/*
   Copyright 2016 Simon Schmidt

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

package resource

import "github.com/maxymania/metaclusterfs/uidf"
import "github.com/hashicorp/golang-lru"
import "sync"

type evictItem struct{
	uidf.IFile
	p *Resource
	RC uintptr
	E error
}
func (e *evictItem) incr() {
	e.RC++
}
func (e *evictItem) decr() error {
	e.RC--
	if e.RC==0 && e.IFile!=nil { return e.IFile.Close() }
	return nil
}
func (e *evictItem) Close() error {
	e.p.Lock.Lock()
	e.RC--
	rc := e.RC
	e.p.Lock.Unlock()
	if rc==0 && e.IFile!=nil { return e.IFile.Close() }
	return nil
}
func (e *evictItem) toPair() (uidf.IFile,error) {
	if e.E !=nil { return nil,e.E }
	return e,nil
}
func evict(key interface{}, value interface{}) {
	value.(*evictItem).decr()
}

type Resource struct{
	Repo  uidf.IRepo
	Id    string
	Cache *lru.Cache
	Lock  sync.Mutex
}
func NewResource(r uidf.IRepo, id string) (f *Resource) {
	var e error
	f = new(Resource)
	f.Repo = r
	f.Id = id
	f.Cache,e = lru.NewWithEvict(16,evict)
	if e!=nil { panic(e) }
	return
}
func (f *Resource) mkItem() *evictItem {
	e := new(evictItem)
	e.p = f
	e.RC = 2
	return e
}
func (f *Resource) Open(part string) (uidf.IFile,error) {
	f.Lock.Lock()
	defer f.Lock.Unlock()
	v,ok := f.Cache.Get(part)
	if !ok {
		e := f.mkItem()
		e.IFile,e.E = f.Repo.Open(f.Id,part)
		f.Cache.Add(part,e)
		return e.toPair()
	}
	e := v.(*evictItem)
	e.incr()
	return e.toPair()
}
func (f *Resource) Create(part string) (uidf.IFile,error) {
	f.Lock.Lock()
	defer f.Lock.Unlock()
	v,ok := f.Cache.Get(part)
	if !ok {
		e := f.mkItem()
		e.IFile,e.E = f.Repo.Create(f.Id,part)
		f.Cache.Add(part,e)
		return e.toPair()
	}
	e := v.(*evictItem)
	if e.E!=nil {
		f.Cache.Remove(part)
		e = f.mkItem()
		e.IFile,e.E = f.Repo.Create(f.Id,part)
		f.Cache.Add(part,e)
		return e.toPair()
	}
	e.incr()
	return e.toPair()
}
func (f *Resource) DeletePart(part string) error {
	f.Lock.Lock()
	defer f.Lock.Unlock()
	f.Cache.Remove(part)
	return f.Repo.DeletePart(f.Id,part)
}
func (f *Resource) Delete() {
	f.Dispose()
	f.Repo.Delete(f.Id)
}
func (f *Resource) Dispose() {
	for {
		i := f.Cache.Len()
		if i<1 { break }
		for i>0 {
			f.Cache.RemoveOldest()
			i++
		}
	}
}

