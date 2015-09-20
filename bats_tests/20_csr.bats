load "fixtures/basics"
load "fixtures/ca"
load "fixtures/csr"

@test "csr new" {
  init_init
  init
  run csr_new
  [ "$status" -eq 0 ]
  cleanup
}

@test "csr new standalone" {
  init_init
  init
  run csr_new_standalone
  [ "$status" -eq 0 ]
  [ -r "$CSR_EXPORT_FILE" ]
  cleanup
}

@test "csr new dn" {
  init_init
  init
  run csr_new_dn
  [ "$status" -eq 0 ]
  cleanup
}

@test "csr list" {
  init_init
  init
  csr_new
  run csr_list
  [ "$status" -eq 0 ]
  [[ "$output" =~ "$CSR_NAME" ]]
  cleanup
}

@test "csr exists" {
  init_init
  init
  csr_new
  run csr_check_exists "$CSR_NAME"
  [ "$status" -eq 0 ]
  cleanup
}

@test "csr delete" {
  init_init
  init
  csr_new
  run csr_delete
  [ "$status" -eq 0 ]
  cleanup
}

@test "csr show" {
  init_init
  init
  csr_new
  run csr_show
  [ "$status" -eq 0 ]
  [[ "$output" =~ "$CSR_NAME" ]]
  cleanup
}

@test "csr show export" {
  init_init
  init
  csr_new
  run csr_show_export
  [ "$status" -eq 0 ]
  [ -r "$CSR_EXPORT_FILE" ]
  tar -xOzf "$CSR_EXPORT_FILE" "${CSR_NAME}-csr.pem" | grep -q "BEGIN CERTIFICATE REQUEST"
  [ "$?" -eq 0 ]
  cleanup
}

@test "csr show export private" {
  init_init
  init
  csr_new
  run csr_show_export_private
  [ "$status" -eq 0 ]
  [ -r "$CSR_EXPORT_FILE" ]
  tar -xOzf "$CSR_EXPORT_FILE" "${CSR_NAME}-csr.pem" | grep -q "BEGIN CERTIFICATE REQUEST"
  [ "$?" -eq 0 ]
  tar -xOzf "$CSR_EXPORT_FILE" "${CSR_NAME}-key.pem" | grep -q "PRIVATE KEY"
  [ "$?" -eq 0 ]
  cleanup
}

@test "csr sign" {
  init_init
  init
  ca_new
  create_external_csr
  csr_import
  run csr_sign
  [ "$status" -eq 0 ]
}

@test "csr sign standalone" {
  skip
  init_init
  init
  ca_new
  create_external_csr
  run csr_sign_standalone
  [ "$status" -eq 0 ]
  [ -r "$CSR_EXPORT_FILE" ]
  tar -xOzf "$CSR_EXPORT_FILE" "${CSR_NAME}-cert.pem" | grep -q "BEGIN CERTIFICATE"
  [ "$?" -eq 0 ]
  cleanup
}
