# This is currently needed since Gin swagger does not support OpenAPI 3.* yet

# Convert the Swagger output (2.0) to OpenAPI 3.0 using [POST] https://converter.swagger.io/api/convert
# Payload is the entire content of the swagger.json file
# The output is saved to the same file

function convert_to_v3() {
  local base_filepath=$1
  # shellcheck disable=SC2155
  local payload=$(cat "$base_filepath.yaml")
  # shellcheck disable=SC2155
  local v3_res_yaml=$(curl -X POST 'https://converter.swagger.io/api/convert' \
    -H "accept: application/yaml" \
    -H "content-type: application/yaml" \
    --data-raw "$payload")
  # shellcheck disable=SC2155
  local v3_res_json=$(curl -X POST 'https://converter.swagger.io/api/convert' \
    -H "accept: application/json" \
    -H "content-type: application/yaml" \
    --data-raw "$payload")

  echo "$v3_res_yaml" >"$base_filepath.yaml"
  echo "$v3_res_json" >"$base_filepath.json"
}

convert_to_v3 "../docs/api/v1/V1_swagger"
convert_to_v3 '../docs/api/v2/V2_swagger'


echo "Converted to OpenAPI 3.0"