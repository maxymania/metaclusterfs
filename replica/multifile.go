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

package replica

import "github.com/maxymania/metaclusterfs/uidf"
import "io"
import "os"
import "syscall"

type multiFile struct{
	files []uidf.IFile
	pos int64
}
func (m *multiFile) onOpen() {} // To be implemented


func (m *multiFile) ReadAt(p []byte, off int64) (n int, err error) {
	for _,f := range m.files {
		n,err = f.ReadAt(p,off)
		if err!=nil && err!=io.EOF { continue }
		return
	}
	return
}


func (m *multiFile) WriteAt(p []byte, off int64) (n int, err error) {
	o := 0
	for _,f := range m.files {
		o,err = f.WriteAt(p,off)
		if n<o { n=o }
	}
	if n==len(p) { err = nil } else if err==nil { err = syscall.EIO }
	return
}


func (m *multiFile) Read(p []byte) (n int, err error) {
	n,err = m.ReadAt(p,m.pos)
	m.pos += int64(n)
	return
}


func (m *multiFile) Write(p []byte) (n int, err error) {
	n,err = m.WriteAt(p,m.pos)
	m.pos += int64(n)
	return
}


func (m *multiFile) Seek(offset int64, whence int) (int64, error) {
	s,e := m.Stat()
	if e!=nil { return m.pos,e }
	switch whence {
		case 0: m.pos = offset
		case 1: m.pos += offset
		case 2:	m.pos = s.Size() + offset
	}
	if m.pos<0 {
		m.pos = 0
	}else if m.pos>=s.Size() {
		m.pos = s.Size()
	}
	return m.pos,nil
}


func (m *multiFile) Close() (e error) {
	for _,f := range m.files {
		g := f.Close()
		if g!=nil { e = g }
	}
	return
}


func (m *multiFile) Stat() (os.FileInfo, error) {
	return m.files[0].Stat()
}


func (m *multiFile) Truncate(size int64) (e error) {
	for _,f := range m.files {
		g := f.Truncate(size)
		if g!=nil { e = g }
	}
	return
}



