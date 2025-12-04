#!/usr/bin/env bats

# Ensure kubectl is installed
setup() {
  export BATS_LIB_PATH="${BATS_LIB_PATH}:/usr/lib"
  bats_load_library bats-support
  bats_load_library bats-assert
  bats_load_library bats-file
  bats_load_library bats-detik/detik.bash

  echo '{"clientId":"opentdf","clientSecret":"secret"}' >client_creds.json

  # Default otdfctl cmd
  local default_otdfctl_cmd="otdfctl --host https://platform.opentdf.local:9443 --with-client-creds-file ./client_creds.json"

  OTDFCTL_CMD="${OTDFCTL_CMD_OVERRIDE:-$default_otdfctl_cmd}"

  export OTDFCTL_CMD

}

@test "List namespaces" {
  # Run the command to list namespaces
  run $OTDFCTL_CMD policy attributes namespaces list --json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl policy attributes namespaces list' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi
}

@test "Create namespace and verify the output" {
  # Run the command to create a namespace
  run $OTDFCTL_CMD policy attributes namespaces create --name demo.com --json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl policy attributes namespaces create' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the output contains the namespace name
  assert_output --partial '"name": "demo.com"'

  # Extract the created namespace ID from the JSON output
  created_id=$(echo "$output" | jq -r '.id')

  # Assert that the created ID is not empty
  assert [ -n "$created_id" ]

  # Optionally, print the created ID for debugging
  echo "Created Namespace ID: $created_id"

  # Save the created namespace ID to a temporary file for use in other tests
  echo "$created_id" >/tmp/created_namespace_id.txt
}

@test "List namespaces and verify the new namespace exists" {
  # Read the created namespace ID from the temporary file
  if [ ! -f /tmp/created_namespace_id.txt ]; then
    echo "Created namespace ID file does not exist."
    exit 1
  fi
  created_id=$(cat /tmp/created_namespace_id.txt)

  # Run the command to list namespaces
  run $OTDFCTL_CMD policy attributes namespaces list --json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl policy attributes namespaces list' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the output contains the newly created namespace
  echo "$output" | jq -e '.namespaces[] | select(.id == "'$created_id'")' >/dev/null
  assert [ "$?" -eq 0 ]
}

@test "Get namespace by ID and verify the output" {
  # Read the created namespace ID from the temporary file
  if [ ! -f /tmp/created_namespace_id.txt ]; then
    echo "Created namespace ID file does not exist."
    exit 1
  fi
  namespace_id=$(cat /tmp/created_namespace_id.txt)

  # Run the command to get the namespace by ID
  run $OTDFCTL_CMD policy attributes namespaces get --id=$namespace_id --json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl policy attributes namespaces get' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the output contains the namespace details
  assert_output --partial '"id": "'$namespace_id'"'
  assert_output --partial '"name": "demo.com"'
  assert_output --partial '"fqn": "https://demo.com"'
  assert_output --partial '"value": true'
}

@test "Create attribute and verify the output" {
  # Read the created namespace ID from the temporary file
  if [ ! -f /tmp/created_namespace_id.txt ]; then
    echo "Created namespace ID file does not exist."
    exit 1
  fi
  namespace_id=$(cat /tmp/created_namespace_id.txt)

  # Run the command to create an attribute
  run $OTDFCTL_CMD policy attributes create --name role -s $namespace_id -r ANY_OF --json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl policy attributes create' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the output contains the attribute details
  assert_output --partial '"id": "'
  assert_output --partial '"namespace": {'
  assert_output --partial '"id": "'$namespace_id'"'
  assert_output --partial '"name": "role"'
  assert_output --partial '"fqn": "https://demo.com/attr/role"'
  assert_output --partial '"value": true'

  # Extract the created attribute ID from the JSON output
  attribute_id=$(echo "$output" | jq -r '.id')

  # Assert that the created ID is not empty
  assert [ -n "$attribute_id" ]

  # Optionally, print the created ID for debugging
  echo "Created Attribute ID: $attribute_id"

  # Save the created attribute ID to a temporary file for use in other tests
  echo "$attribute_id" >/tmp/created_attribute_id.txt
}

