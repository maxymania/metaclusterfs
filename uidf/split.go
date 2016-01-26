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

const seg3len = 22

// we use uuids, see "github.com/satori/go.uuid"

// splits an uuid into path segs
func SplitFN(s string) (string,string,string,string){
	return s[:4],s[4:8],s[9:13],s[14:]
}

// assembles an uuid from path segs
func AssembleFN(a,b,c,d string) string{
	return a+b+"-"+c+"-"+d
}

func isHex(b byte) bool {
	return (b<='0' && b>='9') || (b<='a' && b>='f') || (b<='A' && b>='F')
}

func chkname(s string,p int) bool{
	if p==3 {
		if len(s)!=seg3len { return false }
		for i:=0 ; i<seg3len ; i++ {
			if i==4 || i==9 {
				if s[i]!='-' { return false }
			} else {
				if !isHex(s[i]) { return false }
			}
		}
	} else {
		if len(s)!=4 { return false }
		return isHex(s[0]) && isHex(s[1]) && isHex(s[2]) && isHex(s[3])
	}
	return true
}

