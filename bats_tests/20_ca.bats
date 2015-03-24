load "fixtures/basics"
load "fixtures/ca"

@test "ca new" {
  init_init
  init
  run ca_new
  [ "$status" -eq 0 ]
  cleanup
}

@test "ca new dnscope" {
  init_init
  init
  run ca_new_dnscope
  [ "$status" -eq 0 ]
  cleanup
}
