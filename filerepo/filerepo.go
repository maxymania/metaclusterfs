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
import "github.com/maxymania/metaclusterfs/uidf"

type LocalRes struct{
	Res *resource.Resource
	File *joinf.JoinFile
	Dir *joinf.Directory
}

func free(i interface{}){
	lr := i.(*LocalRes)
	lr.Res.Dispose()
}

type Repository struct{
	lockman.Lockman
	Repo uidf.IRepo
}
func (r *Repository) Init() {
	r.Lockman.Init()
	r.LD = func(s string) interface{} {
		lr := new(LocalRes)
		lr.Res = resource.NewResource(r.Repo,s)
		f,e := joinf.LoadResource(lr.Res)
		if e==nil {
			switch v:=f.(type) {
				case *joinf.JoinFile:
					lr.File = v
				case *joinf.Directory:
					lr.Dir = v
			}
		}
		return lr
	}
	r.F = free
}
func (r *Repository) GetDir(id string) *Directory {
	obj := r.Load(id)
	lr := obj.Obj.(*LocalRes)
	if lr.Dir==nil {
		obj.Decr()
		return nil
	}
	return &Directory{r,lr,obj}
}



