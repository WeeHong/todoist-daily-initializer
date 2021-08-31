#!/bin/sh
ENVFILE=.env
if [[ ! -f "$ENVFILE" ]]; then
    echo "$ENVFILE does not exists."
    exit
fi

go build src/main.go \
  && zip function.zip main task.json .env \
  && aws lambda update-function-code \
      --function-name  todoist-initializer \
      --zip-file fileb://function.zip \
  && rm -rf function.zip main

echo "DONE"