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

package fusebits2

import "github.com/hanwen/go-fuse/fuse"
import "github.com/hanwen/go-fuse/fuse/nodefs"
import "github.com/maxymania/metaclusterfs/filerepo"

type FileNode struct{
	nodefs.Node
	File *filerepo.File
}
func (f *FileNode) Init(){
	f.Node = nodefs.NewDefaultNode()
}
func (f *FileNode) OnForget() {
	f.File.Dispose()
}
func (f *FileNode) GetAttr(out *fuse.Attr, file nodefs.File, context *fuse.Context) (code fuse.Status) {
	out.Mode = fuse.S_IFREG | 0666
	out.Size = uint64(f.File.Size())
	return fuse.OK
}
func (f *FileNode) Truncate(file nodefs.File, size uint64, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(f.File.Truncate(int64(size)))
}
func (f *FileNode) Open(flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	o := new(FileObj)
	o.Init()
	o.Data = f.File
	f.File.Obj.Incr()
	return o,fuse.OK
}

type FileObj struct{
	nodefs.File
	Data *filerepo.File
}
func (f *FileObj) Init() {
	f.File = nodefs.NewDefaultFile()
}
func (f *FileObj) Release() {
	f.Data.Dispose()
}
func (f *FileObj) Read(dest []byte, off int64) (fuse.ReadResult, fuse.Status) {
	n,e := f.Data.ReadAt(dest,off)
	return fuse.ReadResultData(dest[:n]),fuse.ToStatus(e)
}
func (f *FileObj) Write(data []byte, off int64) (written uint32, code fuse.Status) {
	n,e := f.Data.WriteAt(data,off)
	return uint32(n),fuse.ToStatus(e)
}
func (f *FileObj) Truncate(size uint64) fuse.Status {
	return fuse.ToStatus(f.Data.Truncate(int64(size)))
}



