load "fixtures/basics"
load "fixtures/ca"

@test "ca new" {
  init_init
  init
  run ca_new
  [ "$status" -eq 0 ]
  cleanup
}

@test "ca new dnscope" {
  init_init
  init
  run ca_new_dnscope
  [ "$status" -eq 0 ]
  cleanup
}

@test "ca list" {
  init_init
  init
  ca_new
  run ca_list
  [ "$status" -eq 0 ]
  cleanup
}

@test "ca exists" {
  init_init
  init
  ca_new
  run ca_check_exists "testca"
  [ "$status" -eq 0 ]
  cleanup
}

@test "ca delete" {
  init_init
  init
  ca_new
  run ca_delete
  [ "$status" -eq 0 ]
  cleanup
}

@test "ca check deleted" {
  init_init
  init
  ca_new
  ca_delete
  run ca_check_exists "testca"
  [ "$status" -eq 1 ]
  cleanup
}

@test "ca show" {
  init_init
  init
  ca_new
  run ca_show
  [ "$status" -eq 0 ]
  cleanup
}

@test "ca import public" {
  init_init
  init
  create_external_ca
  run ca_import_public
  [ "$status" -eq 0 ]
  cleanup
}

@test "ca import private" {
  init_init
  init
  create_external_ca
  run ca_import_private
  [ "$status" -eq 0 ]
  cleanup
}

@test "ca show import public" {
  init_init
  init
  create_external_ca
  ca_import_public
  run ca_show
  [ "$status" -eq 0 ]
  echo "$output" | grep -q "BEGIN CERTIFICATE"
  [ "$?" -eq 0 ]
  cleanup
}

@test "ca show import private" {
  init_init
  init
  create_external_ca
  ca_import_private
  run ca_show_private
  [ "$status" -eq 0 ]
  echo "$output" | grep -q "BEGIN RSA PRIVATE KEY"
  [ "$?" -eq 0 ]
  cleanup
}

@test "ca update import private" {
  init_init
  init
  create_external_ca
  ca_import_private
  run ca_show
  [ "$status" -eq 0 ]
  old_ca="$output"
  create_external_ca
  run ca_update_private
  [ "$status" -eq 0 ]
  run ca_show
  [ "$status" -eq 0 ]
  new_ca="$output"
  [ "$new_ca" != "$old_ca" ]
  cleanup
}

