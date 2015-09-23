pairing_key_new() {
  output=$($CMD pairing-key new --tags testtag)
  e="$?"
  export PAIRING_ID=$(echo "$output" | awk '/Id/ { print $4 }')
  export PAIRING_KEY=$(echo "$output" | awk '/Key/ { print $4 }')
  return "$e"
}
