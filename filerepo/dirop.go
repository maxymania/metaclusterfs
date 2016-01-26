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

package filerepo

import "github.com/maxymania/metaclusterfs/lockman"
import "github.com/maxymania/metaclusterfs/resource"
import "github.com/maxymania/metaclusterfs/joinf"
import "github.com/satori/go.uuid"
import "syscall"

type Entry struct{
	IsDir bool
	Name string
}

type Directory struct{
	Repo *Repository
	Res *LocalRes
	Obj *lockman.ObjEntry
}
func (d *Directory) Dispose() {
	d.Obj.Decr()
}
func (d *Directory) load(id string) (*Directory,*File,error) {
	obj := d.Repo.Load(id)
	lr := obj.Obj.(*LocalRes)
	if lr.Dir!=nil {
		return &Directory{d.Repo,lr,obj},nil,nil
	} else if lr.File!=nil {
		return nil,&File{lr,obj},nil
	}
	obj.Decr()
	return nil,nil,syscall.EIO
}
func (d *Directory) Lookup(name string) (*Directory,*File,error) {
	d.Obj.Lock()
	defer d.Obj.Unlock()
	id,e := d.Res.Dir.Find(name)
	if e!=nil { return nil,nil,e }
	return d.load(id)
}
func (d *Directory) Listup() []Entry {
	d.Obj.Lock()
	defer d.Obj.Unlock()
	r := d.Res.Dir.ListUp()
	el := make([]Entry,0,len(r))
	for _,p := range r {
		fm,e := joinf.LoadMetaDirect(d.Repo.Repo,p[1])
		if e!=nil { continue }
		el = append(el,Entry{fm.IsDir,p[0]})
	}
	return el
}
func (d *Directory) Create(name string,dir bool) (*Directory,*File,error) {
	d.Obj.Lock()
	defer d.Obj.Unlock()
	_,e := d.Res.Dir.Find(name)
	if e==nil { return nil,nil,syscall.EEXIST }
	id := uuid.NewV4().String()
	res := resource.NewResource(d.Repo.Repo,id)
	if dir {
		e = joinf.CreateDir(res)
	}else{
		e = joinf.CreateFile(res)
	}
	res.Dispose()
	if e!=nil { return nil,nil,e }
	e = d.Res.Dir.Insert(name,id)
	if e!=nil {
		d.Repo.Repo.Delete(id)
		return nil,nil,e
	}
	return d.load(id)
}



