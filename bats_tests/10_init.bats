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

@test "init duplicate" {
  init_init
  init
  mv "${PKIIO_LOCAL_DIR}/${ORG}" "${PKIIO_LOCAL_DIR}/{$ORG2}"
  run init
  [ "$status" -eq 1 ]
  echo "$output" | grep -q "name already exists"
  already_exists="$?"
  [ "$already_exists" -eq 0 ]
  cleanup
}
