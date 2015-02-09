load "fixtures/basics"

@test "top help" {
  run $CMD --help
  [ "$status" -eq 0 ]
}
