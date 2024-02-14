
# Allow issue certificate with role with default issuer from Intermediate PKI
path "pki_int/issue/+" {
   capabilities = ["create",  "update"]
}

## Revole certificate from Intermediate PKI
path "pki_int/revoke" {
  capabilities = ["create",  "update"]
}

## List Revoked Certificates from Intermediate PKI
path "pki_int/certs/revoked" {
  capabilities = ["list"]
}


## List Certificates from Intermediate PKI
path "pki_int/certs" {
  capabilities = ["list"]
}

## Read Certificate from Intermediate PKI
path "pki_int/cert/+" {
  capabilities = ["read"]
}
path "pki_int/cert/+/raw" {
  capabilities = ["read"]
}
path "pki_int/cert/+/raw/pem" {
  capabilities = ["read"]
}
