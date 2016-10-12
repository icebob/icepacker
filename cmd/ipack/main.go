package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli"

	"github.com/icebob/icepacker/src"
)

// GitCommit contains the commit hash. This will be filled in by the compiler.
var GitCommit string = "dev"

// Version is the version of CLI app.
const Version = "0.1.0"

func main() {
	app := cli.NewApp()
	app.Name = "ipack"
	app.Usage = "IcePacker - bundle your files securely"
	app.Version = Version + " (Git: " + GitCommit + ")"

	app.Commands = []cli.Command{
		{
			Name:  "pack",
			Usage: "Create a pack from `SOURCE DIR` to `TARGET_FILE`",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "key, k",
					Value: "",
					Usage: "Key for encrypting",
				},

				cli.StringFlag{
					Name:  "encrypt, e",
					Value: "none",
					Usage: "Type of encryption (none, aes)",
				},

				cli.StringFlag{
					Name:  "compress, c",
					Value: "none",
					Usage: "Type of compression (none, gzip)",
				},
			},
			Action: pack,
		},
		{
			Name:  "unpack",
			Usage: "Extract a `PACK FILE` to `TARGET DIR`",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "key, k",
					Value: "",
					Usage: "Key for decrypting if the file is encrypted",
				},
			},
			Action: unpack,
		},
		{
			Name:  "list",
			Usage: "List files from a `PACK FILE`",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "key, k",
					Value: "",
					Usage: "Key for decrypting if the file is encrypted",
				},
			},
			Action: list,
		},
	}

	app.Run(os.Args)
}

func pack(c *cli.Context) error {
	if len(c.Args()) < 2 {
		cli.ShowCommandHelp(c, "pack")
		return cli.NewExitError("Please set source directory and target filename", 2)
	}

	encryption := 0
	switch c.String("encrypt") {
	case "aes":
		encryption = icepacker.ENCRYPT_AES
		fmt.Println("Encryption: ", "AES128")
	}

	compression := 0
	switch c.String("compress") {
	case "gz", "gzip":
		compression = icepacker.COMPRESS_GZIP
		fmt.Println("Compression: ", "GZIP")
	}

	if encryption > 0 && c.String("key") == "" {
		return cli.NewExitError("Please set the encryption key with --key parameter", 1)
	}

	chanProgress := make(chan icepacker.ProgressState, 10)
	chanFinish := make(chan icepacker.FinishResult)

	start := time.Now()
	go icepacker.Pack(icepacker.PackSettings{
		SourceDir:      c.Args()[0],
		TargetFilename: c.Args()[1],
		Compression:    byte(compression),
		Encryption:     byte(encryption),
		Cipher:         icepacker.NewCipherSettings(c.String("key")),
		OnProgress:     chanProgress,
		OnFinish:       chanFinish,
	})

	done := false
	for {
		select {
		case state := <-chanProgress:
			if state.Err != nil {
				fmt.Printf("ERROR: %s (file: %s)\n", state.Err, state.CurrentFile)
			} else {
				icepacker.PrintProgress("Packing files", state.Index, state.Total)
			}
		case res := <-chanFinish:
			if res.Err != nil {
				return cli.NewExitError(fmt.Sprintf("%s", res.Err), 3)
			}

			elapsed := time.Since(start)
			fmt.Printf("\nPack size: %s\n", icepacker.FormatBytes(res.Size))
			fmt.Printf("File count: %d, skipped duplicate: %d (%s)\n", res.FileCount, res.DupCount, icepacker.FormatBytes(res.DupSize))
			fmt.Printf("Elapsed time: %s\n", elapsed)

			done = true
		}

		if done {
			break
		}
	}

	return nil
}

func unpack(c *cli.Context) error {
	if len(c.Args()) < 2 {
		cli.ShowCommandHelp(c, "unpack")
		return cli.NewExitError("Please set package filename and target directory", 2)
	}

	chanProgress := make(chan icepacker.ProgressState, 10)
	chanFinish := make(chan icepacker.FinishResult)

	start := time.Now()
	go icepacker.Unpack(icepacker.UnpackSettings{
		PackFileName: c.Args()[0],
		TargetDir:    c.Args()[1],
		Cipher:       icepacker.NewCipherSettings(c.String("key")),
		OnProgress:   chanProgress,
		OnFinish:     chanFinish,
	})

	done := false
	for {
		select {
		case state := <-chanProgress:
			if state.Err != nil {
				fmt.Printf("ERROR: %s (file: %s)\n", state.Err, state.CurrentFile)
			} else {
				icepacker.PrintProgress("Unpacking files", state.Index, state.Total)
			}
		case res := <-chanFinish:
			if res.Err != nil {
				return cli.NewExitError(fmt.Sprintf("%s", res.Err), 3)
			}

			elapsed := time.Since(start)
			fmt.Printf("\nTotal size: %s\n", icepacker.FormatBytes(res.Size))
			fmt.Printf("File count: %d\n", res.FileCount)
			fmt.Printf("Elapsed time: %s\n", elapsed)

			done = true
		}

		if done {
			break
		}
	}

	return nil
}

func list(c *cli.Context) error {
	if len(c.Args()) < 1 {
		cli.ShowCommandHelp(c, "list")
		return cli.NewExitError("Please set package filename", 2)
	}

	chanFinish := make(chan icepacker.ListResult)

	go icepacker.ListPack(icepacker.ListSettings{
		PackFileName: c.Args()[0],
		Cipher:       icepacker.NewCipherSettings(c.String("key")),
		OnFinish:     chanFinish,
	})

	res := <-chanFinish

	if res.Err != nil {
		return cli.NewExitError(fmt.Sprintf("%s", res.Err), 3)
	}

	fmt.Println("Files in package:")
	for _, item := range res.FAT.Items {
		fmt.Printf("  %s (%s)\n", filepath.FromSlash(item.Path), icepacker.FormatBytes(item.OrigSize))
	}

	fmt.Printf("\nFile count: %d\n", res.FAT.Count)
	fmt.Printf("Total size: %s\n", icepacker.FormatBytes(res.FAT.Size))

	return nil
}
