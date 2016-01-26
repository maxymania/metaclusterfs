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
import "github.com/maxymania/metaclusterfs/lockman"
import "github.com/maxymania/metaclusterfs/joinf"
import "io"

type MFile struct{
	nodefs.File
	obj  *lockman.ObjEntry
	file *joinf.JoinFile
}
func (m *MFile) Release() {
	m.obj.Decr()
}

func (m *MFile) Read(dest []byte, off int64) (fuse.ReadResult, fuse.Status) {
	m.obj.Lock()
	defer m.obj.Unlock()
	n,e := m.file.ReadAt(dest,off)
	if e==io.EOF { e = nil }
	return fuse.ReadResultData(dest[:n]),fuse.ToStatus(e)
}
func (m *MFile) Write(data []byte, off int64) (written uint32, code fuse.Status) {
	m.obj.Lock()
	defer m.obj.Unlock()
	r,e := m.file.WriteAt(data,off)
	return uint32(r),fuse.ToStatus(e)
}
func (m *MFile) Truncate(size uint64) fuse.Status {
	m.obj.Lock()
	defer m.obj.Unlock()
	return fuse.ToStatus(m.file.Truncate(int64(size)) )
}



