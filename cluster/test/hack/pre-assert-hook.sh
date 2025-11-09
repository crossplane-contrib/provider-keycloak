#!/bin/bash
echo "Hooked"
kubectl get clientdefaultscopes.samlclient.keycloak.crossplane.io/saml-client-default-scopes -o yaml