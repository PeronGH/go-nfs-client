package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/Cyberax/go-nfs-client/nfs4"
)

func main() {
	ctx := context.Background()

	hostname, _ := os.Hostname()

	localPort := flag.Int("port", 0, "Local port to use for NFS connection")
	server := flag.String("server", "", "NFS server to connect to")
	uid := flag.Int("uid", 0, "UID to use for operations")
	gid := flag.Int("gid", 0, "GID to use for operations")
	machineName := flag.String("machine", hostname, "Machine name to use for operations")
	flag.Parse()

	client, err := nfs4.NewNfsClient(ctx, *localPort, *server, nfs4.AuthParams{
		Uid:         uint32(*uid),
		Gid:         uint32(*gid),
		MachineName: *machineName,
	})

	if err != nil {
		panic(err)
	}

	defer client.Close()

	currentPath := "/"

	commands := map[string]CommandHandler{
		"cd": func(args []string) error {
			if len(args) != 1 {
				return errors.New("cd requires exactly one argument")
			}

			p := path.Join(currentPath, args[0])

			dir, err := client.GetFileInfo(p)
			if err != nil {
				return err
			}

			if !dir.IsDir {
				return errors.New("cd requires a directory")
			}

			currentPath = p
			return nil
		},

		"ls": func(args []string) error {
			entries, err := client.GetFileList(currentPath)
			if err != nil {
				return err
			}

			for _, entry := range entries {
				fmt.Println(entry)
			}

			return nil
		},
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s> ", currentPath)

		cmd, err := reader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		fields := strings.Fields(cmd)

		if len(fields) == 0 {
			continue
		}

		handler, ok := commands[fields[0]]
		if !ok {
			fmt.Printf("Unknown command: %s\n", fields[0])
			continue
		}

		err = handler(fields[1:])
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
	}

}

type CommandHandler func(args []string) error
