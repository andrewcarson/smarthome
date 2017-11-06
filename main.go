package main

import (
    "os/user"
    "strings"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sqs"

    "github.com/spf13/viper"

    log "github.com/sirupsen/logrus"
)

func initViper() {
    usr, err := user.Current()
    if (err != nil) {
        log.Fatal(err)
    }

    viper.AddConfigPath(".")
    viper.SetConfigName(usr.Username)
    err = viper.ReadInConfig()
    if (err != nil) {
        log.Info(err)
    }

    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    viper.AutomaticEnv()
}

func initLogrus() {
    log.SetFormatter(&log.JSONFormatter{TimestampFormat: time.RFC3339Nano})
}

func init() {
    initLogrus()
    initViper()
}

func createSqsService() (*sqs.SQS) {
    awsRegion := viper.GetString("aws.region")
    awsAccessKey := viper.GetString("aws.access_key")
    awsSecretKey := viper.GetString("aws.secret_key")
    log.Info(awsRegion)
    sess := session.Must(session.NewSession(&aws.Config{
        Region: &awsRegion,
        Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
    }))

    sqsService := sqs.New(sess)
    return sqsService
}

func main() {

    sqsService := createSqsService()
    queueUrl := viper.GetString("aws.queue_url")

    log.Println("Waiting...")
    result, err := sqsService.ReceiveMessage(&sqs.ReceiveMessageInput{
        AttributeNames: []*string{
            aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
        },
        MessageAttributeNames: []*string{
            aws.String(sqs.QueueAttributeNameAll),
        },
        QueueUrl:            &queueUrl,
        MaxNumberOfMessages: aws.Int64(1),
        VisibilityTimeout:   aws.Int64(3600),
        WaitTimeSeconds:     aws.Int64(viper.GetInt64("aws.sqs_wait_time")),
    })

    if err != nil {
        log.Warn("Error", err)
    }

    log.Infof("Received %d messages", len(result.Messages))

    if len(result.Messages) != 0 {
        resultDelete, err := sqsService.DeleteMessage(&sqs.DeleteMessageInput{
            QueueUrl:      &queueUrl,
            ReceiptHandle: result.Messages[0].ReceiptHandle,
        })

        if err != nil {
            log.Info("Delete Error", err)
            return
        }

        log.Info("Message Deleted", resultDelete)
    }
}
