import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";
import * as goLambda from "@aws-cdk/aws-lambda-go-alpha";
import {
  AttributeType,
  BillingMode,
  Table,
  TableEncryption,
} from "aws-cdk-lib/aws-dynamodb";
import * as events from "aws-cdk-lib/aws-events";
import * as targets from "aws-cdk-lib/aws-events-targets";
import * as sqs from "aws-cdk-lib/aws-sqs";

export class NotionBotStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // lambda use go
    const lambda = new goLambda.GoFunction(this, "GoLambda", {
      entry: "app/notion-bot/main.go",
      timeout: cdk.Duration.seconds(30),
      functionName: "NotionBotFunction",
      environment: {
        LINE_CHANNEL_SECRET: process.env.LINE_CHANNEL_SECRET || "",
        LINE_CHANNEL_TOKEN: process.env.LINE_CHANNEL_TOKEN || "",
      },
    });

    // api gateway
    const api = new cdk.aws_apigateway.RestApi(this, "NotionBotApi", {
      restApiName: "NotionBotApi",
      description: "This service serves NotionBot",
    });
    api.root.addMethod(
      "POST",
      new cdk.aws_apigateway.LambdaIntegration(lambda)
    );

    // dynamodb
    const dynamoTable = new Table(this, "NotionBotTable", {
      tableName: "line-notion-keys",
      partitionKey: { name: "line_id", type: AttributeType.STRING },
      billingMode: BillingMode.PAY_PER_REQUEST, // Use on-demand billing mode
      encryption: TableEncryption.DEFAULT,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      pointInTimeRecovery: true,
    });

    // grant lambda to access dynamodb
    dynamoTable.grantReadWriteData(lambda);

    // lambda use go
    // notify notes on notion to line
    const notifyNoteLambda = new goLambda.GoFunction(this, "NotifyNoteLambda", {
      entry: "app/notify_note/main.go",
      timeout: cdk.Duration.seconds(30),
      functionName: "NotifyNoteFunction",
      environment: {
        LINE_CHANNEL_SECRET: process.env.LINE_CHANNEL_SECRET || "",
        LINE_CHANNEL_TOKEN: process.env.LINE_CHANNEL_TOKEN || "",
      },
    });

    // grant lambda to access dynamodb
    dynamoTable.grantReadWriteData(notifyNoteLambda);

    // execute lambda every 24 hour to notify notes on notion to line
    const rule = new events.Rule(this, "NotifyNoteRule", {
      schedule: events.Schedule.rate(cdk.Duration.hours(24)),
    });
    rule.addTarget(new targets.LambdaFunction(notifyNoteLambda));

    // const deadLetterQueue = new sqs.Queue(this, "DeadLetterQueue", {
    //   retentionPeriod: cdk.Duration.days(3),
    //   encryption: sqs.QueueEncryption.KMS_MANAGED,
    // });

    // const queue = new sqs.Queue(this, "NotionBotQueue", {
    //   visibilityTimeout: cdk.Duration.seconds(30), // lambda のタイムアウト時間に合わせる
    //   encryption: sqs.QueueEncryption.KMS_MANAGED, // default encryption
    //   deadLetterQueue: {
    //     maxReceiveCount: 3,
    //     queue: deadLetterQueue,
    //   },
    // });
  }
}
