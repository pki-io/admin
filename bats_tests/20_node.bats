load "fixtures/basics"
load "fixtures/node"
load "fixtures/org"
load "fixtures/pairing_key"

@test "node new" {
  init_init
  init
  pairing_key_new
  run node_new
  [ "$status" -eq 0 ]
  cleanup
}

@test "node list" {
  init_init
  init
  run node_list
  [ "$status" -eq 0 ]
  cleanup
}

@test "node run" {
  init_init
  init
  pairing_key_new
  node_new
  org_run
  run node_run
  [ "$status" -eq 0 ]
  cleanup
}

@test "node check exists" {
  init_init
  init
  pairing_key_new
  node_new
  org_run
  node_run
  run node_check_exists
  [ "$status" -eq 0 ]
  cleanup
}

@test "node check deleted" {
  init_init
  init
  pairing_key_new
  node_new
  org_run
  node_run
  node_delete
  run node_check_exists
  [ "$status" -eq 1 ]
  cleanup
}

@test "node show" {
  init_init
  init
  pairing_key_new
  node_new
  org_run
  node_run
  run node_show 
  [ "$status" -eq 0 ]
  cleanup
}
