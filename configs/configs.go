package configs

import (
	"context"
	"net"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/lbryio/lbry.go/v2/extras/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	log "github.com/sirupsen/logrus"
	"github.com/tkanos/gonfig"
)

type S3Configs struct {
	ID       string `json:"id"`
	Secret   string `json:"secret"`
	Region   string `json:"region"`
	Bucket   string `json:"bucket"`
	Endpoint string `json:"endpoint"`
}
type Configs struct {
	SlackToken            string    `json:"slack_token"`
	SlackChannel          string    `json:"slack_channel"`
	InternalApisEndpoint  string    `json:"internal_apis_endpoint"`
	InternalApisAuthToken string    `json:"internal_apis_auth_token"`
	LbrycrdString         string    `json:"lbrycrd_string"`
	UseVpn                bool      `json:"use_vpn"`
	WalletS3Config        S3Configs `json:"wallet_s3_config"`
	BlockchaindbS3Config  S3Configs `json:"blockchaindb_s3_config"`
	ThumbnailsS3Config    S3Configs `json:"thumbnails_s3_config"`
}

var Configuration *Configs

func Init(configPath string) error {
	if Configuration != nil {
		return nil
	}
	c := Configs{}
	err := gonfig.GetConf(configPath, &c)
	if err != nil {
		return errors.Err(err)
	}
	Configuration = &c
	return nil
}

func dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	return dialer.DialContext(ctx, network, address)
}

func (s *S3Configs) GetS3AWSConfig() *aws.Config {
	c := &http.Client{
		Transport: &http.Transport{
			DialContext:           dialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	return &aws.Config{
		Credentials:      credentials.NewStaticCredentials(s.ID, s.Secret, ""),
		Region:           &s.Region,
		Endpoint:         &s.Endpoint,
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       c,
	}
}

func (c *Configs) GetHostname() string {
	var hostname string

	var err error
	hostname, err = os.Hostname()
	if err != nil {
		log.Error("could not detect system hostname")
		hostname = "ytsync_unknown"
	}
	reg, err := regexp.Compile("[^a-zA-Z0-9_]+")
	if err == nil {
		hostname = reg.ReplaceAllString(hostname, "_")

	}
	if len(hostname) > 30 {
		hostname = hostname[0:30]
	}
	return hostname
}
