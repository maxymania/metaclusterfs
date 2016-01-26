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

package fusebits

import "github.com/hanwen/go-fuse/fuse"
import "github.com/hanwen/go-fuse/fuse/nodefs"
import "github.com/maxymania/metaclusterfs/joinf"
import "github.com/maxymania/metaclusterfs/lockman"
import "github.com/maxymania/metaclusterfs/resource"
import "syscall"
import "os"
import "github.com/satori/go.uuid"

type Res func(s string)*resource.Resource

type FSMan struct{
	lockman.Lockman
	Dirs  Res
	Files Res
}

type MNode struct{
	nodefs.Node
	lm  *FSMan
	obj *lockman.ObjEntry
}
func NewMNode(lm *FSMan,id string) *MNode{
	return &MNode{nodefs.NewDefaultNode(),lm,lm.Load(id)}
}
func (m *MNode) OnForget() {
	m.obj.Decr()
}
func (m *MNode) Lookup(out *fuse.Attr, name string, context *fuse.Context) (*nodefs.Inode, fuse.Status) {
	m.obj.Lock()
	defer m.obj.Unlock()
	{
		c := m.Inode().GetChild(name)
		if c!=nil { return c,fuse.OK }
	}
	d,ok := m.obj.Obj.(*joinf.Directory)
	if !ok {
		return nil,fuse.ENOTDIR
	}
	id,e := d.Find(name)
	if e!=nil { return nil,fuse.ToStatus(e) }
	co := m.lm.Load(id)
	nin := &MNode{nodefs.NewDefaultNode(),m.lm,co}
	code := nin.GetAttr(out,nil,context)
	if code!=fuse.OK { return nil,code }
	_,isDir := co.Obj.(*joinf.Directory)
	return m.Inode().NewChild(name,isDir,nin),fuse.OK
}
func (m *MNode) GetAttr(out *fuse.Attr, file nodefs.File, context *fuse.Context) (code fuse.Status) {
	/*if file!=nil {
		return file.GetAttr(out)
	}*/
	reg,isReg := m.obj.Obj.(*joinf.JoinFile)
	if isReg {
		out.Mode = fuse.S_IFREG | 0666
		out.Size = uint64(reg.F.FLength)
	}else{
		out.Mode = fuse.S_IFDIR | 0777
	}
	return fuse.OK
}
func (m *MNode) makeKnot(name string,isDir bool) (*nodefs.Inode,*MNode,fuse.Status) {
	m.obj.Lock()
	defer m.obj.Unlock()
	d,ok := m.obj.Obj.(*joinf.Directory)
	if !ok {
		return nil,nil,fuse.ENOTDIR
	}
	{
		c := m.Inode().GetChild(name)
		if c!=nil { return nil,nil,fuse.Status(syscall.EEXIST) }
		id,_ := d.Find(name)
		if id!="" { return nil,nil,fuse.Status(syscall.EEXIST) }
	}
	id := uuid.NewV4().String()
	if isDir {
		r := m.lm.Dirs(id)
		e := joinf.CreateDir(r)
		if e!=nil { return nil,nil,fuse.ToStatus(e) }
		e = d.Insert(name,id)
		if e!=nil {
			r.Delete()
			return nil,nil,fuse.ToStatus(e)
		}
		r.Dispose()
	}else{
		r := m.lm.Files(id)
		e := joinf.CreateFile(r)
		if e!=nil { return nil,nil,fuse.ToStatus(e) }
		e = d.Insert(name,id)
		if e!=nil {
			r.Delete()
			return nil,nil,fuse.ToStatus(e)
		}
		r.Dispose()
	}
	co := m.lm.Load(id)
	nin := &MNode{nodefs.NewDefaultNode(),m.lm,co}
	return m.Inode().NewChild(name,isDir,nin),nin,fuse.OK
}
func (m *MNode) Mkdir(name string, mode uint32, context *fuse.Context) (newNode *nodefs.Inode, code fuse.Status) {
	newNode,_,code = m.makeKnot(name,true)
	return
}
func (m *MNode) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file nodefs.File, child *nodefs.Inode, code fuse.Status) {
	var nin *MNode
	child,nin,code = m.makeKnot(name,true)
	if code!=fuse.OK { return }
	file,code = nin.Open(flags,context)
	return
}
func (m *MNode) Open(flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	m.obj.Lock()
	defer m.obj.Unlock()
	j,ok := m.obj.Obj.(*joinf.JoinFile)
	if !ok { return nil,fuse.EINVAL }
	if (flags&uint32(os.O_TRUNC))!=0 {
		e := j.Truncate(0)
		if e!=nil { return nil,fuse.ToStatus(e) }
	}
	return &MFile{nodefs.NewDefaultFile(),m.obj,j},fuse.OK
}


