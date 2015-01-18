@test "version" {
  run go run *.go --version
  [ "$status" -eq 0 ]
}
