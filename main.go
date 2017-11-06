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
        log.Fatalf("cannot get current user: %s", err)
    }

    viper.AddConfigPath(".")
    viper.SetConfigName(usr.Username)
    err = viper.ReadInConfig()
    if (err != nil) {
        log.Infof("cannot read config: %s", err)
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

    sess := session.Must(session.NewSession(&aws.Config{
        Region: &awsRegion,
        Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
    }))

    sqsService := sqs.New(sess)
    return sqsService
}


func deleteMessage(sqsService *sqs.SQS, queueUrl string, messageId *string, receiptHandle *string) {
    _, err := sqsService.DeleteMessage(&sqs.DeleteMessageInput{
        QueueUrl:      &queueUrl,
        ReceiptHandle: receiptHandle,
    })

    if err != nil {
        log.Infof("delete error for %s: %s", *messageId, err)
    } else {
        log.Infof("message deleted: %s", *messageId)
    }
}

func receiveMessages(sqsService *sqs.SQS, queueUrl string, messages chan string) {

    for {
        log.Info("waiting for message...")
        result, err := sqsService.ReceiveMessage(&sqs.ReceiveMessageInput{
            AttributeNames: []*string{
                aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
            },
            MessageAttributeNames: []*string{
                aws.String(sqs.QueueAttributeNameAll),
            },
            QueueUrl:            &queueUrl,
            MaxNumberOfMessages: aws.Int64(10),
            VisibilityTimeout:   aws.Int64(60),
            WaitTimeSeconds:     aws.Int64(viper.GetInt64("aws.sqs_wait_time")),
        })

        if err != nil {
            log.Warnf("failure in ReceiveMessage: %s", err)
            continue
        }

        log.Infof("received %d messages", len(result.Messages))
        for _, message := range result.Messages {
            messages <- *message.Body

            deleteMessage(sqsService, queueUrl, message.MessageId, message.ReceiptHandle)
        }
    }
}

func main() {

    sqsService := createSqsService()
    queueUrl := viper.GetString("aws.queue_url")
    messages := make(chan string)


    go receiveMessages(sqsService, queueUrl, messages)

    for message := range messages {
        log.Infof("message received: %s", message)
    }
}
