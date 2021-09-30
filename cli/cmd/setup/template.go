package setup

var CloudformationTemplate = `{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Resources": {
    "MantilSetupRole": {
      "Type": "AWS::IAM::Role",
      "Properties": {
        "RoleName": "mantil-setup",
        "AssumeRolePolicyDocument": {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Principal": {
                "Service": "lambda.amazonaws.com"
              },
              "Action": "sts:AssumeRole"
            }
          ]
        },
        "Policies": [
          {
            "PolicyName": "mantil-setup",
            "PolicyDocument": {
              "Version": "2012-10-17",
              "Statement": [
                {
                  "Effect": "Allow",
                  "Action": "*",
                  "Resource": "*"
                }
              ]
            }
          }
        ]
      }
    },
    "MantilSetupLambda": {
      "Type": "AWS::Lambda::Function",
      "Properties": {
        "FunctionName": "mantil-setup",
        "Handler": "bootstrap",
        "Runtime": "provided.al2",
        "Timeout": 900,
        "MemorySize": 512,
        "Layers": [
          "arn:aws:lambda:eu-central-1:553035198032:layer:git-lambda2:8",
          "arn:aws:lambda:eu-central-1:477361877445:layer:terraform-lambda:1"
        ],
        "Code": {
          "S3Bucket": "mantil-downloads",
          "S3Key": "functions/latest/setup.zip"
        },
        "Role": {
          "Fn::GetAtt": [
            "MantilSetupRole",
            "Arn"
          ]
        }
      },
      "DependsOn": [
        "MantilSetupRole",
        "MantilSetupLambdaLogGroup"
      ]
    },
    "MantilSetupLambdaLogGroup": {
      "Type": "AWS::Logs::LogGroup",
      "Properties": {
        "LogGroupName": "/aws/lambda/mantil-setup",
        "RetentionInDays": 14
      }
    }
  }
}`
