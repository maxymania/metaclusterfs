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


type File struct{
	Res *LocalRes
	Obj *lockman.ObjEntry
}
func (f *File) ReadAt(p []byte, off int64) (n int, err error) {
	f.Obj.Lock()
	defer f.Obj.Unlock()
	return f.Res.File.ReadAt(p,off)
}
func (f *File) WriteAt(p []byte, off int64) (n int, err error) {
	f.Obj.Lock()
	defer f.Obj.Unlock()
	return f.Res.File.WriteAt(p,off)
}
func (f *File) Dispose(){
	f.Obj.Decr()
}
func (f *File) Truncate(size int64) error {
	f.Obj.Lock()
	defer f.Obj.Unlock()
	return f.Res.File.Truncate(size)
}
func (f *File) Size() int64 {
	return f.Res.File.Size()
}


