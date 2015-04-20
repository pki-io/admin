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

@test "ca list" {
  init_init
  init
  ca_new
  run ca_list
  [ "$status" -eq 0 ]
  cleanup
}

@test "ca exists" {
  init_init
  init
  ca_new
  run ca_check_exists "testca"
  [ "$status" -eq 0 ]
  cleanup
}

@test "ca delete" {
  init_init
  init
  ca_new
  run ca_delete
  [ "$status" -eq 0 ]
  cleanup
}

@test "ca check deleted" {
  init_init
  init
  ca_new
  ca_delete
  run ca_check_exists "testca"
  [ "$status" -eq 1 ]
  cleanup
}

@test "ca show" {
  init_init
  init
  ca_new
  run ca_show
  [ "$status" -eq 0 ]
  cleanup
}
