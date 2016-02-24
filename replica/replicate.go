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

import "io"

func replicate(dst io.WriterAt,src io.ReaderAt) (err error) {
	buf := make([]byte,1<<12)
	p := int64(0)
	for {
		n,e := src.ReadAt(buf,p)
		err = e
		if n>0 {
			dst.WriteAt(buf[:n],p)
		}
		if err!=nil { break }
	}
	return
}