@test "Create admin value and verify the output" {
  # Read the created attribute ID from the temporary file
  if [ ! -f /tmp/created_attribute_id.txt ]; then
    echo "Created attribute ID file does not exist."
    exit 1
  fi
  attribute_id=$(cat /tmp/created_attribute_id.txt)

  # Run the command to create the admin value
  run $OTDFCTL_CMD policy attributes values create -a $attribute_id --value admin --json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl policy attributes values create' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the output contains the value details
  assert_output --partial '"id": "'
  assert_output --partial '"attribute": {'
  assert_output --partial '"id": "'$attribute_id'"'
  assert_output --partial '"value": "admin"'
  assert_output --partial '"fqn": "https://demo.com/attr/role/value/admin"'
  assert_output --partial '"value": true'

  # Extract the created value ID from the JSON output
  admin_value_id=$(echo "$output" | jq -r '.id')

  # Assert that the created ID is not empty
  assert [ -n "$admin_value_id" ]

  # Optionally, print the created ID for debugging
  echo "Created Admin Value ID: $admin_value_id"

  # Save the created admin value ID to a temporary file for use in other tests
  echo "$admin_value_id" >/tmp/admin_value_id.txt
}

@test "Create developer value and verify the output" {
  # Read the created attribute ID from the temporary file
  if [ ! -f /tmp/created_attribute_id.txt ]; then
    echo "Created attribute ID file does not exist."
    exit 1
  fi
  attribute_id=$(cat /tmp/created_attribute_id.txt)

  # Run the command to create the developer value
  run $OTDFCTL_CMD policy attributes values create -a $attribute_id --value developer --json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl policy attributes values create' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the output contains the value details
  assert_output --partial '"id": "'
  assert_output --partial '"attribute": {'
  assert_output --partial '"id": "'$attribute_id'"'
  assert_output --partial '"value": "developer"'
  assert_output --partial '"fqn": "https://demo.com/attr/role/value/developer"'
  assert_output --partial '"value": true'

  # Extract the created value ID from the JSON output
  developer_value_id=$(echo "$output" | jq -r '.id')

  # Assert that the created ID is not empty
  assert [ -n "$developer_value_id" ]

  # Optionally, print the created ID for debugging
  echo "Created Developer Value ID: $developer_value_id"

  # Save the created developer value ID to a temporary file for use in other tests
  echo "$developer_value_id" >/tmp/developer_value_id.txt
}

@test "Create guest value and verify the output" {
  # Read the created attribute ID from the temporary file
  if [ ! -f /tmp/created_attribute_id.txt ]; then
    echo "Created attribute ID file does not exist."
    exit 1
  fi
  attribute_id=$(cat /tmp/created_attribute_id.txt)

  # Run the command to create the guest value
  run $OTDFCTL_CMD policy attributes values create -a $attribute_id --value guest --json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl policy attributes values create' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the output contains the value details
  assert_output --partial '"id": "'
  assert_output --partial '"attribute": {'
  assert_output --partial '"id": "'$attribute_id'"'
  assert_output --partial '"value": "guest"'
  assert_output --partial '"fqn": "https://demo.com/attr/role/value/guest"'
  assert_output --partial '"value": true'

  # Extract the created value ID from the JSON output
  guest_value_id=$(echo "$output" | jq -r '.id')

  # Assert that the created ID is not empty
  assert [ -n "$guest_value_id" ]

  # Optionally, print the created ID for debugging
  echo "Created Guest Value ID: $guest_value_id"

  # Save the created guest value ID to a temporary file for use in other tests
  echo "$guest_value_id" >/tmp/guest_value_id.txt
}

@test "Get attribute and verify it contains the new values" {
  # Read the created attribute ID from the temporary file
  if [ ! -f /tmp/created_attribute_id.txt ]; then
    echo "Created attribute ID file does not exist."
    exit 1
  fi
  attribute_id=$(cat /tmp/created_attribute_id.txt)

  # Run the command to get the attribute by ID
  run $OTDFCTL_CMD policy attributes get --id=$attribute_id --json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl policy attributes get' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the output contains the attribute details
  assert_output --partial '"id": "'$attribute_id'"'
  assert_output --partial '"name": "role"'
  assert_output --partial '"fqn": "https://demo.com/attr/role"'
  assert_output --partial '"value": true'

  # Extract and check the values array
  values=$(echo "$output" | jq -r '.values[].value')
  assert [ "$(echo "$values" | grep -c 'admin')" -eq 1 ]
  assert [ "$(echo "$values" | grep -c 'developer')" -eq 1 ]
  assert [ "$(echo "$values" | grep -c 'guest')" -eq 1 ]
}

