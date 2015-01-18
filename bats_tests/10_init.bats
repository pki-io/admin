load "fixtures/basics"

@test "init" {
  run init
  [ "$status" -eq 0 ]
  cleanup
}
