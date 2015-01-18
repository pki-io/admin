load "fixtures/basics"
load "fixtures/ca"

@test "ca new" {
  init
  run ca_new
  [ "$status" -eq 0 ]
  cleanup
}
