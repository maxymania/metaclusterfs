# metaclusterfs
Meta Filesystem for Disk Clusters

```go
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/maxymania/metaclusterfs/filerepo"
	"github.com/maxymania/metaclusterfs/fusebits2"

	"github.com/maxymania/metaclusterfs/joinf"
	"github.com/maxymania/metaclusterfs/uidf"
	"github.com/maxymania/metaclusterfs/resource"
	"github.com/satori/go.uuid"
)

const FOLDER = "/path/to/partition"

func main() {
	store := &uidf.FS{FOLDER}
	{
		res := resource.NewResource(store,uuid.Nil.String())
    	joinf.CreateDir(res)
	    res.Dispose()
	}
	repo := new(filerepo.Repository)
	repo.Repo = store
	repo.Init()
	
	// Scans the arg list and sets up flags
	debug := flag.Bool("debug", false, "print debugging messages.")
	flag.Parse()
	if flag.NArg() < 1 {
		// TODO - where to get program name?
		fmt.Println("usage: main MOUNTPOINT BACKING-PREFIX")
		os.Exit(2)
	}

	mountPoint := flag.Arg(0)
	fmt.Println(mountPoint)
	root := new(fusebits2.DirNode)
	root.Init()
	root.Dir = repo.GetDir(uuid.Nil.String())

	conn := nodefs.NewFileSystemConnector(root, nil)
	fmt.Println("OK! Get Ready Now!")
	server, err := fuse.NewServer(conn.RawFS(), mountPoint, nil)
	if err != nil {
		fmt.Printf("Mount fail: %v\n", err)
		os.Exit(1)
	}
	server.SetDebug(*debug)
	fmt.Println("Mounted!")
	server.Serve()
}
```


