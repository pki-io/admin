export SOURCE_PATH=$(pwd)
export ORG_DIR="test-org"
export CMD="$SOURCE_PATH/pki.io"

if [[ ! -x "$CMD" ]]; then
  echo "Can't find pki.io binary at $CMD. Did you run 'make build'?"
  exit 1
fi

init() {
  $CMD init $ORG_DIR
  cd $ORG_DIR
}

cleanup() {
  [ -r pki.io.conf ] && cd ..
  if [[ "$NO_CLEAN" -ne "1" ]]; then
    [ -d $ORG_DIR ] && rm -rf $ORG_DIR
  fi
}
