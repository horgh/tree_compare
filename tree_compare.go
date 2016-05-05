/*
 * This program takes as input a root directory.
 *
 * For every file under this root, it generates a checksum.
 *
 * It outputs the checksum and the filename.
 *
 * You can then compare this with a run on another host to see what
 * differences there may be.
 *
 * The reason I have this is because I rsync a large directory between
 * two hosts. rsync's --checksum flag ends up leading to a timeout due to how
 * long it takes to generate the file list. I intend to run this offline
 * and then resolve whatever differences there are separately.
 */

package main

import (
	"bufio"
	//"crypto/sha256"
	"crypto/md5"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

func main() {
	dir := flag.String("dir", "", "Path to root directory to begin checks.")

	flag.Parse()

	if len(*dir) == 0 {
		log.Print("You must specify a directory.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	err := runChecks(*dir)
	if err != nil {
		log.Printf(err.Error())
		os.Exit(1)
	}
}

// runChecks finds and then computes a checksum and reports each file under
// the given directory.
func runChecks(dir string) error {
	// Find the files to check.
	files, err := findFiles(dir)
	if err != nil {
		return err
	}

	// Sort them.
	sort.Strings(files)

	// Compute and output checksums.
	err = computeAndOutputChecksums(files, dir)
	if err != nil {
		return err
	}

	return nil
}

// findFiles recursively descends a directory tree and collects all regular
// files.
func findFiles(file string) ([]string, error) {
	// Open the file.
	// If it is a regular file, record it.
	// If it is a directory, recursively find files.
	// Otherwise, skip it.

	fh, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("Unable to open: %s: %s", file, err.Error())
	}

	fi, err := fh.Stat()
	if err != nil {
		fh.Close()
		return nil, fmt.Errorf("Unable to stat: %s: %s", file, err.Error())
	}

	var files []string
	if fi.Mode().IsRegular() {
		files = append(files, file)
		fh.Close()
		return files, nil
	}

	if fi.IsDir() {
		names, err := fh.Readdirnames(0)
		if err != nil {
			fh.Close()
			return nil, fmt.Errorf("Unable to read directory files: %s: %s", file,
				err.Error())
		}
		fh.Close()

		for _, name := range names {
			absName := fmt.Sprintf("%s%c%s", file, os.PathSeparator, name)
			subFiles, err := findFiles(absName)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		}
		return files, nil
	}

	log.Printf("Ignoring non-regular and non-directory file: %s", file)
	return files, nil
}

// computeAndOutputChecksums computes a checksum for a file, and then
// outputs it along with its filename.
//
// Before outputting the filename, it strips the given root directory
// prefix.
func computeAndOutputChecksums(files []string, prefix string) error {
	for _, filename := range files {
		fh, err := os.Open(filename)
		if err != nil {
			return err
		}

		reader := bufio.NewReader(fh)

		hasher := md5.New()

		_, err = reader.WriteTo(hasher)
		if err != nil {
			fh.Close()
			return err
		}
		fh.Close()

		outputFilename := strings.TrimPrefix(filename, prefix)

		fmt.Printf("%s: %x\n", outputFilename, hasher.Sum(nil))
	}

	return nil
}
