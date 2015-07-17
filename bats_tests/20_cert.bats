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

@test "cert new standalone" {
  init_init
  init
  run cert_new_standalone
  [ "$status" -eq 0 ]
  [ -r "$CERT_EXPORT_FILE" ]
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

@test "cert new ca standalone" {
  init_init
  init
  ca_new
  run cert_new_ca_standalone
  [ "$status" -eq 0 ]
  [ -r "$CERT_EXPORT_FILE" ]
  tar -xOzf "$CERT_EXPORT_FILE" "${CERT_NAME}-cert.pem" | grep -q "BEGIN CERTIFICATE"
  [ "$?" -eq 0 ]
  tar -xOzf "$CERT_EXPORT_FILE" "${CERT_NAME}-cacert.pem" | grep -q "BEGIN CERTIFICATE"
  [ "$?" -eq 0 ]
  tar -xOzf "$CERT_EXPORT_FILE" "${CERT_NAME}-key.pem" | grep -q "PRIVATE KEY"
  [ "$?" -eq 0 ]
  cleanup
}

@test "cert list" {
  init_init
  init
  cert_new
  run cert_list
  [ "$status" -eq 0 ]
  [[ "$output" =~ "$CERT_NAME" ]]
  cleanup
}

@test "cert exists" {
  init_init
  init
  cert_new
  run cert_check_exists "$CERT_NAME"
  [ "$status" -eq 0 ]
  cleanup
}

@test "cert delete" {
  init_init
  init
  cert_new
  run cert_delete
  [ "$status" -eq 0 ]
  cleanup
}

@test "cert show" {
  init_init
  init
  cert_new
  run cert_show
  [ "$status" -eq 0 ]
  [[ "$output" =~ "$CERT_NAME" ]]
  cleanup
}

@test "cert show export" {
  init_init
  init
  cert_new
  run cert_show_export
  [ "$status" -eq 0 ]
  [ -r "$CERT_EXPORT_FILE" ]
  tar -xOzf "$CERT_EXPORT_FILE" "${CERT_NAME}-cert.pem" | grep -q "BEGIN CERTIFICATE"
  [ "$?" -eq 0 ]
  cleanup
}

@test "cert show export private" {
  init_init
  init
  cert_new
  run cert_show_export_private
  [ "$status" -eq 0 ]
  [ -r "$CERT_EXPORT_FILE" ]
  tar -xOzf "$CERT_EXPORT_FILE" "${CERT_NAME}-cert.pem" | grep -q "BEGIN CERTIFICATE"
  [ "$?" -eq 0 ]
  tar -xOzf "$CERT_EXPORT_FILE" "${CERT_NAME}-key.pem" | grep -q "PRIVATE KEY"
  [ "$?" -eq 0 ]
  cleanup
}

@test "cert import public" {
  init_init
  init
  create_external_cert
  run cert_import_public
  [ "$status" -eq 0 ]
  cleanup
}

@test "cert show import public" {
  init_init
  init
  create_external_cert
  cert_import_public
  run cert_show
  [ "$status" -eq 0 ]
  echo "$output" | grep -q "BEGIN CERTIFICATE"
  [ "$?" -eq 0 ]
  cleanup
}

@test "cert import private" {
  init_init
  init
  create_external_cert
  run cert_import_private
  [ "$status" -eq 0 ]
  cleanup
}

@test "cert show import private" {
  init_init
  init
  create_external_cert
  cert_import_private
  run cert_show_private
  [ "$status" -eq 0 ]
  echo "$output" | grep -q "BEGIN RSA PRIVATE KEY"
  [ "$?" -eq 0 ]
  cleanup
}

