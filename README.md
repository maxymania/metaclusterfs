# metaclusterfs
Meta Filesystem for Disk Clusters

```go
package main

import "fmt"
import "github.com/satori/go.uuid"
import "github.com/maxymania/metaclusterfs/uidf"
import "github.com/maxymania/metaclusterfs/resource"
import "github.com/maxymania/metaclusterfs/joinf"
import _ "github.com/maxymania/metaclusterfs/lockman"
import "github.com/maxymania/metaclusterfs/fusebits"
import "github.com/hanwen/go-fuse/fuse/nodefs"
import "flag"

var store uidf.IRepo

var fs = new(fusebits.FSMan)

func Res(s string) *resource.Resource {
	return resource.NewResource(store,s)
}

func LDFunc(s string) interface{} {
	r := Res(s)
	f,e := joinf.LoadResource(r)
	if e!=nil { return nil }
	return f
}

func FreeFunc(i interface{}) {
	switch v := i.(type) {
	case *joinf.JoinFile:
		v.R.Dispose()
	case *joinf.Directory:
		v.R.Dispose()
	}
}

const dr = "/path/to/partition"

func garbage() {
	fmt.Println()
}

func main(){
	flag.Parse()
	store = &uidf.FS{dr}
	res := resource.NewResource(store,uuid.Nil.String())
	joinf.CreateDir(res)
	res.Dispose()


	fs.Init()
	fs.LD = LDFunc
	fs.F = FreeFunc
	fs.Dirs = Res
	fs.Files = Res
	rnode := fusebits.NewMNode(fs,uuid.Nil.String())
	
	server,_,err := nodefs.MountRoot(flag.Arg(0),rnode, nil)
	if err != nil {
		fmt.Println("Mount fail:", err)
		return
	}
	server.SetDebug(true)
	server.Serve()
}


```
