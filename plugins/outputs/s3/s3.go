package s3

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/internal/config/aws"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	authUtil "github.intuit.com/cloud-runtime/cet-intuit-aws-auth-utils/pkg/client"
)

type S3 struct {
	Region    string `toml:"region"`
	AccessKey string `toml:"access_key"`
	SecretKey string `toml:"secret_key"`
	RoleARN   string `toml:"role_arn"`
	Profile   string `toml:"profile"`
	Filename  string `toml:"shared_credential_file"`
	Token     string `toml:"token"`
	Bucket    string `toml:"bucket"`
	Prefix    string `toml:"prefix"`
	Account   string `toml:"accountid"`
	Debug     bool   `toml:"debug"`
	CredURL   string `toml:"credential_url"`
	Expire    string
	Value     credentials.Value
	svc       *s3.S3

	serializer serializers.Serializer
}

type AuthService struct {
	DefaultRegion     string `json:"aws_default_region"`
	AccessKeyID       string `json:"aws_access_key_id"`
	SecretAccessKey   string `json:"aws_secret_access_key"`
	SessionToken      string `json:"aws_session_token"`
	SessionExpiration string `json:"aws_session_expiration"`
}

var sampleConfig = `
  ## Amazon Region
  #region = ""
  ## Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Assumed credentials via STS if role_arn is specified
  ## 2) explicit credentials from 'access_key' and 'secret_key'
  ## 3) shared profile from 'profile'
  ## 4) environment variables
  ## 5) shared credentials file
  ## 6) EC2 Instance Profile
  ## 7) Custom Credential Provider (TicketMaster)
  #access_key = ""
  #secret_key = ""
  #token = ""
  #role_arn = ""
  #profile = ""
  #shared_credential_file = ""
  credential_url = "https://ticketmaster.cet.a.intuit.com/oim"
  ## S3 Bucket must exist prior to starting telegraf
  ## In AWS the bucket is automatically created for you
  ## Bucket name = intu-oim-[prd|dev]-[LAST_DIGIT_OF_ACCOUNT_ID]-[REGION]
  #bucket = ""
  ## The prefix for the metric names
  prefix = "telegraf."
  ## The local AWS account id without dashes. Example: "123412341234"
  ## This is used in the S3 key prefix
  ## Full key prefix is <accountid>/<hour>/telegraf/<uuid>
  #accountid = ""
  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "wavefront"
  ## To write how many metrics written to S3
  debug = false
`

func (s *S3) SampleConfig() string {
	return sampleConfig
}

func (s *S3) Description() string {
	return "Configuration for AWS S3 output."
}

func (s *S3) Connect() error {
	credentialConfig := &internalaws.CredentialConfig{
		Region:    s.Region,
		AccessKey: s.AccessKey,
		SecretKey: s.SecretKey,
		RoleARN:   s.RoleARN,
		Profile:   s.Profile,
		Filename:  s.Filename,
		Token:     s.Token,
	}
	configProvider := credentialConfig.Credentials()
	svc := s3.New(configProvider)
	s.svc = svc

	s3Key, err := s.newS3Key()
	if err != nil {
		log.Printf("E! S3: s3Key error: %v\n", err)
	}

	testString := "Connect Test"
	acl := "bucket-owner-full-control"
	_, puterr := s.svc.PutObject(&s3.PutObjectInput{
		ACL:    &acl,
		Body:   strings.NewReader(string(testString)),
		Bucket: &s.Bucket,
		Key:    &s3Key,
	})

	if puterr != nil {
		log.Printf("W! S3: Error in Connect Test PutObject API call. Switching to Ticketmaster. Bucket: %+v Error: %+v \n", s.Bucket, puterr.Error())
		s.svc = nil

		_, err := s.Retrieve()
		if err != nil {
			log.Printf("E! S3: Ticket error: %v\n", err)
		}

		_, puterr := s.svc.PutObject(&s3.PutObjectInput{
			ACL:    &acl,
			Body:   strings.NewReader(string(testString)),
			Bucket: &s.Bucket,
			Key:    &s3Key,
		})
		if puterr != nil {
			log.Printf("E! S3: Error in TicketMaster Connect Test PutObject API call. Bucket: %+v Error: %+v \n", s.Bucket, puterr.Error())
		}
		return puterr

	}

	return puterr
}

func (s *S3) SetSerializer(serializer serializers.Serializer) {
	s.serializer = serializer
}

func (s *S3) Close() error {
	return nil
}

