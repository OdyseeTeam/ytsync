package manager

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lbryio/ytsync/v5/configs"
	"github.com/lbryio/ytsync/v5/util"

	"github.com/lbryio/lbry.go/v2/extras/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/sirupsen/logrus"
)

func (s *Sync) getS3Downloader(config *aws.Config) (*s3manager.Downloader, error) {
	s3Session, err := session.NewSession(config)
	if err != nil {
		return nil, errors.Prefix("error starting session", err)
	}
	downloader := s3manager.NewDownloader(s3Session)
	return downloader, nil
}

func (s *Sync) getS3Uploader(config *aws.Config) (*s3manager.Uploader, error) {
	s3Session, err := session.NewSession(config)
	if err != nil {
		return nil, errors.Prefix("error starting session", err)
	}
	uploader := s3manager.NewUploader(s3Session)
	return uploader, nil
}

func (s *Sync) downloadWallet() error {
	defaultWalletDir, defaultTempWalletDir, key, err := s.getWalletPaths()
	if err != nil {
		return errors.Err(err)
	}
	downloader, err := s.getS3Downloader(configs.Configuration.WalletS3Config.GetS3AWSConfig())
	if err != nil {
		return err
	}
	out, err := os.Create(defaultTempWalletDir)
	if err != nil {
		return errors.Prefix("error creating temp wallet", err)
	}
	defer out.Close()

	bytesWritten, err := downloader.Download(out, &s3.GetObjectInput{
		Bucket: aws.String(configs.Configuration.WalletS3Config.Bucket),
		Key:    key,
	})
	if err != nil {
		// Casting to the awserr.Error type will allow you to inspect the error
		// code returned by the service in code. The error code can be used
		// to switch on context specific functionality. In this case a context
		// specific error message is printed to the user based on the bucket
		// and key existing.
		//
		// For information on other S3 API error codes see:
		// http://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html
		if aerr, ok := err.(awserr.Error); ok {
			code := aerr.Code()
			if code == s3.ErrCodeNoSuchKey {
				return errors.Err("wallet not on S3")
			}
		}
		return err
	} else if bytesWritten == 0 {
		return errors.Err("zero bytes written")
	}

	err = os.Rename(defaultTempWalletDir, defaultWalletDir)
	if err != nil {
		return errors.Prefix("error replacing temp wallet for default wallet", err)
	}
	s.state.walletDownloaded = true
	return nil
}

func (s *Sync) downloadBlockchainDB() error {
	if util.IsRegTest() {
		return nil // tests fail if we re-use the same blockchain DB
	}
	defaultBDBPath, defaultTempBDBPath, key, err := s.getBlockchainDBPaths()
	if err != nil {
		return errors.Err(err)
	}
	files, err := filepath.Glob(defaultBDBPath + "*")
	if err != nil {
		return errors.Err(err)
	}
	for _, f := range files {
		err = os.Remove(f)
		if err != nil {
			return errors.Err(err)
		}
	}
	if s.DbChannelData.WipeDB {
		return nil
	}
	downloader, err := s.getS3Downloader(configs.Configuration.BlockchaindbS3Config.GetS3AWSConfig())
	if err != nil {
		return errors.Err(err)
	}
	out, err := os.Create(defaultTempBDBPath)
	if err != nil {
		return errors.Prefix("error creating temp blockchain DB file", err)
	}
	defer out.Close()
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			log.Errorf("error removing temp blockchain DB file: %v", err)
		}
	}(defaultTempBDBPath)

	bytesWritten, err := downloader.Download(out, &s3.GetObjectInput{
		Bucket: aws.String(configs.Configuration.BlockchaindbS3Config.Bucket),
		Key:    key,
	})
	if err != nil {
		// Casting to the awserr.Error type will allow you to inspect the error
		// code returned by the service in code. The error code can be used
		// to switch on context specific functionality. In this case a context
		// specific error message is printed to the user based on the bucket
		// and key existing.
		//
		// For information on other S3 API error codes see:
		// http://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html
		if aerr, ok := err.(awserr.Error); ok {
			code := aerr.Code()
			if code == s3.ErrCodeNoSuchKey {
				return nil // let ytsync sync the database by itself
			}
		}
		return errors.Err(err)
	} else if bytesWritten == 0 {
		return errors.Err("zero bytes written")
	}

	blockchainDbDir := strings.Replace(defaultBDBPath, "blockchain.db", "", -1)
	extractedFileNames, err := util.Untar(defaultTempBDBPath, blockchainDbDir)
	if err != nil {
		return errors.Prefix("error extracting blockchain.db files", err)
	}

	for _, name := range extractedFileNames {
		if !strings.HasSuffix(name, ".zst") {
			continue
		}
		pathToCompressedFile := filepath.Join(blockchainDbDir, name)
		err = util.Unzstd(pathToCompressedFile, blockchainDbDir)
		if err != nil {
			return errors.Prefix("error unzstding blockchain.db files", err)
		}

		err = os.Remove(pathToCompressedFile)
		if err != nil {
			return errors.Err("error removing %s: %s", pathToCompressedFile, err.Error())
		}
	}

	log.Printf("blockchain.db data downloaded and extracted to %s", blockchainDbDir)
	s.state.blockchainDbDownloaded = true
	return nil
}

