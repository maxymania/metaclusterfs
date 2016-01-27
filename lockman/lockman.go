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

package lockman

import "sync"

type LDFunc func(s string) interface{}

type FreeFunc func(i interface{})

type ObjEntry struct{
	sync.Mutex
	l *Lockman
	Id string
	Obj interface{}
	rc uintptr
}
/*
 Works like the Incr() method, but it does not invoke the synchronization.
 */
func (o *ObjEntry) IncrUns() {
	o.rc++
}

// Reference count increment
func (o *ObjEntry) Incr() {
	o.l.Lock()
	o.rc++
	o.l.Unlock()
}

// Reference count decrement
func (o *ObjEntry) Decr() {
	o.l.Lock()
	o.rc--
	if o.rc==0 { o.l.remove(o.Id) }
	o.l.Unlock()
}

type Lockman struct{
	sync.Mutex
	M   map[string]*ObjEntry
	LD  LDFunc
	F   FreeFunc
}
func (l *Lockman) Init(){
	l.M = make(map[string]*ObjEntry)
}
func (l *Lockman) remove(s string) {
	o,ok := l.M[s]
	if !ok { return }
	l.F(o.Obj)
}
func (l *Lockman) Load(s string) *ObjEntry {
	l.Lock()
	defer l.Unlock()
	o,ok := l.M[s]
	if ok {
		o.rc++
		return o
	}
	o = new(ObjEntry)
	o.rc = 1
	o.l = l
	o.Id = s
	o.Obj = l.LD(s)
	l.M[s] = o
	return o
}


