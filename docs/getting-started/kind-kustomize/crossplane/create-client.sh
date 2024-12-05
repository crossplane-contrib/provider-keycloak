# uses curl to invoke the keycloak REST api for crossplane
# gets a token for the master realm
echo "Logging on as admin in keycloak to create the crossplane client and grant it the admin role in the master realm"
mastertoken=$(curl -k -g -d "client_id=admin-cli" -d "username=admin" -d "password=admin" -d "grant_type=password" -d "client_secret=" "http://keycloak.keycloak:80/realms/master/protocol/openid-connect/token" | sed 's/.*access_token":"//g' | sed 's/".*//g');
# echo $mastertoken;

id="9d2308c3-8972-40cf-9cca-1256745c16d4";
url="http://keycloak.keycloak:80/admin/realms/master";
clienturl="$url/clients/$id";

# creates a new client named "crossplane"
curl -X POST -k -g "$url/clients" \
-H "Authorization: Bearer $mastertoken" \
-H "Content-Type: application/json" \
--data-raw '
{
  "id":"'$id'",
  "name":"crossplane",
  "clientId":"crossplane",
  "secret":"xppw_OJKzQjuBoyPlIEePgiWg",
  "clientAuthenticatorType":"client-secret",
  "serviceAccountsEnabled":"true",
  "standardFlowEnabled":"false"
}'

# GETs the service-account-user for the client - GET $url/clients/{id}/service-account-user
userid=$(curl -X GET -k -g "$clienturl/service-account-user" -H "Authorization: Bearer $mastertoken"  | sed 's/.*id":"//g' | sed 's/".*//g')

# lists available realm roles
# GET /{realm}/users/{id}/role-mappings/realm/available
roles=$(curl -X GET -k -g -H "Authorization: Bearer $mastertoken" "$url/roles")

# gets the id of the admin role
admin_id=$(echo $roles | jq -r '.[] | select(.name == "admin") | .id')

# adds service account role admin to the client's user
curl -X POST -k -g "$url/users/$userid/role-mappings/realm/" \
-H "Authorization: Bearer $mastertoken" \
-H "Content-Type: application/json" \
--data-raw '[
{
    "id":"'$admin_id'",
    "name":"admin"
}
]'