package dashboard_service

import (
	"aidashboard/internal/service/netzilo_service"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/xeipuuv/gojsonschema"
)

func GetSpecification(specName string) map[string]interface{} {
	path := "./internal/model/specification/"
	filename := fmt.Sprintf("%s%s.json", path, specName)

	// Read the JSON file
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Unmarshal the JSON data into a map
	var spec map[string]interface{}
	err = json.Unmarshal(data, &spec)
	if err != nil {
		log.Fatalf("Error %v while unmarshaling JSON: %s", err, string(data))
	}

	return spec
}

func GetPrompt(userCommand string) []byte {
	postureCheck := GetSpecification("checks")
	policy := GetSpecification("policies")
	profile := GetSpecification("profiles")

	systemMessage := map[string]interface{}{
		"role":    "system",
		"content": "You are a helpful assistant about computer network. Don't skip or ignore any elements which are between valid elements in the schema of functions.",
	}
	userMessage := map[string]interface{}{
		"role":    "user",
		"content": userCommand,
	}

	// Create a person variable as a map
	prompt := map[string]interface{}{
		"model":    "gpt-4o",
		"messages": []interface{}{systemMessage, userMessage},
		"tools": []interface{}{
			policy,
			profile,
			postureCheck,
		},
		"tool_choice": "auto",
	}

	// Print the person with the policy
	promptJSON, err := json.MarshalIndent(prompt, "", "  ")
	// fmt.Println(string(promptJSON))
	if err != nil {
		log.Fatalf("Error marshaling person to JSON: %v", err)
	}

	return promptJSON
}

func GenerateDashboard(respMessage []byte) (map[string]interface{}, error) {
	// Unmarshal the JSON data into a map
	var message map[string]interface{}
	err := json.Unmarshal(respMessage, &message)
	if err != nil {
		log.Fatalf("Error %v while unmarshaling JSON: %s", err, string(respMessage))
		return nil, err
	}

	response := make(map[string]interface{})
	// Extract the choices from the response
	if toolCalls, ok := message["tool_calls"].([]interface{}); ok && len(toolCalls) > 0 {
		postureCheck, policy, profile, err := ExtractFromGptResponse(toolCalls)
		if err != nil {
			return nil, err
		}

		var postureCheckId string = ""
		if postureCheck != nil {
			postureCheckResp, err := netzilo_service.CreatePostureCheck(postureCheck)
			if err != nil {
				return nil, err
			}
			response["posture-check"] = postureCheckResp
			if policy != nil || profile != nil {
				postureCheckId, _ = postureCheckResp["id"].(string)
			}
		}

		if policy != nil {
			// Create groups or convert group names into id
			policy = AddGroupId2Policy(policy)
			// Add posture_check_id
			policy = AddCheckId2Policy(policy, postureCheckId)
			policyResp, err := netzilo_service.CreatePolicy(policy)
			if err != nil {
				return nil, err
			}
			response["policy"] = policyResp
		}

		if profile != nil {
			// Create groups or convert group names into id
			profile = AddGroupId2Profile(profile)
			// Add posture_check_id
			profile = AddCheckId2Profile(profile, postureCheckId)
			profileResp, err := netzilo_service.CreateProfile(profile)
			if err != nil {
				return nil, err
			}
			response["profile"] = profileResp
		}
	}

	return response, nil
}

func ExtractFromGptResponse(toolCalls []interface{}) (map[string]interface{}, map[string]interface{}, map[string]interface{}, error) {
	var postureCheck, policy, profile map[string]interface{}
	for i := 0; i < len(toolCalls); i++ {
		if toolCall, ok := toolCalls[i].(map[string]interface{}); ok {
			if function, ok := toolCall["function"].(map[string]interface{}); ok {
				if name, ok := function["name"].(string); ok {
					if arguments, ok := function["arguments"].(string); ok {
						// Unmarshal the JSON data into a map
						var params map[string]interface{}
						err := json.Unmarshal([]byte(arguments), &params)
						if err != nil {
							log.Fatalf("Error unmarshaling JSON: %v", err)
							return nil, nil, nil, err
						}
						switch name {
						case "create_posture_check":
							postureCheck = GeneratePostureCheck(params)
							fmt.Println("create_posture_check")
						case "create_policy":
							policy = GeneratePolicy(params)
							fmt.Println("create_policy")
						case "create_profile":
							profile = GenerateProfile(params)
							fmt.Println("create_profile")
						default:
							return nil, nil, nil, err
						}
					}
				}
			}
		}
	}
	return postureCheck, policy, profile, nil
}

func GeneratePostureCheck(params map[string]interface{}) map[string]interface{} {
	// Unmarshal the JSON data into a map
	var check map[string]interface{} = params
	return check
}

