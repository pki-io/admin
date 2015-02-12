export NODENAME="testnode"
node_new() {
  $CMD node new "$NODENAME" --pairing-id "$PAIRING_ID" --pairing-key "$PAIRING_KEY"
}

node_new_offline() {
  $CMD node new "$NODENAME" --pairing-id "$PAIRING_ID" --pairing-key "$PAIRING_KEY" --offline
}

node_run() {
  $CMD node run --name "$NODENAME"
}
