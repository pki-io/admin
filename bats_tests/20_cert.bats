load "fixtures/basics"
load "fixtures/ca"
load "fixtures/cert"

@test "cert new" {
  init_init
  init
  run cert_new
  [ "$status" -eq 0 ]
  [ -r "$CERT_EXPORT_FILE" ]
  cleanup
}

@test "cert new dn" {
  init_init
  init
  run cert_new_dn
  [ "$status" -eq 0 ]
  [ -r "$CERT_EXPORT_FILE" ]
  cleanup
}

@test "cert new ca" {
  init_init
  init
  ca_new
  run cert_new_ca
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
