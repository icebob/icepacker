# :package: IcePacker

IcePacker is a bundler. Written in Go. Pack every files from a source directory to a bundle file.

[![GoDoc](https://godoc.org/github.com/icebob/icepacker/lib?status.svg)](https://godoc.org/github.com/icebob/icepacker/lib)
[![Go Report Card](https://goreportcard.com/badge/github.com/icebob/icepacker)](https://goreportcard.com/report/github.com/icebob/icepacker)

[![Build Status](https://travis-ci.org/icebob/icepacker.svg?branch=master)](https://travis-ci.org/icebob/icepacker)
[![Drone Build Status](https://drone.io/github.com/icebob/icepacker/status.png)](https://drone.io/github.com/icebob/icepacker/latest)

## Key features
* Include & exclude filters
* Support encryption with AES128 with pbkdf2
* Support compression with GZIP
* CLI usage or as a library
* bundle is concatenable behind other file
* skip duplicated files (check by hash of content & size of file)
* save & restore permission of files

### Install

```bash
go get -u github.com/icebob/icepacker
```

## CLI usage (with `icepacker` executable)

### Pack
Use the `icepacker pack` command to create a bundle file. You can also compress and encrypt the bundle. 
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
You can use `icepacker` in your project as a library. In this case you need to import it as:
```go
import "github.com/icebob/icepacker/lib"
```

Used constants in project:
```go
	ENCRYPT_NONE = 0
	ENCRYPT_AES  = 1

	COMPRESS_NONE = 0
	COMPRESS_GZIP = 1
```
#### CipherSettings structure
The `CipherSettings` records the settings of encryption and hashing of key.
```go
type CipherSettings struct {
	Key       string
	Salt      string
	Iteration int
}
```
##### Description of fields
|Name|Required|Description|
-----|--------|--------------------------
`Key`| yes | The key of cipher.
`Salt`| yes | Salt for pbkdf2. Default: `icepacker`.
`Iteration`| yes | Count of iteration for pbkdf2. Default: 10000


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
`OnFinish`|  | On finish chan. Use `FinishResult` struct

The return value is a `FinishResult` struct. Which contains error, count of files, size...etc.


##### Example:
Simple encrypted packing which includes only `js` files except in the `node_modules` folders.
```go
res := icepacker.Pack(icepacker.PackSettings{
	SourceDir:      "/home/user/myfiles",
	TargetFilename: "/home/user/bundle.pack",
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
	SourceDir:      "/home/user/myfiles",
	TargetFilename: "/home/user/bundle.pack",
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

### Unpack
For unpacking, you need to create & load an `UnpackSettings` struct and pass to the `icepacker.Unpack` func.
#### UnpackSettings structure
```go
type UnpackSettings struct {
	PackFileName string
	TargetDir    string
	Includes     string
	Excludes     string
	Cipher       CipherSettings
	OnProgress   chan ProgressState
	OnFinish     chan FinishResult
}
```
##### Description of fields
|Name|Required|Description|
-----|--------|--------------------------
`PackFileName`| yes | The bundle file path. Should be **absolute** path.
`TargetDir`| yes | The output directory path. Should be **absolute** path.
`Includes`|  | Include filter. Use regex. > Currently not used
`Excludes`|  | Exclude filter. Use regex. > Currently not used
`Cipher`|  | If the bundle encrypted, set a `CipherSettings` struct.
`OnProgress`|  | On progress chan. Use `ProgressState` struct 
`OnFinish`|  | On finish chan. Use `FinishResult` struct

The return value is a `FinishResult` struct. Which contains error, count of files, size...etc.


##### Example:
Simple unpacking an encrypted bundle.
```go
res := icepacker.Pack(icepacker.PackSettings{
	PackFileName:   "/home/user/bundle.pack",
	TargetDir: 		"/home/user/myfiles",
	Cipher:         icepacker.NewCipherSettings("secretKey")
})
```

If you want to running `unpack` in a go routine you need to set `OnProgress`and `OnFinish` channels.
##### Example:
Unpacking in a new go routine and show a progressbar on stdout.
```go
// Create channels
chanProgress := make(chan icepacker.ProgressState, 10)
chanFinish := make(chan icepacker.FinishResult)

// Start packing in a go routine
go icepacker.Pack(icepacker.PackSettings{
	PackFileName:   "/home/user/bundle.pack",
	TargetDir: 		"/home/user/myfiles",
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
```

### List
If you only want to list files of the bundle, use the `icepacker.ListPack` method. You need to create & load an `ListSettings` struct and pass to the `icepacker.ListPack` func.
#### ListSettings structure
```go
type ListSettings struct {
	PackFileName string
	Cipher       CipherSettings
	OnFinish     chan ListResult
}
```
##### Description of fields
|Name|Required|Description|
-----|--------|--------------------------
`PackFileName`| yes | The bundle file path. Should be **absolute** path.
`Cipher`|  | If the bundle encrypted, set a `CipherSettings` struct.
`OnFinish`|  | On finish chan. Use `ListResult` struct

The return value is a `ListResult` struct. Which contains error and FAT.

##### Example:
Simple listing an encrypted bundle.

```go
res := icepacker.ListPack(icepacker.ListSettings{
	PackFileName:   "/home/user/bundle.pack",
	Cipher:         icepacker.NewCipherSettings("secretKey")
})
```

If you want to running `ListPack` in a go routine you need to set `OnFinish` channel.

##### Example:
Listing in a new go routine and show the result on stdout.
```go
// Create channels
chanFinish := make(chan icepacker.ListResult)

// Start listing in a go routine
go icepacker.ListPack(icepacker.ListSettings{
	PackFileName:   "/home/user/bundle.pack",
	Cipher:       icepacker.NewCipherSettings(c.String("key")),
	OnFinish:     chanFinish,
})

// Wait for finish
res := <-chanFinish

if res.Err != nil {
	return cli.NewExitError(fmt.Sprintf("%s", res.Err), 3)
}

fmt.Println("Files in package:")
for _, item := range res.FAT.Items {
	fmt.Printf("  %s (%s)\n", item.Path, icepacker.FormatBytes(item.OrigSize))
}

fmt.Printf("\nFile count: %d\n", res.FAT.Count)
fmt.Printf("Total size: %s\n", icepacker.FormatBytes(res.FAT.Size))
```


### Progress & Finish struct
These structs uses in `Pack`, `Unpack` and `ListPack` methods.

#### ProgressState struct

```go
type ProgressState struct {
	Err         error
	Total       int
	Index       int
	CurrentFile string
}
```
##### Description of fields
|Name|Description|
-----|--------------------------
`Err`| Contains an `error`if error occured. Otherwise `nil`.
`Total`| Count of files
`Index`| Index of current file (You can calculate percentage by `Index` and `Total`
`CurrentFile`| Path of the current file

#### FinishResult struct

```go
type FinishResult struct {
	Err       error
	FileCount int64
	Size      int64
	DupCount  int
	DupSize   int64
}
```
##### Description of fields
|Name|Description|
-----|--------------------------
`Err`| Contains an `error`if error occured. Otherwise `nil`.
`FileCount`| Count of files
`Size`| Size of the bundle
`DupCount`| Count of the skipped duplicated files
`DupSize`| Size of the skipped duplicated files

#### ListResult struct

```go
type ListResult struct {
	Err error
	FAT *FAT
}
```
##### Description of fields
|Name|Description|
-----|--------------------------
`Err`| Contains an `error`if error occured. Otherwise `nil`.
`FAT`| `FAT` (file list) of the bundle

## License
icepacker is available under the [MIT license](https://tldrlegal.com/license/mit-license).

## Contact

Copyright (C) 2016 Icebob

[![@icebob](https://img.shields.io/badge/github-icebob-green.svg)](https://github.com/icebob) [![@icebob](https://img.shields.io/badge/twitter-Icebobcsi-blue.svg)](https://twitter.com/Icebobcsi)
