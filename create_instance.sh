#!/bin/bash

# Fail on error
set -e
# Parameters
instance="c7gn.xlarge"
stack_name="GraphInstanceStack"
port="20301"
bucket="s3graphtest10"
pem_file_path="~/Downloads/graphDbIreland.pem"

aws cloudformation create-stack \
    --stack-name $stack_name \
    --template-body file://./ec2_cf.yaml \
    --parameters ParameterKey=InstanceTypeParameter,ParameterValue=$instance \
    --capabilities CAPABILITY_IAM
aws cloudformation wait stack-create-complete --stack-name $stack_name
ip=$(aws cloudformation describe-stacks --stack-name $stack_name | jq -r .Stacks[0].Outputs[0].OutputValue)
echo "Created $instance instance with ip=$ip"

echo "Waiting for 10 seconds for instance initialization"
sleep 10

rm -f access
GOARCH=arm64 make access
# The strict host checking parameter is to trust new connections
scp -o StrictHostKeyChecking=accept-new -i $pem_file_path\
    ./access ubuntu@$ip:~/
ssh -o StrictHostKeyChecking=accept-new -i $pem_file_path\
    ubuntu@$ip "nohup ./access --port $port --bucket $bucket --nolog > server.log 2>&1 </dev/null &"
echo "Started server at $ip:$port"
