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
import "syscall"

type BaseNode struct{
	nodefs.Node
}
func (b *BaseNode) Init(){
	b.Node = nodefs.NewDefaultNode()
}

type DirNode struct{
	nodefs.Node
	Dir *filerepo.Directory
}
func (b *DirNode) Init(){
	b.Node = nodefs.NewDefaultNode()
}
func (b *DirNode) OnForget() {
	b.Dir.Dispose()
}
func (b *DirNode) Lookup(out *fuse.Attr, name string, context *fuse.Context) (*nodefs.Inode, fuse.Status) {
	{
		c := b.Inode().GetChild(name)
		if c!=nil { return c,fuse.OK }
	}
	d,f,e := b.Dir.Lookup(name)
	if e!=nil { return nil,fuse.ToStatus(e) }
	if d!=nil {
		dn := &DirNode{nodefs.NewDefaultNode(),d}
		dn.GetAttr(out,nil,context)
		return b.Inode().NewChild(name,true,dn),fuse.OK
	}else if f!=nil{
		fn := &FileNode{nodefs.NewDefaultNode(),f}
		fn.GetAttr(out,nil,context)
		return b.Inode().NewChild(name,false,fn),fuse.OK
	}
	return nil,fuse.ToStatus(syscall.EIO)
}
func (b *DirNode) Mkdir(name string, mode uint32, context *fuse.Context) (newNode *nodefs.Inode, code fuse.Status) {
	d,f,e := b.Dir.Create(name,true)
	if e!=nil { return nil,fuse.ToStatus(e) }
	if d!=nil {
		dn := &DirNode{nodefs.NewDefaultNode(),d}
		return b.Inode().NewChild(name,true,dn),fuse.OK
	}else if f!=nil{
		fn := &FileNode{nodefs.NewDefaultNode(),f}
		return b.Inode().NewChild(name,false,fn),fuse.OK
	}
	return nil,fuse.ToStatus(syscall.EIO)
}
func (b *DirNode) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file nodefs.File, child *nodefs.Inode, code fuse.Status) {
	d,f,e := b.Dir.Lookup(name)
	if e!=nil {
		d,f,e = b.Dir.Create(name,false)
	}
	if e!=nil { return nil,nil,fuse.ToStatus(e) }
	if d!=nil {
		d.Dispose()
	}else if f!=nil{
		fn := &FileNode{nodefs.NewDefaultNode(),f}
		fobj,_ := fn.Open(flags,context)
		return fobj,b.Inode().NewChild(name,false,fn),fuse.OK
	}
	return nil,nil,fuse.ToStatus(syscall.EIO)
	panic("")
}
func (b *DirNode) OpenDir(context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	arr := b.Dir.Listup()
	ents := make([]fuse.DirEntry,len(arr))
	for i,e := range arr {
		if e.IsDir {
			ents[i].Mode = fuse.S_IFDIR | 0777
		}else{
			ents[i].Mode = fuse.S_IFREG | 0666
		}
		ents[i].Name = e.Name
	}
	return ents,fuse.OK
}
func (b *DirNode) GetAttr(out *fuse.Attr, file nodefs.File, context *fuse.Context) (code fuse.Status) {
	out.Mode = fuse.S_IFDIR | 0777
	return fuse.OK
}

