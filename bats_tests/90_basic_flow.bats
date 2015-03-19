load "fixtures/basics"
load "fixtures/node"
load "fixtures/ca"
load "fixtures/org"
load "fixtures/pairing_key"

@test "basic flow" {
  # can't run init for some reason
  init_init
  init
  pairing_key_new
  ca_new
  node_new
  org_run
  run node_run
  [ "$status" -eq 0 ]
  cleanup
}

