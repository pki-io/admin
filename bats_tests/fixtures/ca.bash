ca_new() {
  $CMD ca new testca --tags testtag
}

ca_new_dnscope() {
  $CMD ca new testca --tags testtag --dn-l lll --dn-st stst --dn-o ooo --dn-ou ouou --dn-c ccc --dn-street street --dn-postal postal
}
