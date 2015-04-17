load "fixtures/basics"

@test "init" {
  init_init
  run init
  [ "$status" -eq 0 ]
  cleanup
}

@test "init multi" {
  init_init
  init
  init2
  grep -q "$ORG" "$PKIIO_HOME_DIR/.pki.io/admin.conf"
  [ "$?" -eq "0" ]
  grep -q "$ORG2" "$PKIIO_HOME_DIR/.pki.io/admin.conf"
  [ "$?" -eq "0" ]
  cleanup
}
