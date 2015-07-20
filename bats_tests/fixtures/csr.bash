export CSR_NAME="testcsr"
export CSR_EXPORT_FILE="${CSR_NAME}.tar.gz"
export CSR_TAG="testtag"
export CSR_EXTERNAL_CSR_NAME="external-csr"

create_external_csr() {
  openssl genrsa -out ${CSR_EXTERNAL_CSR_NAME}-key.pem 2048 >/dev/null 2>&1
  openssl req -new -batch -key ${CSR_EXTERNAL_CSR_NAME}-key.pem -out ${CSR_EXTERNAL_CSR_NAME}-csr.pem -days 365 >/dev/null 2>&1
}

csr_new() {
  $CMD csr new $CSR_NAME --tags $CSR_TAG
}

csr_new_standalone() {
  $CMD csr new $CSR_NAME --tags $CSR_TAG --standalone $CSR_EXPORT_FILE
}

csr_new_dn() {
  $CMD csr new $CSR_NAME --tags $CSR_TAG --dn-l lll --dn-st stst --dn-o ooo --dn-ou ouou --dn-c ccc --dn-street street --dn-postal postal
}

csr_new_ca_standalone() {
  $CMD csr new $CSR_NAME --tags $CSR_TAG --ca $CA_NAME --standalone $CSR_EXPORT_FILE
}

csr_list() {
  $CMD csr list
}

csr_check_exists() {
  $CMD csr list | grep -q "$1"
} 

csr_delete() {
  $CMD csr delete $CSR_NAME --confirm-delete "this is just a test"
}

csr_show() {
  $CMD csr show $CSR_NAME
}

csr_show_private() {
  $CMD csr show $CSR_NAME --private
}

csr_show_export() {
  $CMD csr show $CSR_NAME --export $CSR_EXPORT_FILE
}

csr_show_export_private() {
  $CMD csr show $CSR_NAME --export $CSR_EXPORT_FILE --private
}

csr_sign() {
  $CMD csr sign $CSR_NAME ${CSR_EXTERNAL_CSR_NAME}-csr.pem --ca $CA_NAME --tags $CSR_TAG
}

csr_sign_standalone() {
  q
  $CMD csr sign $CSR_NAME ${CSR_EXTERNAL_CSR_NAME}-csr.pem --ca $CA_NAME --tags $CSR_TAG --standalone $CSR_EXPORT_FILE
}
