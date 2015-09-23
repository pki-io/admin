export ADMINNAME="admin2"

admin_list() {
  $CMD admin list
}

admin_invite() {
  output=$($CMD admin invite "$ADMINNAME")
  e="$?"
  export INVITE_ID=$(echo "$output" | awk '/Id/ { print $4 }')
  export INVITE_KEY=$(echo "$output" | awk '/Key/ { print $4 }')
  return "$e"
}

admin_new() {
  PKIIO_HOME="$PKIIO_HOME2_DIR" $CMD admin new "$ADMINNAME" --invite-id "$INVITE_ID" --invite-key "$INVITE_KEY"
}

admin_run() {
  $CMD admin run
}

admin_complete() {
  PKIIO_HOME="$PKIIO_HOME2_DIR" $CMD admin complete "$ADMINNAME" --invite-id "$INVITE_ID" --invite-key "$INVITE_KEY"
}

admin_delete() {
  $CMD admin delete "$ADMINNAME" --confirm-delete "this is just a test"
}

admin_check_exists() {
  admin_list | grep -q "$1"
}

admin_show() {
  $CMD admin show admin
}
