load "fixtures/basics"

@test "version" {
  run $CMD --version
  [ "$status" -eq 0 ]
}
