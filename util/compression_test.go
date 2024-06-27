package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockChainDB(t *testing.T) {
	blockchainDbPath := "./tmp/blockchain.db"

	if _, err := os.Stat(blockchainDbPath); os.IsNotExist(err) {
		t.Fatalf("no directory found")
	}
	files, err := filepath.Glob(blockchainDbPath + "*")
	if err != nil {
		t.Fatalf(err.Error())
	}
	tarPath := strings.Replace(blockchainDbPath, "blockchain.db", "", -1) + "test" + ".tar"

	err = ZstdCompressFiles(files)
	if err != nil {
		t.Fatalf(err.Error())
	}

	compressedFileNames := make([]string, len(files))
	for i, file := range files {
		compressedFileNames[i] = file + ".zst"
	}

	err = CreateTarball(tarPath, compressedFileNames)
	if err != nil {
		t.Fatalf(err.Error())
	}
	errStrings := make([]string, 0)
	filesToDelete := append(files, compressedFileNames...)
	for _, f := range filesToDelete {
		err = os.Remove(f)
		if err != nil {
			errStrings = append(errStrings, err.Error())
		}
	}
	if len(errStrings) > 0 {
		t.Fatalf(strings.Join(errStrings, "\n"))
	}

	blockchainDbDir := strings.Replace(blockchainDbPath, "blockchain.db", "", -1)
	extractedFileNames, err := Untar(tarPath, blockchainDbDir)
	if err != nil {
		t.Fatalf(err.Error())
	}
	for _, name := range extractedFileNames {
		pathToCompressedFile := filepath.Join(blockchainDbDir, name)
		err = Unzstd(pathToCompressedFile, blockchainDbDir)
		if err != nil {
			t.Fatalf(err.Error())
		}

		err = os.Remove(pathToCompressedFile)
		if err != nil {
			t.Fatalf(err.Error())
		}
	}
	err = os.Remove(tarPath)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestIndividualFileCompression(t *testing.T) {
	// Create a temporary directory for the test files
	tempDir, err := os.MkdirTemp("", "test_compression")
	if err != nil {
		t.Fatalf("Failed to create a temporary directory: %s", err)
	}
	defer os.RemoveAll(tempDir)

	testFiles := []string{
		"testfile1.txt",
		"testfile2.txt",
		"testfile3.txt",
	}
	content := []string{
		"Test content 1",
		"Test content 2",
		"Test content 3",
	}

	fullPath := make([]string, len(testFiles))
	for i, fileName := range testFiles {
		filePath := filepath.Join(tempDir, fileName)
		if err := os.WriteFile(filePath, []byte(content[i]), 0644); err != nil {
			t.Fatalf("Failed to write to the test file: %s", err)
		}
		fullPath[i] = filePath
	}

	if err := ZstdCompressFiles(fullPath); err != nil {
		t.Fatalf("Failed to compress the file: %s", err)
	}

	for i, path := range fullPath {
		zstdFilePath := path + ".zst"
		decompressionPath := filepath.Join(tempDir, "decompressed")
		err := Unzstd(zstdFilePath, decompressionPath)
		if err != nil {
			t.Fatalf("Failed to decompress the file: %s", err)
		}
		decompressedContent, err := os.ReadFile(filepath.Join(decompressionPath, testFiles[i]))

		if err != nil {
			t.Fatalf("Failed to read the decompressed file: %s", err)
		}
		if !assert.Equal(t, content[i], string(decompressedContent)) {
			t.Fatalf("Decompressed content does not match original content.")
		}
	}
}
