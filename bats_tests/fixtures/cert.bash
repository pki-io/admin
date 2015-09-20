export CERT_NAME="testcert"
export CERT_EXPORT_FILE="${CERT_NAME}.tar.gz"
export CERT_TAG="testtag"
export CERT_EXTERNAL_CERT_NAME="external-cert"

create_external_cert() {
  openssl genrsa -out ${CERT_EXTERNAL_CERT_NAME}-key.pem 2048 >/dev/null 2>&1
  openssl req -new -batch -x509 -key ${CERT_EXTERNAL_CERT_NAME}-key.pem -out ${CERT_EXTERNAL_CERT_NAME}-cert.pem -days 365 >/dev/null 2>&1
}

cert_new() {
  $CMD cert new $CERT_NAME --tags $CERT_TAG
}

cert_new_standalone() {
  $CMD cert new $CERT_NAME --tags $CERT_TAG --standalone $CERT_EXPORT_FILE
}

cert_new_dn() {
  $CMD cert new $CERT_NAME --tags $CERT_TAG --dn-l lll --dn-st stst --dn-o ooo --dn-ou ouou --dn-c ccc --dn-street street --dn-postal postal
}

cert_new_ca() {
  $CMD cert new $CERT_NAME --tags $CERT_TAG --ca $CA_NAME
}

cert_new_ca_standalone() {
  $CMD cert new $CERT_NAME --tags $CERT_TAG --ca $CA_NAME --standalone $CERT_EXPORT_FILE
}

cert_list() {
  $CMD cert list
}

cert_check_exists() {
  $CMD cert list | grep -q "$1"
} 

cert_delete() {
  $CMD cert delete $CERT_NAME --confirm-delete "this is just a test"
}

cert_show() {
  $CMD cert show $CERT_NAME
}

cert_show_private() {
  $CMD cert show $CERT_NAME --private
}

cert_show_export() {
  $CMD cert show $CERT_NAME --export $CERT_EXPORT_FILE
}

cert_show_export_private() {
  $CMD cert show $CERT_NAME --export $CERT_EXPORT_FILE --private
}

cert_import_public() {
  $CMD cert new $CERT_NAME --cert ${CERT_EXTERNAL_CERT_NAME}-cert.pem --tags $CERT_TAG
}

cert_import_private() {
  $CMD cert new $CERT_NAME --cert ${CERT_EXTERNAL_CERT_NAME}-cert.pem --key ${CERT_EXTERNAL_CERT_NAME}-key.pem --tags $CERT_TAG
}
