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

package joinf

import "github.com/maxymania/metaclusterfs/resource"
import "io"
import "syscall"
import "strconv"
import "encoding/gob"
import "encoding/csv"
import "errors"
import "bytes"

type DirPage struct{
	Version int64
}
func LoadDPs(r *resource.Resource, part string) (*DirPage,error) {
	m,e := r.Open(part+"inf")
	if e!=nil { return nil,e }
	f := new(DirPage)
	defer m.Close()
	dec := gob.NewDecoder(io.NewSectionReader(m,0,MB64))
	e = dec.Decode(f)
	return f,e
}
func SaveDPs(r *resource.Resource,f *DirPage, part string) error {
	if f==nil { return errors.New("BARF!") }
	m,e := r.Create(part+"inf")
	if e!=nil { return e }
	defer m.Close()
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	e = enc.Encode(f)
	if e!=nil { return e }
	_,e = m.WriteAt(b.Bytes(),0)
	return e
}

const MAXDIR = 1<<10


func CreateDir(r *resource.Resource) error {
	m,e := LoadMeta(r)
	if e!=nil { m = new(FileMetadata) }
	m.IsDir = true
	return SaveMeta(r,m)
}

type Directory struct{
	R *resource.Resource
	F *FileMetadata
}
func (dir *Directory) load(i int64) ([][]string,error) {
	part := strconv.FormatInt(i,10)+"."
	p,e := LoadDPs(dir.R,part)
	if e!=nil { return nil,e }
	pver := part+strconv.FormatInt(p.Version,10)
	f,e := dir.R.Open(pver)
	if e!=nil { return nil,e }
	defer f.Close()
	s,e := f.Stat()
	if e!=nil { return nil,e }
	nr := csv.NewReader(io.NewSectionReader(f,0,s.Size()))
	return nr.ReadAll()
}
func (dir *Directory) save(i int64,l [][]string) error {
	b := new(bytes.Buffer)
	w := csv.NewWriter(b)
	w.WriteAll(l)
	w.Flush()

	// done!

	part := strconv.FormatInt(i,10)+"."
	p,e := LoadDPs(dir.R,part)
	if e!=nil { p = new(DirPage); e=nil }
	p.Version = ( p.Version + 1 ) % 8
	e = SaveDPs(dir.R,p,part)
	if e!=nil { return e }
	pver := part+strconv.FormatInt(p.Version,10)
	f,e := dir.R.Create(pver)
	if e!=nil { return e }
	defer f.Close()
	e = f.Truncate(int64(b.Len() ))
	if e!=nil { return e }
	_,e = f.WriteAt(b.Bytes(),0)
	return e
}
func (dir *Directory) find(i int64,s string) (string,error) {
	l,e := dir.load(i)
	if e!=nil { return "",e }
	for _,row := range l {
		if len(row)<2 { continue }
		if row[0]==s { return row[1],nil }
	}
	return "",nil
}
func (dir *Directory) Find(s string) (string,error) {
	for i := int64(0) ; i<dir.F.Pages ; i++ {
		s,e := dir.find(i,s)
		if e!=nil { return "",e }
		if s!="" { return s,nil }
	}
	return "",syscall.ENOENT
}
func (dir *Directory) delete(i int64,s string) (string,error) {
	l,e := dir.load(i)
	if e!=nil { return "",e }
	for j,row := range l {
		if len(row)<2 { continue }
		if row[0]==s {
			r := row[1]
			copy(l[j:],l[j+1:])
			l = l[:len(l)-1]
			dir.save(i,l)
			return r,nil
		}
	}
	return "",nil
}
func (dir *Directory) Delete(s string) (string,error) {
	for i := int64(0) ; i<dir.F.Pages ; i++ {
		s,e := dir.delete(i,s)
		if e!=nil { return "",e }
		if s!="" { return s,nil }
	}
	return "",syscall.ENOENT
}
func (dir *Directory) Insert(s, id string) error {
	for i := int64(0) ; i<dir.F.Pages ; i++ {
		println(i,dir.F.Pages)
		l,_ := dir.load(i)
		if len(l)>=MAXDIR { continue }
		l = append(l,[]string{s,id})
		return dir.save(i,l)
	}
	dir.F.Pages++
	e := SaveMeta(dir.R,dir.F)
	if e!=nil { return e }
	return dir.save(dir.F.Pages-1,[][]string{[]string{s,id}})
}


