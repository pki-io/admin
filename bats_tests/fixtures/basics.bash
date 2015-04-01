export SOURCE_PATH=$(pwd)
export ORG="test-org"
export CMD="$SOURCE_PATH/pki.io"

if [[ ! -x "$CMD" ]]; then
  echo "Can't find pki.io binary at $CMD. Did you run 'make build'?"
  exit 1
fi

init_init() {
  export PKIIO_LOCAL_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'pkiiotmp')
  export PKIIO_HOME_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'pkiiotmp')
  export PKIIO_HOME2_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'pkiiotmp')
  export PKIIO_LOCAL="$PKIIO_LOCAL_DIR"
  export PKIIO_HOME="$PKIIO_HOME_DIR"

  echo "MAKING $PKIIO_LOCAL_DIR $PKIIO_HOME_DIR" >> /tmp/wtf.txt
}

init() {
  $CMD init $ORG
  cd "$PKIIO_LOCAL/$ORG"
  export PKIIO_LOCAL=""
}

cleanup() {
  #echo "CLEANING $PKIIO_LOCAL_DIR $PKIIO_HOME_DIR" >> /tmp/wtf.txt
  if [[ "$NO_CLEAN" -ne "1" ]]; then
    [ -d "$PKIIO_LOCAL_DIR" ] && rm -rf "$PKIIO_LOCAL_DIR"
    [ -d "$PKIIO_HOME_DIR" ] && rm -rf "$PKIIO_HOME_DIR"
    [ -d "$PKIIO_HOME2_DIR" ] && rm -rf "$PKIIO_HOME2_DIR"
  fi
  export PKIIO_HOME=""
  export PKIIO_LOCAL=""
  export PKIIO_HOME_DIR=""
  export PKIIO_HOME2_DIR=""
  export PKIIO_LOCAL_DIR=""
}
