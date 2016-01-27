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

package uidf

import "os"
import "syscall"

type FS struct{
	Prefix string
}
func (fs *FS) of_id(id string) string {
	a,b,c,d := SplitFN(id)
	return fs.Prefix+"/"+a+"/"+b+"/"+c+"/"+d
}
func (fs *FS) Open(id,part string) (f IFile,e error) {
	fn := fs.of_id(id)+"/"+part
	f,e = os.OpenFile(fn,os.O_RDWR,0)
	return
}
func (fs *FS) Create(id,part string) (f IFile,e error) {
	dn := fs.of_id(id)
	e = os.MkdirAll(dn,0777)
	if e!=nil { return }
	fn := dn+"/"+part
	f,e = os.OpenFile(fn,os.O_RDWR|os.O_CREATE,0666)
	return
}
func (fs *FS) DeletePart(id,part string) error {
	fn := fs.of_id(id)+"/"+part
	return os.Remove(fn)
}
func (fs *FS) Delete(id string) error {
	//dn := fs.of_id(id)
	a,b,c,d := SplitFN(id)
	an := fs.Prefix+"/"+a
	bn := an+"/"+b
	cn := bn+"/"+c
	dn := cn+"/"+d
	e := os.RemoveAll(dn)
	if e==nil {
		syscall.Rmdir(cn)
		syscall.Rmdir(bn)
		syscall.Rmdir(an)
	}
	return e
}

func (fs *FS) ListUp(id chan <- string,q *bool) {
	close(id)
}
func (fs *FS) ListUpParts(id string,part chan <- string,q *bool) {
	close(part)
}

func (fs *FS) Info(s *IRepoStat) error {
	var stat syscall.Statfs_t
	e := syscall.Statfs(fs.Prefix,&stat)
	if e!=nil { return e }
	s.Unit = stat.Bsize
	s.TotalUnits = stat.Blocks

	// https://boostgsoc13.github.io/boost.afio/doc/html/afio/reference/structs/statfs_t.html
	// free blocks avail to non-superuser (Windows, POSIX) 
	s.FreeUnits = stat.Bavail
	return nil
}


