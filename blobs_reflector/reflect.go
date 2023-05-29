package blobs_reflector

import (
	"encoding/json"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/lbryio/lbry.go/v2/extras/errors"
	"github.com/lbryio/reflector.go/cmd"
	"github.com/lbryio/reflector.go/db"
	"github.com/lbryio/reflector.go/reflector"
	"github.com/lbryio/reflector.go/store"
	"github.com/sirupsen/logrus"

	"github.com/lbryio/ytsync/v5/util"
)

var dbHandle *db.SQL

func ReflectAndClean() error {
	err := reflectBlobs()
	if err != nil {
		return err
	}
	return util.CleanupLbrynet()
}

func loadConfig(path string) (cmd.Config, error) {
	var c cmd.Config

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, errors.Err("config file not found")
		}
		return c, err
	}

	err = json.Unmarshal(raw, &c)
	return c, err
}

func reflectBlobs() error {
	if util.IsBlobReflectionOff() {
		return nil
	}
	logrus.Infoln("reflecting blobs...")
	//make sure lbrynet is off
	running, err := util.IsLbrynetRunning()
	if err != nil {
		return err
	}
	if running {
		return errors.Prefix("cannot reflect blobs as the daemon is running", err)
	}
	logrus.SetLevel(logrus.InfoLevel)
	defer logrus.SetLevel(logrus.DebugLevel)
	ex, err := os.Executable()
	if err != nil {
		return errors.Err(err)
	}
	exPath := filepath.Dir(ex)
	config, err := loadConfig(exPath + "/prism_config.json")
	if err != nil {
		return errors.Err(err)
	}
	if dbHandle == nil {
		dbHandle = new(db.SQL)
		err = dbHandle.Connect(config.DBConn)
		if err != nil {
			return errors.Err(err)
		}
	}
	st := store.NewDBBackedStore(store.NewS3Store(config.AwsID, config.AwsSecret, config.BucketRegion, config.BucketName), dbHandle, false)

	uploadWorkers := 10
	uploader := reflector.NewUploader(dbHandle, st, uploadWorkers, false, false)
	usr, err := user.Current()
	if err != nil {
		return errors.Err(err)
	}
	blobsDir := usr.HomeDir + "/.lbrynet/blobfiles/"
	err = uploader.Upload(blobsDir)
	if err != nil {
		return errors.Err(err)
	}
	if uploader.GetSummary().Err > 0 {
		return errors.Err("not al blobs were reflected. Errors: %d", uploader.GetSummary().Err)
	}
	log.Println("done reflecting blobs")
	return nil
}
