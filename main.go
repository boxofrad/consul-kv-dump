package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/consul-net-rpc/go-msgpack/codec"
	"github.com/hashicorp/consul/agent/consul/fsm"
	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/consul/snapshot"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("usage: %s /path/to/snapshot /path/to/output", os.Args[0])
	}

	input, err := os.Open(os.Args[1])
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(os.Args[2])
	if err != nil {
		return err
	}
	defer output.Close()

	readFile, _, err := snapshot.Read(hclog.NewNullLogger(), input)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(output)
	if err := writer.WriteByte('['); err != nil {
		return err
	}
	if err := fsm.ReadSnapshot(readFile, buildHandler(writer)); err != nil {
		return err
	}
	if _, err := writer.WriteString("]\n"); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}

type handler func(*fsm.SnapshotHeader, structs.MessageType, *codec.Decoder) error

func buildHandler(w io.Writer) handler {
	first := true

	return func(_ *fsm.SnapshotHeader, t structs.MessageType, d *codec.Decoder) error {
		if t != structs.KVSRequestType {
			var a any
			return d.Decode(&a)
		}

		var de structs.DirEntry
		if err := d.Decode(&de); err != nil {
			return err
		}

		if !first {
			if _, err := w.Write([]byte(",")); err != nil {
				return err
			}
		}

		b, err := json.Marshal(struct {
			Key   string
			Value []byte
		}{de.Key, de.Value})
		if err != nil {
			return err
		}

		if _, err := w.Write(b); err != nil {
			return err
		}

		first = false
		return nil
	}
}
