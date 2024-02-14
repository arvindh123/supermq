#!/usr/bin/bash
# Copyright (c) Abstract Machines
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

scriptdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
echo "$scriptdir"
export MAGISTRALA_DIR=$scriptdir/../../

# cd $scriptdir

# echo "$MAGISTRALA_DIR"

readDotEnv() {
    set -o allexport
    source $MAGISTRALA_DIR/docker/.env
    set +o allexport
}

vaultCreatePolicy() {
    vault policy write -namespace=${MG_VAULT_NAMESPACE} -address=${MG_VAULT_ADDR} things_cert_issue $scriptdir/things_cert_issue.hcl
}

vaultCreateRole() {
    echo "Creating new AppRole"
    # vault auth enable approle
    vault write -namespace=${MG_VAULT_NAMESPACE} -address=${MG_VAULT_ADDR} auth/approle/role/things_cert_issuer \
    token_policies=things_cert_issue  secret_id_num_uses=0 \
    secret_id_ttl=0 token_ttl=1h token_max_ttl=3h  token_num_uses=0
}

vaultWriteCustomRoleID(){
    echo "Writing custom role id"
    vault read -namespace=${MG_VAULT_NAMESPACE} -address=${MG_VAULT_ADDR} auth/approle/role/things_cert_issuer/role-id
    vault write -namespace=${MG_VAULT_NAMESPACE} -address=${MG_VAULT_ADDR} auth/approle/role/things_cert_issuer/role-id role_id=${MG_CERTS_VAULT_APPROLE_ROLEID}
}

vaultWriteCustomSecret() {
    echo "Writing custom secret"
    vault write -namespace=${MG_VAULT_NAMESPACE} -address=${MG_VAULT_ADDR} -f auth/approle/role/things_cert_issuer/secret-id
    vault write -namespace=${MG_VAULT_NAMESPACE} -address=${MG_VAULT_ADDR} auth/approle/role/things_cert_issuer/custom-secret-id secret_id=${MG_CERTS_VAULT_APPROLE_SECRET} num_uses=0 ttl=0
}

vaultTestRoleLogin() {
echo "Testing custom roleid secret by logging in"
vault write -namespace=${MG_VAULT_NAMESPACE} -address=${MG_VAULT_ADDR} auth/approle/login \
    role_id=${MG_CERTS_VAULT_APPROLE_ROLEID} \
    secret_id=${MG_CERTS_VAULT_APPROLE_SECRET}

}
if ! command -v jq &> /dev/null
then
    echo "jq command could not be found, please install it and try again."
    exit
fi

readDotEnv


vault login  -namespace=${MG_VAULT_NAMESPACE} -address=${MG_VAULT_ADDR} ${MG_VAULT_TOKEN}

vaultCreatePolicy
vaultCreateRole
vaultCreateRole
vaultWriteCustomRoleID
vaultWriteCustomSecret
vaultTestRoleLogin

exit 0
