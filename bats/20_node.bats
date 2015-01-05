load "fixtures/basics"
load "fixtures/node"

@test "node new" {
  init
  run node_new
  [ "$status" -eq 0 ]
  cleanup
}
