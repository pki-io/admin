load "fixtures/basics"
load "fixtures/ca"
load "fixtures/cert"

@test "cert new" {
  init_init
  init
  run cert_new
  [ "$status" -eq 0 ]
  cleanup
}

@test "cert new dn" {
  init_init
  init
  run cert_new_dn
  [ "$status" -eq 0 ]
  cleanup
}

@test "cert new ca" {
  init_init
  init
  ca_new
  run cert_new_ca
  [ "$status" -eq 0 ]
  cleanup
}
