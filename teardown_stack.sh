echo "Triggering delete operation"
aws cloudformation delete-stack \
    --stack-name GraphInstanceStack
aws cloudformation wait stack-delete-complete --stack-name GraphInstanceStack
echo "Teardown complete"