func (s *Sync) getWalletPaths() (defaultWallet, tempWallet string, key *string, err error) {
	defaultWallet = os.Getenv("HOME") + "/.lbryum/wallets/default_wallet"
	tempWallet = os.Getenv("HOME") + "/.lbryum/wallets/tmp_wallet"
	key = aws.String("/wallets/" + s.DbChannelData.ChannelId)
	if util.IsRegTest() {
		defaultWallet = os.Getenv("HOME") + "/.lbryum_regtest/wallets/default_wallet"
		tempWallet = os.Getenv("HOME") + "/.lbryum_regtest/wallets/tmp_wallet"
		key = aws.String("/regtest/" + s.DbChannelData.ChannelId)
	}

	lbryumDir := os.Getenv("LBRYUM_DIR")
	if lbryumDir != "" {
		defaultWallet = lbryumDir + "/wallets/default_wallet"
		tempWallet = lbryumDir + "/wallets/tmp_wallet"
	}

	if _, err := os.Stat(defaultWallet); !os.IsNotExist(err) {
		return "", "", nil, errors.Err("default_wallet already exists")
	}
	return
}

func (s *Sync) getBlockchainDBPaths() (defaultDB, tempDB string, key *string, err error) {
	lbryumDir := os.Getenv("LBRYUM_DIR")
	if lbryumDir == "" {
		if util.IsRegTest() {
			lbryumDir = os.Getenv("HOME") + "/.lbryum_regtest"
		} else {
			lbryumDir = os.Getenv("HOME") + "/.lbryum"
		}
	}
	defaultDB = lbryumDir + "/lbc_mainnet/blockchain.db"
	tempDB = lbryumDir + "/lbc_mainnet/tmp_blockchain.tar"
	key = aws.String("/blockchain_dbs/" + s.DbChannelData.ChannelId + ".tar")
	if util.IsRegTest() {
		defaultDB = lbryumDir + "/lbc_regtest/blockchain.db"
		tempDB = lbryumDir + "/lbc_regtest/tmp_blockchain.tar"
		key = aws.String("/regtest_dbs/" + s.DbChannelData.ChannelId + ".tar")
	}
	return
}

func (s *Sync) uploadWallet() error {
	log.Println("uploading wallet to S3...")
	defaultWalletDir := util.GetDefaultWalletPath()
	key := aws.String("/wallets/" + s.DbChannelData.ChannelId)
	if util.IsRegTest() {
		key = aws.String("/regtest/" + s.DbChannelData.ChannelId)
	}

	if _, err := os.Stat(defaultWalletDir); os.IsNotExist(err) {
		return errors.Err("default_wallet does not exist")
	}

	uploader, err := s.getS3Uploader(configs.Configuration.WalletS3Config.GetS3AWSConfig())
	if err != nil {
		return err
	}

	file, err := os.Open(defaultWalletDir)
	if err != nil {
		return err
	}
	defer file.Close()

	start := time.Now()

	for time.Since(start) < 30*time.Minute {
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(configs.Configuration.WalletS3Config.Bucket),
			Key:    key,
			Body:   file,
		})
		if err != nil {
			log.Printf("error uploading wallet to S3, retrying in 30 seconds. %s", err.Error())
			time.Sleep(30 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		return errors.Prefix("there was a problem uploading the wallet to S3", errors.Err(err))
	}
	log.Println("wallet uploaded to S3")

	return os.Remove(defaultWalletDir)
}

func (s *Sync) uploadBlockchainDB() error {
	log.Println("uploading blockchain db to S3...")
	defaultBDBDir, _, key, err := s.getBlockchainDBPaths()
	if err != nil {
		return errors.Err(err)
	}

	if _, err := os.Stat(defaultBDBDir); os.IsNotExist(err) {
		//return errors.Err("blockchain.db does not exist")
		//
		return nil
	}
	files, err := filepath.Glob(defaultBDBDir + "*")
	if err != nil {
		return errors.Err(err)
	}
	tarPath := strings.Replace(defaultBDBDir, "blockchain.db", "", -1) + s.DbChannelData.ChannelId + ".tar"

	err = util.ZstdCompressFiles(files)
	if err != nil {
		return err
	}

	compressedFileNames := make([]string, len(files))

	for i, file := range files {
		compressedFileNames[i] = file + ".zst"
	}

	err = util.CreateTarball(tarPath, compressedFileNames)
	if err != nil {
		return err
	}

	uploader, err := s.getS3Uploader(configs.Configuration.BlockchaindbS3Config.GetS3AWSConfig())
	if err != nil {
		return err
	}

	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(configs.Configuration.BlockchaindbS3Config.Bucket),
		Key:    key,
		Body:   file,
	})
	if err != nil {
		return err
	}
	log.Println("blockchain.db files uploaded to S3")
	err = os.Remove(tarPath)
	if err != nil {
		return errors.Err(err)
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
		return errors.Err(strings.Join(errStrings, "\n"))
	}
	return nil
}
