package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"runtime/trace"
	"time"

	"github.com/icebob/icepacker/src"
)

const PPROF = false

func main() {
	fmt.Printf("\n\n\n")
	if PPROF {
		//defer profile.Start(profile.TraceProfile).Stop()
		f, _ := os.Create("r:\\cpu.prof")
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

		fTrace, _ := os.Create("r:\\trace.out")
		trace.Start(fTrace)
		defer trace.Stop()
	}

	os.MkdirAll("r:\\icepack-test", icepacker.DEFAULT_PERMISSION)

	TestPack("test1", icepacker.COMPRESS_NONE, icepacker.ENCRYPT_NONE, "icepack")
	TestPack("test2", icepacker.COMPRESS_NONE, icepacker.ENCRYPT_NONE, "icepack")
	//TestPack("test3", icepacker.COMPRESS_NONE, icepacker.ENCRYPT_NONE, "icepack")

	TestPack("filetest", icepacker.COMPRESS_GZIP, icepacker.ENCRYPT_NONE, "icepack")

}

func TestPack(name string, compression byte, encryption byte, key string) {

	fmt.Printf("\n=============== [ TEST: %s ] ==================\n", name)

	source := fmt.Sprintf("r:\\icepack-test\\%s", name)
	target := fmt.Sprintf("r:\\icepack-test\\%s.pack", name)

	if _, err := os.Stat(source); err == nil {
		os.Remove(target)
		fmt.Printf("Packing '%s'...\n", name)

		chanProgress := make(chan icepacker.ProgressState, 10)
		chanFinish := make(chan icepacker.FinishResult)

		start := time.Now()
		go icepacker.Pack(icepacker.PackSettings{
			SourceDir:      source,
			TargetFilename: target,
			Compression:    compression,
			Encryption:     encryption,
			Cipher:         icepacker.NewCipherSettings(key),
			//Includes:       "\\.txt$",
			//Excludes:       "dir1",
			OnProgress: chanProgress,
			OnFinish:   chanFinish,
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
					panic(res.Err)
				}

				fmt.Printf("\nPack size: %s\n", icepacker.FormatBytes(res.Size))
				fmt.Printf("File count: %d, skipped duplicate: %d (%s)\n", res.FileCount, res.DupCount, icepacker.FormatBytes(res.DupSize))

				done = true
			}

			if done {
				break
			}
		}

		elapsed := time.Since(start)
		fmt.Printf("\n-----------------[ Packed: %s, time: %s]-------------------\n\n", name, elapsed)
	} else {
		fmt.Printf("\n-----------------[ NO SOURCE DIR: %s]-------------------\n\n", name)
	}

	source = fmt.Sprintf("r:\\icepack-test\\%s.pack", name)
	target = fmt.Sprintf("r:\\icepack-test\\unpacked\\%s", name)
	os.RemoveAll(target)

	if _, err := os.Stat(source); err == nil {

		fmt.Printf("Unpacking '%s'...\n", name)

		chanProgress := make(chan icepacker.ProgressState, 10)
		chanFinish := make(chan icepacker.FinishResult)

		start := time.Now()
		go icepacker.Unpack(icepacker.UnpackSettings{
			PackFileName: source,
			TargetDir:    target,
			Cipher:       icepacker.NewCipherSettings(key),
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
					panic(res.Err)
				}

				fmt.Printf("\nTotal size: %s\n", icepacker.FormatBytes(res.Size))
				fmt.Printf("File count: %d\n", res.FileCount)

				done = true
			}

			if done {
				break
			}
		}

		elapsed := time.Since(start)
		fmt.Printf("=============== [ DONE: %s, time: %s ] ==================\n\n", name, elapsed)
	} else {
		fmt.Printf("=============== [ NO PACKED FILE: %s ] ==================\n\n", name)
	}
}
