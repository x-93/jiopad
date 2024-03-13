#!/bin/bash
rm -rf /tmp/karlsend-temp

NUM_CLIENTS=128
karlsend --devnet --appdir=/tmp/karlsend-temp --profile=6061 --rpcmaxwebsockets=$NUM_CLIENTS &
KARLSEND_PID=$!
KARLSEND_KILLED=0
function killKarlsendIfNotKilled() {
  if [ $KARLSEND_KILLED -eq 0 ]; then
    kill $KARLSEND_PID
  fi
}
trap "killKarlsendIfNotKilled" EXIT

sleep 1

rpc-idle-clients --devnet --profile=7000 -n=$NUM_CLIENTS
TEST_EXIT_CODE=$?

kill $KARLSEND_PID

wait $KARLSEND_PID
KARLSEND_EXIT_CODE=$?
KARLSEND_KILLED=1

echo "Exit code: $TEST_EXIT_CODE"
echo "Karlsend exit code: $KARLSEND_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $KARLSEND_EXIT_CODE -eq 0 ]; then
  echo "rpc-idle-clients test: PASSED"
  exit 0
fi
echo "rpc-idle-clients test: FAILED"
exit 1
