package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// PackageDiffChecker is a struct for comparing and identifying differences
// between files in two directories.
type PackageDiffChecker struct {
	SrcFiles []string // List of source files.
	SrcPath  string   // Source directory path.
	DstFiles []string // List of destination files.
	DstPath  string   // Destination directory path.
	SrcIsGno bool     // Indicates if the SrcFiles are gno files.
}

// Differences represents the differences between source and destination packages.
type Differences struct {
	SameNumberOfFiles bool             // Indicates whether the source and destination have the same number of files.
	FilesDifferences  []FileDifference // Differences in individual files.
}

// FileDifference represents the differences between source and destination files.
type FileDifference struct {
	Status          string            // Diff status of the processed files.
	SourceName      string            // Name of the source file.
	DestinationName string            // Name of the destination file.
	SrcLineDiff     []LineDifferrence // Differences in source file lines.
	DstLineDiff     []LineDifferrence // Differences in destination file lines.
}

// NewPackageDiffChecker creates a new PackageDiffChecker instance with the specified
// source and destination paths. It initializes the SrcFiles and DstFiles fields by
// listing files in the corresponding directories.
func NewPackageDiffChecker(srcPath, dstPath string, srcIsGno bool) (*PackageDiffChecker, error) {
	srcFiles, err := listDirFiles(srcPath)
	if err != nil {
		return nil, err
	}

	dstFiles, err := listDirFiles(dstPath)
	if err != nil {
		return nil, err
	}

	return &PackageDiffChecker{
		SrcFiles: srcFiles,
		SrcPath:  srcPath,
		DstFiles: dstFiles,
		DstPath:  dstPath,
		SrcIsGno: srcIsGno,
	}, nil
}

// Differences calculates and returns the differences between source and destination
// packages. It compares files line by line using the Myers algorithm.
func (p *PackageDiffChecker) Differences() (*Differences, error) {
	d := &Differences{
		SameNumberOfFiles: p.hasSameNumberOfFiles(),
		FilesDifferences:  make([]FileDifference, 0),
	}

	srcFilesExt, dstFileExt := p.inferFileExtensions()
	allFiles := p.listAllPossibleFiles()

	for _, trimmedFileName := range allFiles {
		srcFileName := trimmedFileName + srcFilesExt
		srcFilePath := p.SrcPath + "/" + srcFileName
		dstFileName := trimmedFileName + dstFileExt
		dstFilePath := p.DstPath + "/" + dstFileName

		fileDiff, err := NewFileDiff(srcFilePath, dstFilePath)
		if err != nil {
			return nil, err
		}

		srcDiff, dstDiff := fileDiff.Differences()

		d.FilesDifferences = append(d.FilesDifferences, FileDifference{
			Status:          p.getStatus(srcDiff, dstDiff).String(),
			SourceName:      srcFileName,
			DestinationName: dstFileName,
			SrcLineDiff:     srcDiff,
			DstLineDiff:     dstDiff,
		})
	}

	return d, nil
}

// listAllPossibleFiles returns a list of unique file names without extensions
// from both source and destination directories.
func (p *PackageDiffChecker) listAllPossibleFiles() []string {
	files := p.SrcFiles
	files = append(files, p.DstFiles...)

	for i := 0; i < len(files); i++ {
		files[i] = strings.TrimSuffix(files[i], ".go")
		files[i] = strings.TrimSuffix(files[i], ".gno")
	}

	unique := make(map[string]bool, len(files))
	uniqueFiles := make([]string, len(unique))
	for _, file := range files {
		if len(file) != 0 {
			if !unique[file] {
				uniqueFiles = append(uniqueFiles, file)
				unique[file] = true
			}
		}
	}

	return uniqueFiles
}

// inferFileExtensions by returning the src and dst files extensions.
func (p *PackageDiffChecker) inferFileExtensions() (string, string) {
	if p.SrcIsGno {
		return ".gno", ".go"
	}

	return ".go", ".gno"
}

// getStatus determines the diff status based on the differences in source and destination.
// It returns a diffStatus indicating whether there is no difference, missing in source, missing in destination, or differences exist.
func (p *PackageDiffChecker) getStatus(srcDiff, dstDiff []LineDifferrence) diffStatus {
	slicesAreEquals := slices.Equal(srcDiff, dstDiff)
	if slicesAreEquals {
		return noDiff
	}

	if len(srcDiff) == 0 {
		return missingInSrc
	}

	if len(dstDiff) == 0 {
		return missingInDst
	}

	if !slicesAreEquals {
		return hasDiff
	}

	return 0
}

// hasSameNumberOfFiles checks if the source and destination have the same number of files.
func (p *PackageDiffChecker) hasSameNumberOfFiles() bool {
	return len(p.SrcFiles) == len(p.DstFiles)
}

// listDirFiles returns a list of file names in the specified directory.
func listDirFiles(dirPath string) ([]string, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		return []string{}, nil
	}

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "can't close "+dirPath)
		}
	}()

	filesInfo, err := f.Readdir(0)
	if err != nil {
		return nil, fmt.Errorf("can't list file in directory :%w", err)
	}

	fileNames := make([]string, 0)
	for _, info := range filesInfo {
		fileNames = append(fileNames, info.Name())
	}

	return fileNames, nil
}
