load "fixtures/basics"
load "fixtures/node"
load "fixtures/pairing_key"

@test "node new" {
  init
  pairing_key_new
  run node_new
  [ "$status" -eq 0 ]
  cleanup
}
