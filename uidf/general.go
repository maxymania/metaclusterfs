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

import "io"
import "os"
import "errors"

var translateError = errors.New("odd unit size")

type IRepoStat struct{
	Unit int64
	TotalUnits uint64
	FreeUnits uint64
}
func (i *IRepoStat) Translate(unit int64) error {
	if i.Unit<unit {
		if (unit%i.Unit)!=0 { return translateError }
		n := unit/i.Unit
		i.Unit = unit
		i.TotalUnits /= uint64(n)
		i.FreeUnits /= uint64(n)
	}else if i.Unit>unit{
		if (i.Unit%unit)!=0 { return translateError }
		n := i.Unit/unit
		i.Unit = unit
		i.TotalUnits *= uint64(n)
		i.FreeUnits *= uint64(n)
	}
	return nil
}

type IFile interface{
	io.ReaderAt
	io.WriterAt
	io.ReadWriteSeeker
	io.Closer
	Stat() (os.FileInfo,error)
	Truncate(size int64) error
}

type IRepo interface{
	Open(id,part string) (IFile,error)
	Create(id,part string) (IFile,error)
	DeletePart(id,part string) error
	Delete(id string) error
	ListUp(id chan <- string,q *bool)
	ListUpParts(id string,part chan <- string,q *bool)
	Info(s *IRepoStat) error
}


