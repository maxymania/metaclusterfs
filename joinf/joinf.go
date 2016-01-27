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
import "github.com/maxymania/metaclusterfs/uidf"
import "io"
import "encoding/gob"
import "bytes"
import "errors"
import "strconv"
import "syscall"

const MB64 = 1<<26

type FileMetadata struct{
	IsDir bool
	/* File related */
	FLength int64
	/* Directory related */
	Pages int64
}

func LoadMetaDirect(r uidf.IRepo,id string) (*FileMetadata,error) {
	m,e := r.Open(id,"meta")
	if e!=nil { return nil,e }
	f := new(FileMetadata)
	defer m.Close()
	dec := gob.NewDecoder(io.NewSectionReader(m,0,MB64))
	e = dec.Decode(f)
	return f,e
}

func LoadMeta(r *resource.Resource) (*FileMetadata,error) {
	m,e := r.Open("meta")
	if e!=nil { return nil,e }
	f := new(FileMetadata)
	defer m.Close()
	dec := gob.NewDecoder(io.NewSectionReader(m,0,MB64))
	e = dec.Decode(f)
	return f,e
}
func SaveMeta(r *resource.Resource,f *FileMetadata) error {
	if f==nil { return errors.New("BARF!") }
	m,e := r.Create("meta")
	if e!=nil { return e }
	defer m.Close()
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	e = enc.Encode(f)
	if e!=nil { return e }
	_,e = m.WriteAt(b.Bytes(),0)
	return e
}

type FileParts struct{
	Parts int64
}

func LoadParts(r *resource.Resource) (*FileParts,error) {
	m,e := r.Open("parts")
	if e!=nil { return nil,e }
	f := new(FileParts)
	defer m.Close()
	dec := gob.NewDecoder(io.NewSectionReader(m,0,MB64))
	e = dec.Decode(f)
	return f,e
}
func SaveParts(r *resource.Resource,f *FileParts) error {
	if f==nil { return errors.New("BARF!") }
	m,e := r.Create("parts")
	if e!=nil { return e }
	defer m.Close()
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	e = enc.Encode(f)
	if e!=nil { return e }
	_,e = m.WriteAt(b.Bytes(),0)
	return e
}

func cutdown(p *[]byte, off *int64, mark int64) (rp []byte,ro int64) {
	n := int(mark-*off)
	if n<len(*p) {
		rp = (*p)[:n]
		ro = *off
		*p = (*p)[n:]
		*off = mark
	} else {
		rp = *p
		ro = *off
		*p = nil
	}
	return
}

func CreateFile(r *resource.Resource) error {
	m,e := LoadMeta(r)
	if e!=nil { m = new(FileMetadata) }
	p,e := LoadParts(r)
	if e!=nil { p = new(FileParts) }
	e = SaveMeta(r,m)
	if e!=nil { return e }
	return SaveParts(r,p)
}

/*
 Loads an resource (file or directory) and returns a *JoinFile respectively a *Directory
 structure.
 */
func LoadResource(r *resource.Resource) (interface{},error) {
	f,e := LoadMeta(r)
	if e!=nil { return nil,e }
	if f.IsDir {
		return &Directory{r,f},nil
	} else {
		p,e := LoadParts(r)
		if e!=nil { return nil,e }
		return &JoinFile{r,f,p},nil
	}
}

type JoinFile struct{
	R *resource.Resource
	F *FileMetadata
	P *FileParts
}

func (j *JoinFile) Size() int64 {
	return j.F.FLength
}

func (j *JoinFile) readSeg(p []byte, off int64) (n int, err error) {
	index := off / MB64
	offset := off % MB64
	part := strconv.FormatInt(index,10)
	f,e := j.R.Open(part)
	if e!=nil { return 0,syscall.EIO }
	defer f.Close()
	n,err = f.ReadAt(p,offset)
	if err==io.EOF { err = syscall.EIO }
	return
}

func (j *JoinFile) ReadAt(p []byte, off int64) (n int, err error) {
	end := off + int64(len(p))
	if off>=j.F.FLength {
		return 0,io.EOF
	}
	if end>j.F.FLength {
		p = p[:int(j.F.FLength-off)]
	}
	n = len(p)
	for len(p)>0 {
		mark := (off - (off & MB64)) + MB64
		np,no := cutdown(&p,&off,mark)
		_,err = j.readSeg(np,no)
		if err!=nil { break }
	}
	return
}

func (j *JoinFile) writeSeg(p []byte, off int64) (n int, err error) {
	index := off / MB64
	offset := off % MB64
	part := strconv.FormatInt(index,10)
	f,e := j.R.Open(part)
	if e!=nil { return 0,syscall.EIO }
	defer f.Close()
	n,err = f.WriteAt(p,offset)
	if err==io.EOF { err = syscall.EIO }
	return
}

func (j *JoinFile) WriteAt(p []byte, off int64) (n int, err error) {
	N := len(p)
	end := off + int64(len(p))
	if end>j.F.FLength {
		err = j.Truncate(end)
		if err!=nil { return }
	}
	n = len(p)
	for len(p)>0 {
		mark := (off - (off & MB64)) + MB64
		np,no := cutdown(&p,&off,mark)
		_,err = j.writeSeg(np,no)
		if err!=nil { break }
	}
	if n<N && err==nil { err = syscall.EIO }
	return
}

func (j *JoinFile) grow(size int64) error {
	start := j.P.Parts
	last := size / MB64
	end := (size+MB64-1) / MB64
	if end>start {
		j.P.Parts = end
		e := SaveParts(j.R,j.P)
		if e!=nil { return e }
		if start>0 { // Fill up the last block
			part := strconv.FormatInt(start-1,10)
			f,e := j.R.Open(part)
			if e!=nil { return e }
			f.Truncate(MB64)
			f.Close()
		}
		for ; end>start ; start++ {
			part := strconv.FormatInt(start,10)
			f,e := j.R.Create(part)
			if e!=nil { return e }
			if start!=last {
				f.Truncate(MB64)
			}
			f.Close()
		}
	}else if end<start {
		for end<start {
			start--
			part := strconv.FormatInt(start,10)
			j.R.DeletePart(part)
		}
		j.P.Parts = end
		SaveParts(j.R,j.P)
	}
	return nil
}

func (j *JoinFile) Truncate(size int64) error {
	e := j.grow(size)
	if e!=nil { return e }
	if (size % MB64)>0 { // adjust the size of the last block
		part := strconv.FormatInt(size / MB64,10)
		f,e := j.R.Create(part)
		if e!=nil { return e }
		f.Truncate(size % MB64)
		f.Close()
	}
	j.F.FLength = size
	return SaveMeta(j.R,j.F)
}
// NewSectionReader



