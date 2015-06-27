export CERT_NAME="testcert"
export CERT_EXPORT_FILE="${CERT_NAME}.tar.gz"
cert_new() {
        $CMD cert new testcert --export $CERT_EXPORT_FILE
}

cert_new_dn() {
    $CMD cert new $CERT_NAME --export $CERT_EXPORT_FILE --dn-l lll --dn-st stst --dn-o ooo --dn-ou ouou --dn-c ccc --dn-street street --dn-postal postal
}

cert_new_ca() {
    $CMD cert new $CERT_NAME --ca $CA_NAME --export $CERT_EXPORT_FILE
}
