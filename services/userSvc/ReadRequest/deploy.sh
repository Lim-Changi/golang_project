#!/bin/bash
export AWS_PAGER=''

PUBLISH=false
if [ "$#" -eq 1 ]; then
	param=$(tr [A-Z] [a-z] <<< "${1}")
  echo "${param}"
	if [ "${param}" == "publish" ]; then
		PUBLISH=true
		echo ">>>> Option publish: ${PUBLISH}"
	fi
fi
# ENV
Test='test'
Name='changi'
#TableName='pss-repo'
#GSIIndexName='SK-DATA-index'
#RecordTTLDays="100"
#UserPoolID="ap-northeast-2_966VKbg0r"
#StripeKey="rk_test_51L3bknFu7SsgJESTWjkD4YBWJn8rF19GG4BPDbvMwu1msIRSiFkNAEZ0gE4sHj6nrXpJKT3yz5zy4A1isKPgJdyM00tNTQQRnF"
#ApiProductID='default'
#ToolProductID='prod_M4rQWODLnyhmbI'

# common
svc_name='user'
svc_name_upper=$(tr [a-z] [A-Z] <<< "$svc_name")
func_name='ReadSvc'  
lambda_name="${svc_name_upper}-${func_name}"
role_name="arn:aws:iam::175343220571:role/lambda_exec"

LAMBDA_ENV=$(cat <<EOF
{"Variables":{\
"Test":"$Test", \
"Name":"$Name" \
}}
EOF
)

LambdaMemorySizeMB=512
LambdaTimeoutSec=10

# 주의: 빌드대상 경로를 상대경로로만 지정가능함
BUILD_TIME=`date -u +%Y-%m-%dT%H:%M:%SZ`
echo "BUILD_TIME=$BUILD_TIME"
REVISION=$(git rev-parse --short HEAD)
echo "REVISION=$REVISION"
VERSION=$(git describe --tags $(git rev-list --tags --max-count=1))
echo "VERSION=$VERSION"

BASEDIR=$(dirname $0)
#echo ${BASEDIR}
pushd ${BASEDIR} > /dev/null

OUT_FILE_PATH="/tmp/${svc_name}/${func_name}" 
mkdir -p /tmp/${svc_name}
env GOOS=linux GOARCH=amd64 go build \
	-ldflags "-X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME' -X 'main.Revision=$REVISION'" \
	-o $OUT_FILE_PATH .
re=$?
popd > /dev/null
if [ $re -ne 0 ]; then
	echo Build error.
	exit
fi
echo "Build output: $OUT_FILE_PATH"
if [ $PUBLISH = false ]; then
	echo "BUILD_ONLY!!! - ${lambda_name}"
	exit
fi

rm -f /tmp/${svc_name}/${func_name}.zip
zip -j /tmp/${svc_name}/${func_name}.zip $OUT_FILE_PATH
re=$?
if [ $re -ne 0 ]; then
	echo Make zip file error.
	exit
fi

aws lambda get-function --function-name ${lambda_name}
re=$?
if [ $re -eq 255 ] || [ $re -eq 254 ]; then
	echo ">>> try to create lambda <${lambda_name}>"
	aws lambda create-function --function-name ${lambda_name} --runtime go1.x \
		--role ${role_name} \
		--handler ${func_name} \
		--zip-file fileb:///tmp/${svc_name}/${func_name}.zip \
		--memory-size ${LambdaMemorySizeMB} \
		--timeout ${LambdaTimeoutSec} \
		--environment "${LAMBDA_ENV}"

elif [ $re -eq 0 ]; then
	echo ">>> try to update lambda <${lambda_name}>"
	aws lambda wait function-updated --function-name ${lambda_name}

	aws lambda update-function-code --function-name ${lambda_name} \
		--zip-file fileb:///tmp/${svc_name}/${func_name}.zip

	aws lambda wait function-updated --function-name ${lambda_name}

	aws lambda update-function-configuration --function-name ${lambda_name} \
		--role ${role_name} \
		--handler ${func_name} \
		--memory-size ${LambdaMemorySizeMB} \
		--timeout ${LambdaTimeoutSec} \
		--environment "${LAMBDA_ENV}"

	aws lambda wait function-updated --function-name ${lambda_name}

else
	echo ">> aws lambda get-function error. please check network connection"
	exit
fi


aws lambda invoke --function-name ${lambda_name} --invocation-type Event /dev/stdout 1> /dev/null 

