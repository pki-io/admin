ca_new() {
  $CMD ca new testca --tags testtag
}

ca_new_dnscope() {
  $CMD ca new testca --tags testtag --dn-l lll --dn-st stst --dn-o ooo --dn-ou ouou --dn-c ccc --dn-street street --dn-postal postal
}

ca_list() {
  $CMD ca list
}

ca_check_exists() {
  $CMD ca list | grep -q "$1"
}

ca_delete() {
  $CMD ca delete testca --confirm-delete "this is just a test"
}

ca_show() {
  $CMD ca show testca
}
