node_new() {
  go run $SOURCE_PATH/*.go node new testnode --pairing-id "$PAIRING_ID" --pairing-key "$PAIRING_KEY"
}