@test "Create subject condition set and verify the output" {
  # Run the command to create the subject condition set
  run $OTDFCTL_CMD policy subject-condition-sets create -s '[ { "condition_groups": [ { "conditions": [ { "subject_external_selector_value": ".clientId", "operator": 1, "subject_external_values": [ "opentdf" ] } ], "boolean_operator": 1 } ] } ]' --json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl policy subject-condition-sets create' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the output contains the subject condition set details
  assert_output --partial '"id": "'
  assert_output --partial '"subject_external_selector_value": ".clientId"'
  assert_output --partial '"operator": 1'
  assert_output --partial '"opentdf"'
  assert_output --partial '"boolean_operator": 1'

  # Extract the created subject condition set ID from the JSON output
  subject_condition_set_id=$(echo "$output" | jq -r '.id')

  # Assert that the created ID is not empty
  assert [ -n "$subject_condition_set_id" ]

  # Optionally, print the created ID for debugging
  echo "Created Subject Condition Set ID: $subject_condition_set_id"

  # Save the created subject condition set ID to a temporary file for use in other tests
  echo "$subject_condition_set_id" >/tmp/subject_condition_set_id.txt
}

@test "Create subject mapping and verify the output" {
  # Read the created developer value ID from the temporary file
  if [ ! -f /tmp/developer_value_id.txt ]; then
    echo "Developer value ID file does not exist."
    exit 1
  fi
  developer_value_id=$(cat /tmp/developer_value_id.txt)

  # Read the created subject condition set ID from the temporary file
  if [ ! -f /tmp/subject_condition_set_id.txt ]; then
    echo "Subject condition set ID file does not exist."
    exit 1
  fi
  subject_condition_set_id=$(cat /tmp/subject_condition_set_id.txt)

  # Run the command to create the subject mapping
  run $OTDFCTL_CMD policy subject-mappings create --action read --attribute-value-id $developer_value_id --subject-condition-set-id $subject_condition_set_id --json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl policy subject-mappings create' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the output contains the subject mapping details
  assert_output --partial '"id": "'
  assert_output --partial '"attribute_value": {'
  assert_output --partial '"id": "'$developer_value_id'"'
  assert_output --partial '"value": "developer"'
  assert_output --partial '"subject_condition_set": {'
  assert_output --partial '"id": "'$subject_condition_set_id'"'
  assert_output --partial '"subject_external_selector_value": ".clientId"'
  assert_output --partial '"operator": 1'
  assert_output --partial '"opentdf"'

  # Extract the created subject mapping ID from the JSON output
  subject_mapping_id=$(echo "$output" | jq -r '.id')

  # Assert that the created ID is not empty
  assert [ -n "$subject_mapping_id" ]

  # Optionally, print the created ID for debugging
  echo "Created Subject Mapping ID: $subject_mapping_id"

  # Save the created subject mapping ID to a temporary file for use in other tests
  echo "$subject_mapping_id" >/tmp/subject_mapping_id.txt
}

@test "Create KAS Key Mapping and verify the output" {
  # Fetch the base64 encoded PEM from the secret
  encoded_pem=$(kubectl get secret kas-private-keys -n $KUBE_NAMESPACE -o jsonpath='{.data.kas-cert\.pem}')

  # Check if the fetch was successful
  if [[ -z "$encoded_pem" ]]; then
    echo "Error: Could not retrieve kas-cert.pem from secret kas-private-keys." >&2
    exit 1
  fi

  # Check if jq command was successful
  if [[ $? -ne 0 ]]; then
    echo "Error: jq command failed to construct JSON." >&2
    exit 1
  fi

  run $OTDFCTL_CMD policy kas-registry create --uri "https://kas.opentdf.local:9443/kas" --json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl policy kas-registry create' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Extract created kas id
  kas_id=$(echo "$output" | jq -r '.id')

  # Create KAS Key Public_Key Only
  run $OTDFCTL_CMD policy kas-registry key create --kas "https://kas.opentdf.local:9443/kas" --key-id "r1" --algorithm "rsa:2048" --mode "public_key" --public-key-pem "$encoded_pem" --json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl policy kas-registry key create' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  key_id=$(echo "$output" | jq -r '.key.id')

  # Read the created developer value ID from the temporary file
  if [ ! -f /tmp/developer_value_id.txt ]; then
    echo "Developer value ID file does not exist."
    exit 1
  fi
  developer_value_id=$(cat /tmp/developer_value_id.txt)

  # Create key mapping to developer value
  run $OTDFCTL_CMD policy attribute value key assign --value "$developer_value_id" --key-id "$key_id"
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl policy attribute value key assign' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi
}

