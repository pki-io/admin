load "fixtures/basics"

@test "init" {
  init_init
  run init
  [ "$status" -eq 0 ]
  cleanup
}
