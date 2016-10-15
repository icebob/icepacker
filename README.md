# IcePacker

IcePacker is a bundler. Written in Go. Pack every files from a source directory to a bundle file.

[![GoDoc](https://godoc.org/github.com/icebob/icepacker/lib?status.svg)](https://godoc.org/github.com/icebob/icepacker/lib)
[![Go Report Card](https://goreportcard.com/badge/github.com/icebob/icepacker)](https://goreportcard.com/report/github.com/icebob/icepacker)

[![Build Status](https://travis-ci.org/icebob/icepacker.svg?branch=master)](https://travis-ci.org/icebob/icepacker)
[![Drone Build Status](https://drone.io/github.com/icebob/icepacker/status.png)](https://drone.io/github.com/icebob/icepacker/latest)

## Key features
* Include & exclude filters
* Support encryption with AES128
* Support compression with GZIP
* CLI usage or as a library
* bundle is concatenable after other file
* skip duplicates
* save & restore permission of files

### Install

```bash
go get -u github.com/icebob/icepacker
```

## CLI usage (with `icepacker` executable)

### Pack
Use the `icepacker pack` command to create a bundle file. You can compress and encrypt the bundle. 
For compression use the `--compress`or `-c` flag. The next parameter is the compression type. Currently icepacker supports `gzip`.
For encryption use the `--encrypt` or `-e`flag. The next parameter is the encryption type. Currently icepacker supports `aes`. In this case, you need to set your key with `--key` or `-k` flag.
> Note! The bundle doesn't contain the parent folder.

#### Examples
Create a `myproject.pack` bundle file from the content of the `myproject` folder:
```bash
icepacker pack ./myproject myproject.pack
```

Create an AES encrypted bundle file:
```bash
icepacker pack --encrypt aes --key SeCr3tKeY ./myproject myproject.pack
```

Create a GZIP compressed bundle file:
```bash
icepacker pack --compress gzip ./myproject myproject.pack
```

### Unpack
Use the `icepacker unpack` command to extract files from a bundle file. The unpacker can recognize that the bundle is encrypted or compressed. No need additional flags. But if the bundle is encrypted, you need to set the key with `--key` or `-k` flag.

#### Examples
Extract files from the `myproject.pack` bundle file to the `myproject` folder:
```bash
icepacker unpack myproject.pack ./myproject
```

Extract encrypted bundle file:
```bash
icepacker unpack --key SeCr3tKeY myproject.pack ./myproject
```

### List
Use the `icepacker list` command to list all files what the bundle contains. 

#### Examples
List files from the `myproject.pack` bundle file:
```bash
icepacker list myproject.pack
```

List files from an encrypted bundle file:
```bash
icepacker list --key SeCr3tKeY myproject.pack
```



## Library usage  
Used constants in lib:
```go
	ENCRYPT_NONE = 0
	ENCRYPT_AES  = 1

	COMPRESS_NONE = 0
	COMPRESS_GZIP = 1
```

### Pack
For packing, you need to create & load a `PackSettings` struct and pass to the `icepacker.Pack` func.
#### PackSettings structure
```go
type PackSettings struct {
	SourceDir      string
	TargetFilename string
	Includes       string
	Excludes       string
	Compression    byte
	Encryption     byte
	Cipher         CipherSettings
	OnProgress     chan ProgressState
	OnFinish       chan FinishResult
}
```
##### Description of fields
|Name|Required|Description|
-----|--------|--------------------------
`SourceDir`| yes | The source directory path. Should be **absolute** path.
`TargetFilename`| yes | The output bundle file. Should be **absolute** path.
`Includes`|  | Include filter. Use regex.
`Excludes`|  | Exclude filter. Use regex.
`Compression`|  | 0 - none, 1 - GZIP
`Encryption`|  | 0 - none, 1 - AES
`Cipher`|  | If use encryption, set a `CipherSettings` struct.
`OnProgress`|  | On progress chan. Use `ProgressState` struct 
`OnFinish`|  | On finish chan. User `FinishResult` struct

The return value is a `FinishResult` struct. Which contains error, count of files, size...etc. 

##### Example:
Simple encrypted packing which includes only `js` files except in the `node_modules` folders.
```go
res := icepacker.Pack(icepacker.PackSettings{
	SourceDir:      source,
	TargetFilename: target,
	Compression:    COMPRESSION_NONE,
	Encryption:     ENCRYPTION_AES,
	Cipher:         icepacker.NewCipherSettings("secretKey"),
	Includes:       ".js$",
	Excludes:       "node_modules\\/",
})
```

If you want to running `pack` in a go routine you need to set `OnProgress`and `OnFinish` channels.
##### Example:
Packing in a new go routine and show a progressbar on stdout.
```go
// Create channels
chanProgress := make(chan icepacker.ProgressState, 10)
chanFinish := make(chan icepacker.FinishResult)

// Start packing in a go routine
go icepacker.Pack(icepacker.PackSettings{
	SourceDir:      source,
	TargetFilename: target,
	Compression:    COMPRESSION_NONE,
	Encryption:     ENCRYPTION_AES,
	Cipher:         icepacker.NewCipherSettings("secretKey"),
	OnProgress:     chanProgress,
	OnFinish:       chanFinish,
})

// Wait for progress & finish
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
			fmt.Printf("%s", res.Err)
			return
		}

		fmt.Printf("\nPack size: %s\n", icepacker.FormatBytes(res.Size))
		fmt.Printf("File count: %d, skipped duplicate: %d (%s)\n", res.FileCount, res.DupCount, icepacker.FormatBytes(res.DupSize))
		fmt.Printf("Elapsed time: %s\n", elapsed)

		done = true
	}

	if done {
		break
	}
}
```


## TODO
* CLI: pack: if no output, put result to stdout
* CLI: unpack: if no input, read content from stdin
* CLI: unpack: if the target exists, with -a flag, append the result to the target file  
* Lib: resource reader methods. Open bundle, and get only one file as a slice (OpenPack, GetFile, ClosePack)

* Checksum calc & check
 
## License
icepacker is available under the [MIT license](https://tldrlegal.com/license/mit-license).

## Contact

Copyright (C) 2016 Icebob

[![@icebob](https://img.shields.io/badge/github-icebob-green.svg)](https://github.com/icebob) [![@icebob](https://img.shields.io/badge/twitter-Icebobcsi-blue.svg)](https://twitter.com/Icebobcsi)
