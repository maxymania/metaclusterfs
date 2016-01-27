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

/*
 Replica implements a replication mechanism for the uidf storage abstraction
 layer.
 */
package replica

import "github.com/maxymania/metaclusterfs/uidf"
import "sort"
import "sync"
import "syscall"
import "time"

const intervall = time.Second*10

func NewMultiRepo(repos []uidf.IRepo,stdunit int64, quorum int) uidf.IRepo {
	mr := new(multiRepo)
	mr.list = make(repoList,len(repos))
	mr.stdunit = stdunit
	mr.quorum = quorum
	if mr.stdunit<(1<<9) { mr.stdunit = 1<<12 }
	if mr.quorum<1 { mr.quorum = 2 }
	for i,r := range repos {
		mr.list[i].repo = r
	}
	mr.lastsort = time.Now()
	mr.isort()
	return mr
}

type repoElem struct{
	size uint64
	max uint64
	repo uidf.IRepo
}

type repoList []repoElem
func(r repoList) Len() int { return len(r) }
func(r repoList) Less(i, j int) bool {
	return r[i].size<r[j].size
}
func(r repoList) Swap(i, j int) {
	t := r[i]
	r[i]=r[j]
	r[j]=t
}

type multiRepo struct{
	sync.Mutex
	list repoList
	stdunit int64
	quorum int
	lastsort time.Time
}

func (m *multiRepo) nsort() {
	m.Lock(); defer m.Unlock()
	ot := time.Now()
	if ot.Sub(m.lastsort) > intervall{
		m.lastsort = ot
		m.isort()
	}
}
func (m *multiRepo) isort() {
	var stat uidf.IRepoStat
	for i := range m.list {
		if m.list[i].repo.Info(&stat)==nil {
			if stat.Translate(m.stdunit)==nil {
				m.list[i].size = stat.TotalUnits
				m.list[i].max  = stat.FreeUnits
			}
		}
	}
	sort.Sort(m.list)
}


func (m *multiRepo) Open(id, part string) (uidf.IFile, error) {
	mf := new(multiFile)
	mf.files = make([]uidf.IFile,0,m.quorum)
	for _,rle := range m.list {
		f,e := rle.repo.Create(id,part)
		if e!=nil { continue }
		mf.files = append(mf.files,f)
	}
	mf.onOpen()
	return mf,nil
}


func (m *multiRepo) Create(id, part string) (uidf.IFile, error) {
	m.nsort()
	j := 0
	mf := new(multiFile)
	mf.files = make([]uidf.IFile,0,m.quorum)
	err := error(syscall.EIO)
	for _,rle := range m.list {
		if j==m.quorum { break } // quorum reached
		f,e := rle.repo.Create(id,part)
		if e!=nil { err = e; continue }
		mf.files = append(mf.files,f)
		j++
	}
	if j<m.quorum {
		mf.Close()
		return nil,err // Quorum cannot be reached!
	}
	mf.onOpen()
	return mf,nil
}


func (m *multiRepo) DeletePart(id, part string) error {
	for _,rle := range m.list {
		rle.repo.DeletePart(id,part)
	}
	return nil
}


func (m *multiRepo) Delete(id string) error {
	for _,rle := range m.list {
		rle.repo.Delete(id)
	}
	return nil
}


func (m *multiRepo) ListUp(id chan<- string, q *bool) {
	close(id)
}


func (m *multiRepo) ListUpParts(id string, part chan<- string, q *bool) {
	close(part)
}


func (m *multiRepo) Info(s *uidf.IRepoStat) error {
	return syscall.ENOSYS
}


