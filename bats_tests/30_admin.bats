load "fixtures/basics"
load "fixtures/admin"

@test "admin list" {
  init_init
  init
  run admin_list
  [ "$status" -eq 0 ]
  cleanup
}

@test "admin check exists" {
  init_init
  init
  run admin_check_exists "admin"
  [ "$status" -eq 0 ]
  cleanup
}

@test "admin invite" {
  init_init
  init
  run admin_invite
  [ "$status" -eq 0 ]
  cleanup
}

@test "admin new" {
  init_init
  init
  admin_invite
  run admin_new
  [ "$status" -eq 0 ]
  cleanup
}

@test "admin run" {
  init_init
  init
  admin_invite
  admin_new
  run admin_run
  [ "$status" -eq 0 ]
  cleanup
}

@test "admin complete" {
  init_init
  init
  admin_invite
  admin_new
  admin_run
  run admin_complete
  [ "$status" -eq 0 ]
  cleanup
}

@test "admin delete" {
  init_init
  init
  admin_invite
  admin_new
  admin_run
  admin_complete
  run admin_delete
  [ "$status" -eq 0 ]
  cleanup
}


@test "admin check deleted" {
  init_init
  init
  admin_invite
  admin_new
  admin_run
  admin_complete
  run admin_check_exists "admin2"
  [ "$status" -eq 0 ]
  admin_delete
  run admin_check_exists "admin2"
  [ "$status" -eq 1 ]
  cleanup
}

@test "admin show" {
  init_init
  init
  run admin_show
  [ "$status" -eq 0 ]
  cleanup
}