@test "Create TDF3 file and verify the output" {
  # Run the command to create a TDF3 file without attributes
  run bash -c 'echo "my first encrypted tdf" | $OTDFCTL_CMD encrypt -o opentdf-example.tdf --tdf-type tdf3'
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl encrypt tdf3' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the TDF3 file is created
  assert_file_exist opentdf-example.tdf
}

@test "Create nanoTDF file and verify the output" {
  # Run the command to create a nanoTDF file without attributes
  run bash -c 'echo "my first encrypted tdf" | $OTDFCTL_CMD encrypt -o opentdf-example.nano.tdf --tdf-type nano'
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl encrypt nano' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the nanoTDF file is created
  assert_file_exist opentdf-example.nano.tdf
}

@test "Decrypt TDF3 file and verify the output" {
  # Run the command to decrypt the TDF3 file
  run $OTDFCTL_CMD decrypt --tdf-type tdf3 opentdf-example.tdf
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl decrypt tdf3' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the decrypted output is as expected
  assert_output "my first encrypted tdf"
}

@test "Decrypt nanoTDF file and verify the output" {
  # Run the command to decrypt the nanoTDF file
  run $OTDFCTL_CMD decrypt --tdf-type nano opentdf-example.nano.tdf
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl decrypt nano' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the decrypted output is as expected
  assert_output "my first encrypted tdf"
}

@test "Encrypt TDF3 file with attributes and verify the output" {
  # Run the command to create a TDF3 file with attributes
  run bash -c 'echo "my first encrypted tdf" | $OTDFCTL_CMD encrypt -o opentdf-example.tdf --tdf-type tdf3 --attr https://demo.com/attr/role/value/guest'
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl encrypt tdf3 with attributes' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the TDF3 file is created
  assert_file_exist opentdf-example.tdf
}

@test "Encrypt nanoTDF file with attributes and verify the output" {
  # Run the command to create a nanoTDF file with attributes
  run bash -c 'echo "my first encrypted tdf" | $OTDFCTL_CMD encrypt -o opentdf-example.nano.tdf --tdf-type nano --attr https://demo.com/attr/role/value/guest'
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl encrypt nano with attributes' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the nanoTDF file is created
  assert_file_exist opentdf-example.nano.tdf
}

@test "Decrypt TDF3 file with attributes and expect failure" {
  # Run the command to decrypt the TDF3 file
  run $OTDFCTL_CMD decrypt --tdf-type tdf3 opentdf-example.tdf

  # Assert that the command failed
  assert_failure

  # Assert that the output contains the expected error message
  assert_output --partial 'ERROR    Failed to decrypt file:'
  assert_output --partial 'kao unwrap failed for split {https://platform.opentdf.local:9443/kas }: permission_denied: request error'
}

@test "Decrypt nanoTDF file with attributes and expect failure" {
  # Run the command to decrypt the nanoTDF file
  run $OTDFCTL_CMD decrypt --tdf-type nano opentdf-example.nano.tdf

  # Assert that the command failed
  assert_failure

  # Assert that the output contains the expected error message
  assert_output --partial 'ERROR    Failed to decrypt file:'
  assert_output --partial 'rpc error: code = PermissionDenied desc = forbidden'
}