func GeneratePolicy(params map[string]interface{}) map[string]interface{} {
	// Unmarshal the JSON data into a map
	var policy map[string]interface{} = params
	// Process protocol
	if rules, ok := policy["rules"].([]interface{}); ok && len(rules) > 0 {
		for i := 0; i < len(rules); i++ {
			if rule, ok := rules[i].(map[string]interface{}); ok {
				var protocol string
				var ok bool
				protocol, ok = rule["protocol"].(string)
				if !ok {
					protocol = "all"
					rule["protocol"] = protocol
				}
				// Current setting rules for bi-directional are wrong, maybe
				if protocol == "all" || protocol == "icmp" {
					rule["bidirectional"] = true
				} else {
					if _, ok = rule["bidirectional"].(string); !ok {
						rule["bidirectional"] = true
					}
				}
			}
		}
	}
	return policy
}

func GenerateProfile(params map[string]interface{}) map[string]interface{} {
	var profile map[string]interface{} = params
	// Set default value of os as "windows"
	if _, ok := profile["os"].(string); !ok {
		profile["os"] = "Windows"
	}
	return profile
}

func AddGroupId2Policy(policyWithoutGroupId map[string]interface{}) map[string]interface{} {
	policy := policyWithoutGroupId
	if rules, ok := policy["rules"].([]interface{}); ok && len(rules) > 0 {
		for i := 0; i < len(rules); i++ {
			if rule, ok := rules[i].(map[string]interface{}); ok {
				dest_src_groups := []string{"destinations", "sources"}
				for _, dest_src := range dest_src_groups {
					var groupNames []string
					seen := make(map[string]struct{}) // Use a map to track seen group names
					if groups, ok := rule[dest_src].([]interface{}); ok && len(groups) > 0 {
						for i := 0; i < len(groups); i++ {
							if group, ok := groups[i].(string); ok && group != "" {
								if _, exists := seen[group]; !exists { // Check if the group has already been added
									seen[group] = struct{}{}               // Mark the group as seen
									groupNames = append(groupNames, group) // Append the group name
								}
							}
						}
					}
					rule[dest_src] = ConvertGroupNames2IDs(groupNames)
				}
			}
		}
	}
	return policy
}

func AddGroupId2Profile(profileWithoutGroupId map[string]interface{}) map[string]interface{} {
	profile := profileWithoutGroupId
	var groupNames []string
	if groups, ok := profile["groups"].([]interface{}); ok && len(groups) > 0 {
		for i := 0; i < len(groups); i++ {
			if group, ok := groups[i].(string); ok && group != "" {
				groupNames = append(groupNames, group)
			}
		}
	}
	profile["groups"] = ConvertGroupNames2IDs(groupNames)
	return profile
}

func AddCheckId2Policy(policyWithoutPostureCheckId map[string]interface{}, postureCheckId string) map[string]interface{} {
	policy := policyWithoutPostureCheckId
	policy["source_posture_checks"] = []interface{}{postureCheckId}
	return policy
}

func AddCheckId2Profile(profileWithoutPostureCheckId map[string]interface{}, postureCheckId string) map[string]interface{} {
	profile := profileWithoutPostureCheckId
	if components, ok := profile["components"].(map[string]interface{}); ok {
		if netzilo_workspace, ok := components["netzilo_workspace"].(map[string]interface{}); ok {
			netzilo_workspace["checks"] = []interface{}{postureCheckId}
		}
		if browser_extension, ok := components["browser_extension"].([]interface{}); ok && len(browser_extension) > 0 {
			for i := 0; i < len(browser_extension); i++ {
				if component, ok := browser_extension[i].(map[string]interface{}); ok {
					component["checks"] = []interface{}{postureCheckId}
				}
			}
		}
	}
	return profile
}

func ConvertGroupNames2IDs(groupNames []string) []interface{} {
	// Unmarshal the JSON data into a map
	var groupIDs []string

	groups, err := netzilo_service.GetGroups()
	if err != nil {
		return nil
	}

	for _, groupName := range groupNames {
		// Convert each element to string
		groupIDs = append(groupIDs, fmt.Sprintf("%v", ConvertOneGroupName2ID(groupName, groups)))
	}
	if len(groupIDs) == 0 {
		groupIDs = append(groupIDs, ConvertOneGroupName2ID("All", groups))
	}

	// Convert []string to []interface{}
	interfaceSlice := make([]interface{}, len(groupIDs))
	for i, v := range groupIDs {
		interfaceSlice[i] = v
	}
	return interfaceSlice
}

func ConvertOneGroupName2ID(groupName string, groups []interface{}) string {
	var groupID string = ""
	for i := 0; i < len(groups); i++ {
		if group, ok := groups[i].(map[string]interface{}); ok {
			if id, ok := group["id"].(string); ok {
				groupID = id
				break
			}
		}
	}
	return groupID
}

func JsonValidator(jsonData string, jsonSchema string) bool {
	// Define JSON schema and JSON data
	schemaLoader := gojsonschema.NewStringLoader(jsonSchema)
	documentLoader := gojsonschema.NewStringLoader(jsonData)

	// Validate the JSON document against the schema
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		fmt.Printf("Error occurred: %s\n", err)
		return false
	}

	if result.Valid() {
		fmt.Println("The document is valid")
		return true
	} else {
		fmt.Println("The document is not valid. See errors:")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
		return false
	}
}
