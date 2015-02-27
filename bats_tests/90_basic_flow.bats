load "fixtures/basics"
load "fixtures/node"
load "fixtures/ca"
load "fixtures/org"
load "fixtures/pairing_key"

@test "basic flow" {
  # can't run init for some reason
  init_init
  init
  run pairing_key_new
  [ "$status" -eq 0 ]
  run ca_new
  [ "$status" -eq 0 ]
  run node_new
  [ "$status" -eq 0 ]
  run org_run
  [ "$status" -eq 0 ]
  run node_run
  [ "$status" -eq 0 ]
  cleanup
}