@test "Create subject mapping for guest access and verify the output" {
  # Read the created guest value ID from the temporary file
  if [ ! -f /tmp/guest_value_id.txt ]; then
    echo "Guest value ID file does not exist."
    exit 1
  fi
  guest_value_id=$(cat /tmp/guest_value_id.txt)

  # Read the created subject condition set ID from the temporary file
  if [ ! -f /tmp/subject_condition_set_id.txt ]; then
    echo "Subject condition set ID file does not exist."
    exit 1
  fi
  subject_condition_set_id=$(cat /tmp/subject_condition_set_id.txt)

  # Run the command to create the subject mapping
  run $OTDFCTL_CMD policy subject-mappings create --action read --attribute-value-id $guest_value_id --subject-condition-set-id $subject_condition_set_id --json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl policy subject-mappings create' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the output contains the subject mapping details
  assert_output --partial '"id": "'
  assert_output --partial '"attribute_value": {'
  assert_output --partial '"id": "'$guest_value_id'"'
  assert_output --partial '"value": "guest"'
  assert_output --partial '"subject_condition_set": {'
  assert_output --partial '"id": "'$subject_condition_set_id'"'
  assert_output --partial '"subject_external_selector_value": ".clientId"'
  assert_output --partial '"operator": 1'
  assert_output --partial '"opentdf"'

  # Extract the created subject mapping ID from the JSON output
  subject_mapping_id=$(echo "$output" | jq -r '.id')

  # Assert that the created ID is not empty
  assert [ -n "$subject_mapping_id" ]

  # Optionally, print the created ID for debugging
  echo "Created Subject Mapping ID: $subject_mapping_id"

  # Save the created subject mapping ID to a temporary file for use in other tests
  echo "$subject_mapping_id" >/tmp/guest_subject_mapping_id.txt
}

@test "Decrypt TDF3 file with new subject mapping and verify the output" {
  # Run the command to decrypt the TDF3 file
  run $OTDFCTL_CMD decrypt --tdf-type tdf3 opentdf-example.tdf
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl decrypt tdf3' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the decrypted output is as expected
  assert_output "my first encrypted tdf"
}

@test "Decrypt nanoTDF file with new subject mapping and verify the output" {
  # Run the command to decrypt the nanoTDF file
  run $OTDFCTL_CMD decrypt --tdf-type nano opentdf-example.nano.tdf
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl decrypt nano' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the decrypted output is as expected
  assert_output "my first encrypted tdf"
}

@test "Create and Decrypt with External KAS" {
  # Check we can reach kas.opentdf.local
  run curl -f -sS https://kas.opentdf.local:9443/kas/v2/kas_public_key
  if [ "$status" -ne 0 ]; then
    echo "Error: 'curl kas public key' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  run jq --raw-output '.kid' <<<"$output"
  if [ "$status" -ne 0 ]; then
    echo "Error: 'jq .kid' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi
  assert_output "r1"

  # Run the command to create a TDF3 file with attributes
  run bash -c 'echo "my first encrypted tdf" | $OTDFCTL_CMD encrypt -o opentdf-key-mapping-example.tdf --tdf-type tdf3 --attr https://demo.com/attr/role/value/developer'
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl encrypt for external KAS' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  # Assert that the TDF3 file is created
  assert_file_exist opentdf-key-mapping-example.tdf

  run unzip -o opentdf-key-mapping-example.tdf
  if [ "$status" -ne 0 ]; then
    echo "Error: 'unzip tdf' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi

  run jq '.encryptionInformation.keyAccess | length' 0.manifest.json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'jq keyAccess length' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi
  assert_output "1"

  run jq --raw-output '.encryptionInformation.keyAccess[0].url' 0.manifest.json
  if [ "$status" -ne 0 ]; then
    echo "Error: 'jq keyAccess url' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi
  assert_output "https://kas.opentdf.local:9443/kas" "Expected KAS URL to be https://kas.opentdf.local:9443/kas, but got $output"

  # Decrypt TDF with external kas
  run $OTDFCTL_CMD decrypt --tdf-type tdf3 opentdf-key-mapping-example.tdf
  if [ "$status" -ne 0 ]; then
    echo "Error: 'otdfctl decrypt external KAS TDF' failed with status $status." >&2
    echo "Output: $output" >&2
    return 1
  fi
  assert_output "my first encrypted tdf"
}
