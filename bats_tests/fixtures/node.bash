export NODENAME="testnode"
export TAGS="testtag"

node_new() {
  $CMD node new "$NODENAME" --pairing-id "$PAIRING_ID" --pairing-key "$PAIRING_KEY"
}

node_new_offline() {
  $CMD node new "$NODENAME" --pairing-id "$PAIRING_ID" --pairing-key "$PAIRING_KEY" --offline
}

node_run() {
  $CMD node run --name "$NODENAME"
}

node_cert() {
  $CMD node cert --name "$NODENAME" --tags "$TAGS"
}

node_delete() {
  $CMD node delete --name "$NODENAME" --confirm-delete "this is just a test"
}

node_list() {
  $CMD node list
}

node_check_exists() {
  $CMD node list | grep -q "$NODENAME"
}

node_show() {
  $CMD node show --name "$NODENAME"
}
