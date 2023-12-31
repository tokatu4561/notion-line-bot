import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as goLambda from "@aws-cdk/aws-lambda-go-alpha"
import { AttributeType, BillingMode, Table, TableEncryption } from 'aws-cdk-lib/aws-dynamodb'

// import * as sqs from 'aws-cdk-lib/aws-sqs';

export class NotionBotStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // The code that defines your stack goes here

    // example resource
    // const queue = new sqs.Queue(this, 'NotionBotQueue', {
    //   visibilityTimeout: cdk.Duration.seconds(300)
    // });

    // lambda use go
    const lambda = new goLambda.GoFunction(this, 'GoLambda', {
      entry: 'lambda/go/main.go',
    });

    // dynamodb
    const dynamoTable = new Table(this, 'NotionBotTable', {
      partitionKey: { name: 'id', type: AttributeType.STRING },
    });
  }
}
