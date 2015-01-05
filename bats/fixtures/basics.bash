[ -r Makefile ] || exit 1
export SOURCE_PATH=$(pwd)
export ORG_DIR="test-org"

init() {
  go run $SOURCE_PATH/*.go init $ORG_DIR
  cd $ORG_DIR
}

cleanup() {
  [ -r pki.io.conf ] && cd ..
  if [[ "$NO_CLEAN" -ne "1" ]]; then
    [ -d $ORG_DIR ] && rm -rf $ORG_DIR
  fi
}