func (s *S3) Write(metrics []telegraf.Metric) error {
	count := 0
	lineSep := []byte{'\n'}
	if len(metrics) == 0 {
		return nil
	}

	var tmpValue []byte

	for _, m := range metrics {
		m.AddPrefix(s.Prefix)
		values, err := s.serializer.Serialize(m)

		if err != nil {
			return err
		}
		if s.Debug {
			count += bytes.Count(values, lineSep)
		}
		tmpValue = append(tmpValue, values...)
	}

	err := s.WriteToS3(tmpValue)
	if err != nil {
		return err
	}

	if s.Debug {
		log.Printf("D! S3: Wrote: %v metrics to S3\n", count)
	}
	return nil
}

func (s *S3) WriteToS3(data []byte) error {
	if len(s.Expire) > 1 {
		_, err := s.Retrieve()
		if err != nil {
			log.Printf("E! S3: Ticket error: %v\n", err)
		}
	}

	s3Key, err := s.newS3Key()
	if err != nil {
		log.Printf("E! S3: s3Key error: %v\n", err)
	}

	acl := "bucket-owner-full-control"
	_, puterr := s.svc.PutObject(&s3.PutObjectInput{
		ACL:    &acl,
		Body:   strings.NewReader(string(data)),
		Bucket: &s.Bucket,
		Key:    &s3Key,
	})

	if puterr != nil {
		log.Printf("E! S3: Unable to write metrics to S3 : %+v \n", puterr)
		s.Expire = ""

		_, err := s.Retrieve()
		if err != nil {
			log.Printf("E! S3: Backup Ticket error: %v\n", err)
		}

		_, puterr := s.svc.PutObject(&s3.PutObjectInput{
			ACL:    &acl,
			Body:   strings.NewReader(string(data)),
			Bucket: &s.Bucket,
			Key:    &s3Key,
		})
		if puterr != nil {
			log.Printf("E! S3: Error backup writing metrics to S3. Bucket: %+v Error: %+v \n", s.Bucket, puterr.Error())
		}
		return puterr

	}
	return puterr
}

func (s *S3) newS3Key() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40

	t := time.Now()
	h := t.Hour()
	account := strings.Replace(s.Account, "-", "", -1)
	if len(account) < 1 {
		account = "unset"
	}
	return fmt.Sprintf("%s/%d/%s/%x-%x-%x-%x-%x", account, h, "telegraf", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func (s *S3) Retrieve() (credentials.Value, error) {
	defaultReturn := credentials.Value{}

	//Check if credentials are expired. If yes, return
	if s.IsExpired() == false {
		return s.Value, nil
	}

	log.Printf("D! S3: nil or expired credentials. Renewing from: %s", s.CredURL)
	var authClient AuthService

	// Build blob
	authBlob, err := authUtil.BuildAuthBlob(&aws.Config{})
	if err != nil {
		return defaultReturn, err
	}

	// POST blob to TicketMaster service
	resp, err := http.Post(s.CredURL, "body/type", bytes.NewBuffer([]byte(authBlob)))
	if err != nil {
		return defaultReturn, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return defaultReturn, fmt.Errorf("Non-200 response %d", resp.StatusCode)
	}

	// Read body from TicketMaster service
	body_byte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return defaultReturn, err
	}

	// Unmarshal JSON response
	err = json.Unmarshal(body_byte, &authClient)
	if err != nil {
		return defaultReturn, err
	}

	val := credentials.Value{
		AccessKeyID:     authClient.AccessKeyID,
		SecretAccessKey: authClient.SecretAccessKey,
		SessionToken:    authClient.SessionToken,
		ProviderName:    "TicketMaster",
	}
	s.Value = val

	s.Expire = authClient.SessionExpiration

	credentialsRet := credentials.NewStaticCredentials(val.AccessKeyID, val.SecretAccessKey, val.SessionToken)

	configProvider := aws.NewConfig().WithRegion("us-west-2").WithCredentials(credentialsRet)
	svc := s3.New(session.New(), configProvider)
	s.svc = svc

	return val, nil
}

func (s *S3) IsExpired() bool {
	if s.Expire != "" {
		layout := "2006-01-02T15:04:05Z07:00"
		t, err := time.Parse(layout, s.Expire)
		if err != nil {
			log.Println(err)
		}

		minBeforeExpire := t.Add(-5 * time.Minute)
		now := time.Now()
		diff := now.Sub(minBeforeExpire)

		if now.Unix() < minBeforeExpire.Unix() {
			log.Printf("D! S3: Expire timestamp: %s Expires in: %s", s.Expire, diff)
			return false
		} else {
			log.Printf("D! S3: Expire timestamp: %s Expires in %s", s.Expire, diff)
			return true
		}
	}
	return true
}

func init() {
	outputs.Add("s3", func() telegraf.Output {
		return &S3{}
	})
}
