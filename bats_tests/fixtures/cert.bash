cert_new() {
        # Should export to std out and grep for files or something
        $CMD cert new testcert --export /dev/null
}

cert_new_dn() {
    $CMD cert new testcert --export /dev/null --dn-l lll --dn-st stst --dn-o ooo --dn-ou ouou --dn-c ccc --dn-street street --dn-postal postal
}

cert_new_ca() {
    $CMD cert new testcert --ca testca --export /dev/null
}
