@test "top help" {
  run go run *.go --help
  [ "$status" -eq 0 ]
}
