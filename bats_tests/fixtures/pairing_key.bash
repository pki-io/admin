pairing_key_new() {
  output=$($CMD pairing-key new --tags testtag)
  export PAIRING_ID=$(echo "$output" | awk '/Pairing ID/ { print $3 }')
  export PAIRING_KEY=$(echo "$output" | awk '/Pairing key/ { print $3 }')
}
