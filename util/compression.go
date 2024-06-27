package util

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"

	"github.com/lbryio/lbry.go/v2/extras/errors"
)

// ZstdCompressFiles individually compresses each file in the filePaths slice and saves them as .zst files.
func ZstdCompressFiles(filePaths []string) error {
	for _, filePath := range filePaths {
		zstdFilePath := filePath + ".zst"
		err := compressFile(filePath, zstdFilePath)
		if err != nil {
			return errors.Err("Could not compress file '%s', got error '%s'", filePath, err.Error())
		}
	}
	return nil
}

// compressFile creates a zstd compressed version of the specified file.
func compressFile(sourcePath, destinationPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return errors.Err("Could not open source file '%s', got error '%s'", sourcePath, err.Error())
	}
	defer inputFile.Close()

	outputFile, err := os.Create(destinationPath)
	if err != nil {
		return errors.Err("Could not create destination file '%s', got error '%s'", destinationPath, err.Error())
	}
	defer outputFile.Close()

	encoder, err := zstd.NewWriter(outputFile)
	if err != nil {
		return errors.Err("Could not create zstd encoder, got error '%s'", err.Error())
	}
	defer encoder.Close()

	_, err = io.Copy(encoder, inputFile)
	if err != nil {
		return errors.Err("Could not compress the file '%s', got error '%s'", sourcePath, err.Error())
	}

	return nil
}

// Unzstd decompresses a zstd file to a target directory.
func Unzstd(zstdFilePath, target string) error {
	file, err := os.Open(zstdFilePath)
	if err != nil {
		return errors.Err("Could not open zstd file '%s', got error '%s'", zstdFilePath, err.Error())
	}
	defer file.Close()

	decoder, err := zstd.NewReader(file)
	if err != nil {
		return errors.Err("Could not create zstd decoder, got error '%s'", err.Error())
	}
	defer decoder.Close()

	originalFileName := strings.TrimSuffix(filepath.Base(zstdFilePath), ".zst")
	err = os.MkdirAll(target, 0755)
	if err != nil {
		return errors.Err("Could not create target directory '%s', got error '%s'", target, err.Error())
	}
	extractedFile, err := os.Create(filepath.Join(target, originalFileName))
	if err != nil {
		return errors.Err("Could not create file for extraction '%s', got error '%s'", filepath.Join(target, originalFileName), err.Error())
	}
	defer extractedFile.Close()

	_, err = io.Copy(extractedFile, decoder)
	if err != nil {
		return errors.Err("Could not extract the zstd content, got error '%s'", err.Error())
	}

	return nil
}
