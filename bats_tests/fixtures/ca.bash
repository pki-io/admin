export CA_NAME="testca"
export CA_TAG="testtag"
export CA_DN_ARG="--dn-l lll --dn-st stst --dn-o ooo --dn-ou ouou --dn-c ccc --dn-street street --dn-postal postal"
export CA_EXTERNAL_CA_NAME="external-ca"

create_external_ca() {
  openssl genrsa -out ${CA_EXTERNAL_CA_NAME}-key.pem 2048 >/dev/null 2>&1
  openssl req -new -batch -x509 -extensions v3_ca -key ${CA_EXTERNAL_CA_NAME}-key.pem -out ${CA_EXTERNAL_CA_NAME}-cert.pem -days 3650 >/dev/null 2>&1
}

ca_new() {
  $CMD ca new $CA_NAME --tags $CA_TAG
}

ca_new_dnscope() {
  $CMD ca new $CA_NAME --tags $CA_TAG $CA_DN_ARG
}

ca_list() {
  $CMD ca list
}

ca_check_exists() {
  $CMD ca list | grep -q "$1"
}

ca_delete() {
  $CMD ca delete $CA_NAME --confirm-delete "this is just a test"
}

ca_show() {
  $CMD ca show $CA_NAME
}

ca_show_private() {
  $CMD ca show $CA_NAME --private
}

ca_import_public() {
  $CMD ca import $CA_NAME ${CA_EXTERNAL_CA_NAME}-cert.pem --tags $CA_TAG $CA_DN_ARG
}

ca_import_private() {
  $CMD ca import $CA_NAME ${CA_EXTERNAL_CA_NAME}-cert.pem ${CA_EXTERNAL_CA_NAME}-key.pem --tags $CA_TAG $CA_DN_ARG
}
